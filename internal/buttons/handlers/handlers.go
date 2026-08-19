package handlers

import (
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/utils/tgclient"
	"tracker-bot/internal/utils/tgctx"

	"github.com/rs/zerolog/log"
)

type ReplyModule struct {
	bot          tgclient.BotAPI
	track        *handlers.Module
	subscription *handlers.Module
	entry        *handlers.Module
	profile      *handlers.Module
	learning     *handlers.Module
}

func New(bot tgclient.BotAPI, track *handlers.Module, subscription *handlers.Module, entry *handlers.Module, profile *handlers.Module, learning *handlers.Module) *ReplyModule {
	return &ReplyModule{
		bot:          bot,
		track:        track,
		subscription: subscription,
		entry:        entry,
		profile:      profile,
		learning:     learning,
	}
}

// replyButtonsByKey maps a translation key (see internal/i18n) to its
// handler — keyed by key rather than the literal button text, since the
// text is rendered per-user-language (see internal/buttons/entry.
// EntryReplyMenu) and would otherwise only match one language.
func (r *ReplyModule) replyButtonsByKey() map[string]func(*tgctx.MsgContext) {
	return map[string]func(*tgctx.MsgContext){
		i18n.KeyEntryButtonProfile:      r.handleShowProfileMenu,
		i18n.KeyEntryButtonTrack:        r.handleShowTrackingMenu,
		i18n.KeyEntryButtonLearning:     r.handleShowLearningMenu,
		i18n.KeyEntryButtonSubscription: r.handleShowSubscriptionMenu,
	}
}

func (r *ReplyModule) HandleReplyButtons(ctx *tgctx.MsgContext) bool {
	key, ok := i18n.Key(ctx.Language, ctx.Text)
	if !ok {
		log.Warn().Msgf("Unknown reply button: %s", ctx.Text)
		return false
	}

	handler, ok := r.replyButtonsByKey()[key]
	if !ok {
		return false
	}
	handler(ctx)
	return true
}

func (r *ReplyModule) handleShowProfileMenu(ctx *tgctx.MsgContext) {
	r.profile.ShowProfileMenu(ctx)
}

func (r *ReplyModule) handleShowTrackingMenu(ctx *tgctx.MsgContext) {
	r.track.ShowTrackingMenu(ctx)
}

func (r *ReplyModule) handleShowLearningMenu(ctx *tgctx.MsgContext) {
	r.learning.ShowLearningMenu(ctx)
}

func (r *ReplyModule) handleShowSubscriptionMenu(ctx *tgctx.MsgContext) {
	r.subscription.ShowSubscriptionMenu(ctx)
}
