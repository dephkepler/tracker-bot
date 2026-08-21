package roadmap

import (
	"fmt"
	"strconv"
	"strings"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// digestButtonMaxRunes caps the card label in a reminder push, where each card
// gets a row to itself and so has the whole width. That is still only forty
// characters or so, which is why the checklist screen numbers its buttons
// instead and writes the cards out in the message text.
const digestButtonMaxRunes = 40

// Difficulty and kind are shown as icons rather than words: a card row also
// carries the card's own text, and three words of metadata in front of it
// would leave no room for the thing the user actually wrote.
func difficultyIcon(difficulty int) string {
	switch difficulty {
	case models.RoadmapCardEasy:
		return "🟢"
	case models.RoadmapCardHard:
		return "🔴"
	default:
		return "🟡"
	}
}

func kindIcon(kind models.RoadmapCardKind) string {
	switch kind {
	case models.RoadmapCardArticle:
		return "📄"
	case models.RoadmapCardBook:
		return "📚"
	case models.RoadmapCardLecture:
		return "🎧"
	default:
		return "📌"
	}
}

// RoadmapEntryInlineMenu renders the Roadmap root. hasOrphans adds the
// "without a goal" row, which only exists when something actually needs
// re-attaching — otherwise it is a dead end.
func RoadmapEntryInlineMenu(lang i18n.Lang, remindersActive, hasOrphans bool) tgbotapi.InlineKeyboardMarkup {
	reminderLabel := i18n.T(lang, i18n.KeyRoadmapButtonStartReminders)
	if remindersActive {
		reminderLabel = i18n.T(lang, i18n.KeyRoadmapButtonManageReminders)
	}
	rows := [][]tgbotapi.InlineKeyboardButton{
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonGoals), RoadmapCBGoalsOpen),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonCreateGoal), RoadmapCBGoalCreate),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(reminderLabel, RoadmapCBPushOpen),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonArchive), RoadmapCBArchiveOpen),
		),
	}
	if remindersActive {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonStopReminders), RoadmapCBPushStop),
		))
	}
	if hasOrphans {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonOrphans), RoadmapCBOrphansOpen),
		))
	}
	rows = append(rows,
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonProgress), RoadmapCBStats)),
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonHome), "go_home")),
	)
	return buttonbuilder.IK(rows...)
}

func RoadmapBackToMainInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)),
	)
}

// RoadmapGoalsInlineMenu lists the goals with their overall card progress.
func RoadmapGoalsInlineMenu(lang i18n.Lang, goals []models.RoadmapGoalItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(goals)+2)
	for _, g := range goals {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(GoalItemLabel(lang, g), fmt.Sprintf("%s%d", RoadmapCBGoalOpen, g.ID)),
		))
	}
	if len(goals) < models.MaxRoadmapGoalsPerUser {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonCreateGoal), RoadmapCBGoalCreate),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)))
	return buttonbuilder.IK(rows...)
}

func GoalItemLabel(lang i18n.Lang, g models.RoadmapGoalItem) string {
	return i18n.T(lang, i18n.KeyRoadmapGoalItemFmt, g.Name, g.DoneCards, g.TotalCards, PercentDone(g.DoneCards, g.TotalCards))
}

// RoadmapGoalDetailInlineMenu lists one goal's technologies plus the
// goal-level actions.
func RoadmapGoalDetailInlineMenu(lang i18n.Lang, goalID int64, items []models.RoadmapItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+4)
	for _, item := range items {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(RoadmapItemLabel(lang, item), fmt.Sprintf("%s%d", RoadmapCBOpen, item.ID)),
		))
	}
	if len(items) < models.MaxRoadmapsPerGoal {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAddTech), fmt.Sprintf("%s%d", RoadmapCBTechCreate, goalID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(
		// "✏️ Rename" reuses the Learning key on purpose — identical text,
		// and two keys rendering the same string collide in i18n's reverse
		// index (see TestCatalog_NoTextCollisionsWithinLanguage).
		buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonRename), fmt.Sprintf("%s%d", RoadmapCBGoalRename, goalID)),
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonArchiveGoal), fmt.Sprintf("%s%d", RoadmapCBGoalArchive, goalID)),
	))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBGoalsOpen)))
	return buttonbuilder.IK(rows...)
}

func RoadmapItemLabel(lang i18n.Lang, item models.RoadmapItem) string {
	key := i18n.KeyRoadmapButtonToggleOffFmt
	if item.Active {
		key = i18n.KeyRoadmapButtonToggleOnFmt
	}
	return i18n.T(lang, key, item.Name, item.DoneCards, item.TotalCards)
}

