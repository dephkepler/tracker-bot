package handlers

import (
	"errors"
	"tracker-bot/internal/buttons/roadmap"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/internal/utils/tgctx"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

// The AI half of the Roadmap handlers. Every method here makes a network call
// that can take the better part of a minute, so the dispatcher runs them in a
// goroutine — Dispatcher.Run processes updates serially, and a synchronous
// call would freeze the bot for every user for the duration. Consequently
// nothing in this file may touch userSession: state that has to survive the
// call lives in the dispatcher's own synchronized store.

// aiErrorText maps a service error onto what the user should read. A disabled
// provider is reachable only from a stale keyboard, since the buttons are not
// drawn with AI off.
func aiErrorText(lang i18n.Lang, err error) string {
	switch {
	case errors.Is(err, models.ErrRoadmapAIDisabled):
		return i18n.T(lang, i18n.KeyRoadmapAIDisabled)
	case errors.Is(err, models.ErrRoadmapAIEmptyResult), errors.Is(err, models.ErrRoadmapNoCardsParsed):
		return i18n.T(lang, i18n.KeyRoadmapAIEmpty)
	default:
		return i18n.T(lang, i18n.KeyRoadmapAIFailed)
	}
}

// HandleRoadmapAIPlan generates a checklist for one technology. The "working"
// notice replaces the technology screen the button was tapped on, and the
// screen is re-rendered over it once the cards land, so the flow costs one
// message rather than three.
func (m *Module) HandleRoadmapAIPlan(ctx *tgctx.MsgContext, roadmapID int64) {
	m.sendOrEditRoadmap(ctx, true, i18n.T(ctx.Language, i18n.KeyRoadmapAIWorking), nil)

	added, rejected, err := m.roadmapaisvc.GeneratePlan(ctx.Ctx, ctx.DBUserID, roadmapID, string(ctx.Language))
	if err != nil {
		log.Error().Err(err).Int64("user_id", ctx.DBUserID).Int64("roadmap_id", roadmapID).Msg("generate roadmap plan failed")
		m.sendOrEditRoadmap(ctx, true, aiErrorText(ctx.Language, err), nil)
		// Leave the user on the technology rather than on the error: the
		// checklist is what they need to act on either way.
		m.ShowRoadmapDetail(ctx, roadmapID, false)
		return
	}

	m.ShowRoadmapDetail(ctx, roadmapID, true)
	m.noticeCardsAdded(ctx, added, rejected)
}

// PromptRoadmapAIPaste opens the AI-tagged paste flow. Same shape as
// PromptAddRoadmapCards, minus the tag documentation — the point of this path
// is not having to tag anything.
func (m *Module) PromptRoadmapAIPaste(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAIPastePrompt))
	msg.ReplyMarkup = roadmap.RoadmapAddCardsReplyMenu(ctx.Language)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send roadmap ai paste prompt failed")
	}
}

// ProcessRoadmapAIPaste tags one pasted message and stores the cards. Like
// ProcessAddRoadmapCards, the caller keeps the flow open — the user may paste
// several messages and leaves via "Done".
func (m *Module) ProcessRoadmapAIPaste(ctx *tgctx.MsgContext, roadmapID int64) {
	added, rejected, err := m.roadmapaisvc.AddCardsFromTextAI(ctx.Ctx, ctx.DBUserID, roadmapID, ctx.Text, string(ctx.Language))
	if err != nil {
		log.Error().Err(err).Int64("user_id", ctx.DBUserID).Int64("roadmap_id", roadmapID).Msg("ai tag roadmap cards failed")
		m.sendRoadmapPlain(ctx.ChatID, aiErrorText(ctx.Language, err), nil)
		return
	}
	m.noticeCardsAdded(ctx, added, rejected)
}