// RoadmapOrphansInlineMenu lists technologies attached to no goal.
func RoadmapOrphansInlineMenu(lang i18n.Lang, items []models.RoadmapItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(RoadmapItemLabel(lang, item), fmt.Sprintf("%s%d", RoadmapCBOpen, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// RoadmapDetailInlineMenu shows one technology's cards — each row is the card
// (tap to tick), a difficulty cycle, and a delete — then the
// technology-level actions. An unattached technology gets an extra "attach to
// a goal" row, and its Back leads to the orphan list rather than a goal.
func RoadmapDetailInlineMenu(lang i18n.Lang, item models.RoadmapItem, cards []models.RoadmapCardItem, aiEnabled bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(cards)+6)
	for i, c := range cards {
		row := buttonbuilder.IR(
			buttonbuilder.IB(CardButtonLabel(i+1, c), fmt.Sprintf("%s%d", RoadmapCBCardToggle, c.ID)),
			buttonbuilder.IB("🎚", fmt.Sprintf("%s%d", RoadmapCBCardDiff, c.ID)),
			buttonbuilder.IB("🗑", fmt.Sprintf("%s%d", RoadmapCBCardDelete, c.ID)),
		)
		if aiEnabled && !c.IsDone {
			// Only on pending cards: quizzing something already ticked off
			// is the one case where the button has nothing to offer, and a
			// fourth button on every row costs real width.
			row = append(row, buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAIQuiz), fmt.Sprintf("%s%d", RoadmapCBAIQuiz, c.ID)))
		}
		rows = append(rows, row)
	}

	toggleLabel := i18n.T(lang, i18n.KeyRoadmapLabelExcludedReminder)
	if item.Active {
		toggleLabel = i18n.T(lang, i18n.KeyRoadmapLabelIncludedInReminder)
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(toggleLabel, fmt.Sprintf("%s%d", RoadmapCBToggle, item.ID)),
	))
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAddCards), fmt.Sprintf("%s%d", RoadmapCBAddCards, item.ID)),
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonSetCriteria), fmt.Sprintf("%s%d", RoadmapCBSetCriteria, item.ID)),
	))
	if aiEnabled {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAIPlan), fmt.Sprintf("%s%d", RoadmapCBAIPlan, item.ID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAIPaste), fmt.Sprintf("%s%d", RoadmapCBAIPaste, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonRename), fmt.Sprintf("%s%d", RoadmapCBRename, item.ID)),
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonArchiveThis), fmt.Sprintf("%s%d", RoadmapCBArchiveThis, item.ID)),
	))

	back := RoadmapCBOrphansOpen
	if item.GoalID != nil {
		back = fmt.Sprintf("%s%d", RoadmapCBGoalOpen, *item.GoalID)
	} else {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAssignGoal), fmt.Sprintf("%s%d", RoadmapCBAssignOpen, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), back)))
	return buttonbuilder.IK(rows...)
}

// RoadmapAssignGoalInlineMenu offers the goals an unattached technology can
// move into.
func RoadmapAssignGoalInlineMenu(lang i18n.Lang, roadmapID int64, goals []models.RoadmapGoalItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(goals)+1)
	for _, g := range goals {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(GoalItemLabel(lang, g), fmt.Sprintf("%s%d:%d", RoadmapCBAssignPick, roadmapID, g.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), fmt.Sprintf("%s%d", RoadmapCBOpen, roadmapID)),
	))
	return buttonbuilder.IK(rows...)
}

// RoadmapQuizResultInlineMenu offers the one action a graded answer leads to:
// ticking the card. The toggle resolves its own technology, so the grade
// message turns into that technology's checklist when tapped.
func RoadmapQuizResultInlineMenu(lang i18n.Lang, cardID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonQuizDone), fmt.Sprintf("%s%d", RoadmapCBCardToggle, cardID)),
		),
	)
}

// ParseAssignPayload splits a "<roadmapID>:<goalID>" assign payload.
func ParseAssignPayload(data string) (roadmapID, goalID int64, ok bool) {
	raw := strings.TrimPrefix(data, RoadmapCBAssignPick)
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	roadmapID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	goalID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return roadmapID, goalID, true
}

// CardButtonLabel renders one card's button as its number from the list in the
// message text, plus its state.
//
// The text used to be on the button and was unreadable: an inline button is
// about 35 characters wide, four of them share a row, and a generated card runs
// to a hundred characters or more, so it arrived as "Re…tables,…". The number
// ties the button to the line above it, where the card is written out in full.
func CardButtonLabel(number int, c models.RoadmapCardItem) string {
	return fmt.Sprintf("%d %s", number, CardStateIcons(c))
}