// noticeCardsAdded reports the counts. Deliberately the same confirmation as
// the manual paste path — from the user's side the outcome is identical.
func (m *Module) noticeCardsAdded(ctx *tgctx.MsgContext, added, rejected int) {
	text := i18n.T(ctx.Language, i18n.KeyRoadmapAddCardsAdded, added)
	if rejected > 0 {
		text += i18n.T(ctx.Language, i18n.KeyRoadmapAIRejectedFmt, rejected)
	}
	m.sendRoadmapPlain(ctx.ChatID, text, nil)
}

// AskRoadmapQuiz generates a question about a card and sends it with a Cancel
// keyboard. remember is called with the quiz before the question goes out, so
// the answer cannot arrive before the dispatcher knows what was asked.
func (m *Module) AskRoadmapQuiz(ctx *tgctx.MsgContext, cardID int64, remember func(models.RoadmapQuiz)) {
	notice := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAIQuizWorking))
	working, err := m.bot.Send(notice)
	if err != nil {
		log.Error().Err(err).Msg("send roadmap quiz notice failed")
	}

	quiz, err := m.roadmapaisvc.QuizCard(ctx.Ctx, ctx.DBUserID, cardID, string(ctx.Language))
	if err != nil {
		log.Error().Err(err).Int64("user_id", ctx.DBUserID).Int64("card_id", cardID).Msg("generate roadmap quiz failed")
		m.replaceOrSend(ctx.ChatID, working.MessageID, aiErrorText(ctx.Language, err))
		return
	}

	remember(quiz)

	// The notice has served its purpose; the question needs its own message
	// because a Cancel reply-keyboard cannot be attached to an edit.
	if working.MessageID > 0 {
		if _, err := m.bot.Request(tgbotapi.NewDeleteMessage(ctx.ChatID, working.MessageID)); err != nil {
			log.Debug().Err(err).Msg("delete roadmap quiz notice failed")
		}
	}

	// The question carries the user's own card text, so it goes out through
	// the plain-text path for the same reason the screens do — an unbalanced
	// _ or * would otherwise cost them the question entirely.
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAIQuizPromptFmt, quiz.Question))
	msg.ReplyMarkup = roadmap.RoadmapWaitingReplyMenu(ctx.Language)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send roadmap quiz question failed")
	}
}

// GradeRoadmapQuiz judges an answer and offers to tick the card.
func (m *Module) GradeRoadmapQuiz(ctx *tgctx.MsgContext, quiz models.RoadmapQuiz) {
	grade, err := m.roadmapaisvc.GradeQuizAnswer(ctx.Ctx, quiz, ctx.Text, string(ctx.Language))
	if err != nil {
		log.Error().Err(err).Int64("user_id", ctx.DBUserID).Int64("card_id", quiz.CardID).Msg("grade roadmap quiz failed")
		m.sendRoadmapPlain(ctx.ChatID, aiErrorText(ctx.Language, err), nil)
		return
	}

	header := i18n.T(ctx.Language, i18n.KeyRoadmapAIQuizPartial)
	switch grade.Verdict {
	case models.RoadmapQuizCorrect:
		header = i18n.T(ctx.Language, i18n.KeyRoadmapAIQuizCorrect)
	case models.RoadmapQuizWrong:
		header = i18n.T(ctx.Language, i18n.KeyRoadmapAIQuizWrong)
	}

	text := i18n.T(ctx.Language, i18n.KeyRoadmapAIQuizGradeFmt, header, grade.Feedback)
	menu := roadmap.RoadmapQuizResultInlineMenu(ctx.Language, quiz.CardID)
	// Offered whatever the verdict: whether the answer counts as knowing the
	// card is the user's call, not the model's.
	m.sendRoadmapPlain(ctx.ChatID, text, &menu)
}

// replaceOrSend edits a previously sent notice, falling back to a fresh
// message when there is nothing to edit.
func (m *Module) replaceOrSend(chatID int64, messageID int, text string) {
	if messageID > 0 {
		if _, err := m.bot.Send(tgbotapi.NewEditMessageText(chatID, messageID, text)); err == nil {
			return
		}
	}
	m.sendRoadmapPlain(chatID, text, nil)
}