// CardStateIcons is the difficulty (or a tick) plus the kind, shared by the
// button and the list line so the two cannot drift apart.
func CardStateIcons(c models.RoadmapCardItem) string {
	lead := difficultyIcon(c.Difficulty)
	if c.IsDone {
		lead = "✅"
	}
	return lead + kindIcon(c.Kind)
}

// Rune-based, not byte-based, so a Cyrillic or Arabic card isn't cut
// mid-character.
func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
}

// RoadmapArchiveInlineMenu lists archived goals and technologies together,
// each with restore/delete — mirrors learning.LearningArchiveInlineMenu.
func RoadmapArchiveInlineMenu(lang i18n.Lang, goals []models.RoadmapGoalItem, items []models.RoadmapItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, (len(goals)+len(items))*2+1)
	for _, g := range goals {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapArchiveGoalItemFmt, g.Name, g.TotalRoadmaps), "noop"),
		))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelRestore), fmt.Sprintf("%s%d", RoadmapCBArchiveGoalRestore, g.ID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelDeleteForever), fmt.Sprintf("%s%d", RoadmapCBArchiveGoalDelete, g.ID)),
		))
	}
	for _, item := range items {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapArchiveItemFmt, item.Name, item.DoneCards, item.TotalCards), "noop"),
		))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelRestore), fmt.Sprintf("%s%d", RoadmapCBArchiveRestore, item.ID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelDeleteForever), fmt.Sprintf("%s%d", RoadmapCBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// RoadmapDigestInlineMenu gives every card in a reminder push its own tick
// button, so a card can be closed straight from the notification. It uses
// RoadmapCBDigestToggle rather than the checklist's prefix so the handler
// knows to re-render this digest message.
func RoadmapDigestInlineMenu(lang i18n.Lang, cards []models.RoadmapDigestCard) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(cards)+1)
	for _, c := range cards {
		label := "✅ " + difficultyIcon(c.Difficulty) + kindIcon(c.Kind) + " " + truncateRunes(c.Text, digestButtonMaxRunes)
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(label, fmt.Sprintf("%s%d", RoadmapCBDigestToggle, c.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonGoals), RoadmapCBGoalsOpen),
	))
	return buttonbuilder.IK(rows...)
}

// Reply button menus

func RoadmapWaitingReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancelX))),
	)
}

// The mastery criteria is optional, so Skip sits next to Cancel.
func RoadmapCriteriaReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyRoadmapButtonSkipCriteria))),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancelX))),
	)
}

// "Done" leaves the flow, since the user may paste several messages in a row.
func RoadmapAddCardsReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonDone))),
	)
}

// A different emoji than Learning's "⏱ " so the two interval pickers' buttons
// can never resolve to each other if one is left on screen — dispatch is
// screen-gated either way, but this keeps it obvious.
const pushIntervalPrefix = "⏰ "

func RoadmapPushIntervalReplyMenu(lang i18n.Lang, intervals []int) tgbotapi.ReplyKeyboardMarkup {
	rows := make([][]tgbotapi.KeyboardButton, 0, (len(intervals)+1)/2+1)
	for i := 0; i < len(intervals); i += 2 {
		if i+1 < len(intervals) {
			rows = append(rows, buttonbuilder.RR(
				buttonbuilder.RB(FormatPushIntervalButton(intervals[i])),
				buttonbuilder.RB(FormatPushIntervalButton(intervals[i+1])),
			))
		} else {
			rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(FormatPushIntervalButton(intervals[i]))))
		}
	}
	rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack))))
	return buttonbuilder.RK(rows...)
}

// Not translated — a plain unit abbreviation, same reasoning as
// learning.FormatPushIntervalButton's own "min" suffix.
func FormatPushIntervalButton(minutes int) string {
	return fmt.Sprintf("%s%d min", pushIntervalPrefix, minutes)
}

func ParsePushIntervalButtonMinutes(text string) (int, bool) {
	if !strings.HasPrefix(text, pushIntervalPrefix) || !strings.HasSuffix(text, " min") {
		return 0, false
	}
	numStr := strings.TrimSuffix(strings.TrimPrefix(text, pushIntervalPrefix), " min")
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// PercentDone is done/total as a whole percentage, 0 for an empty plan rather
// than a division by zero.
func PercentDone(done, total int) int {
	if total <= 0 {
		return 0
	}
	return done * 100 / total
}
