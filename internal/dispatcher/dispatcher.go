package dispatcher

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	adminbtn "tracker-bot/internal/buttons/admin"
	challengebtn "tracker-bot/internal/buttons/challenge"
	learningbtn "tracker-bot/internal/buttons/learning"
	onboardingbtn "tracker-bot/internal/buttons/onboarding"
	profilebtn "tracker-bot/internal/buttons/profile"
	roadmapbtn "tracker-bot/internal/buttons/roadmap"
	trackbtn "tracker-bot/internal/buttons/track"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/internal/service"
	"tracker-bot/pkg/apptime"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	h "tracker-bot/internal/buttons/handlers"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/utils/tgctx"
)

type Dispatcher struct {
	bot          *tgbotapi.BotAPI
	appCtx       context.Context
	entrysvc     service.EntryService
	profilesvc   service.ProfileService
	track        *handlers.Module
	subscription *handlers.Module
	entry        *handlers.Module
	profile      *handlers.Module
	learning     *handlers.Module
	roadmap      *handlers.Module

	reply   *h.ReplyModule
	uistate service.UIStateService

	// sessions holds all per-user in-memory state (screen, timezone, pending
	// "waiting for X" flags, report scratch state) — see session.go.
	sessions *sessionStore
}

const (
	screenHome           = "home"
	screenTrackMain      = "track_main"
	screenTrackManage    = "track_manage"
	screenTrackTimer     = "track_timer"
	screenTrackTimerDel  = "track_timer_delete"
	screenAdmin          = "admin"
	screenAdminUsers     = "admin_users"
	screenTrackArchive   = "track_archive"
	screenCreateActivity = "create_activity"
	screenTrackReports   = "track_reports"

	screenLearningMain       = "learning_main"
	screenLearningWordBase   = "learning_word_base"
	screenLearningArchive    = "learning_archive"
	screenLearningTimer      = "learning_timer"
	screenLearningReviewPick = "learning_review_pick"

	screenRoadmapMain       = "roadmap_main"
	screenRoadmapGoals      = "roadmap_goals"
	screenRoadmapGoalDetail = "roadmap_goal_detail"
	screenRoadmapCards      = "roadmap_cards"
	screenRoadmapOrphans    = "roadmap_orphans"
	screenRoadmapArchive    = "roadmap_archive"
	screenRoadmapTimer      = "roadmap_timer"

	screenChallengeList     = "challenge_list"
	screenChallengeCreate   = "challenge_create"
	screenChallengeCalendar = "challenge_calendar"
	screenChallengeGrid     = "challenge_grid"
	screenChallengeArchive  = "challenge_archive"
)

func New(
	bot *tgbotapi.BotAPI,
	appCtx context.Context,
	entrysvc service.EntryService,
	profilesvc service.ProfileService,
	uistate service.UIStateService,
	track *handlers.Module,
	subscription *handlers.Module,
	entry *handlers.Module,
	profile *handlers.Module,
	learning *handlers.Module,
	roadmap *handlers.Module,
) *Dispatcher {
	if bot == nil {
		log.Fatal().Msg("Dispatcher: nil bot interfaces.BotAPI")
	}

	if appCtx == nil {
		appCtx = context.Background()
	}

	d := &Dispatcher{
		bot:          bot,
		appCtx:       appCtx,
		entrysvc:     entrysvc,
		profilesvc:   profilesvc,
		uistate:      uistate,
		track:        track,
		subscription: subscription,
		entry:        entry,
		profile:      profile,
		learning:     learning,
		roadmap:      roadmap,
		sessions:     newSessionStore(),
	}

	d.reply = h.New(bot, track, subscription, entry, profile, learning, roadmap)
	return d
}

// Run listens for Telegram updates and routes them by update type.
func (d *Dispatcher) Run() {
	go d.sweepSessionsLoop()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := d.bot.GetUpdatesChan(u)

	for update := range updates {
		switch {
		case update.Message != nil:
			d.handleMessage(update.Message)

		case update.CallbackQuery != nil:
			d.handleCallback(update.CallbackQuery)
		}
	}
}

// sweepSessionsLoop periodically evicts sessions for users who haven't
// messaged the bot in a while, so this process's memory doesn't grow forever.
func (d *Dispatcher) sweepSessionsLoop() {
	const (
		interval = time.Hour
		maxIdle  = 30 * 24 * time.Hour
	)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-d.appCtx.Done():
			return
		case <-ticker.C:
			d.sessions.sweep(maxIdle)
		}
	}
}

// ensureUser creates/loads user in DB and stores DB id in context.
func (d *Dispatcher) ensureUser(ctx *tgctx.MsgContext, chatID int64, from *tgbotapi.User) bool {
	if from == nil {
		return false
	}
	ctx.Username = from.UserName

	in := &models.UserInput{
		TgUserID: int64(from.ID),
		UserName: &from.UserName,
	}

	dbID, isNew, err := d.entrysvc.EnsureUser(ctx.Ctx, in)
	if err != nil {
		log.Error().Err(err).Msg("ensure user failed")
		// Language isn't known yet at this point (loading it is further
		// down, and failed here before we got there) — Default is the best
		// we can do for this rare DB-error path.
		out := tgbotapi.NewMessage(chatID, i18n.T(i18n.Default, i18n.KeyCommonGenericError))
		_, _ = d.bot.Send(out)
		return false
	}
	ctx.DBUserID = dbID
	ctx.IsNewUser = isNew

	sess := d.sessions.get(ctx.UserID)
	sess.dbID = dbID

	// Cold session (first message since process start/restart, or after an
	// idle eviction) — restore the screen the user was actually on instead
	// of defaulting to "".
	if !sess.screenLoaded && d.uistate != nil {
		if screen, err := d.uistate.GetScreen(ctx.Ctx, dbID); err != nil {
			log.Error().Err(err).Int64("user_id", dbID).Msg("load screen failed")
		} else {
			sess.screen = screen
			sess.screenLoaded = true
		}
	}

	// Same cold-session treatment for the user's detected timezone and
	// interface language — one GetProfileStats call covers both.
	if (!sess.tzLoaded || !sess.langLoaded) && d.profilesvc != nil {
		if stats, err := d.profilesvc.GetProfileStats(ctx.Ctx, ctx.UserID); err != nil {
			log.Error().Err(err).Int64("user_id", ctx.UserID).Msg("load profile timezone/language failed")
		} else {
			if stats.TimeZone != nil {
				sess.tz = *stats.TimeZone
				sess.tzLoaded = true
			}
			if stats.Language != nil {
				sess.lang = *stats.Language
				sess.langLoaded = true
			}
		}
	}
	ctx.Location = apptime.Resolve(sess.tz)
	ctx.Language = i18n.Normalize(sess.lang)
	return true
}

// userLocation returns the resolved timezone for a user, falling back to
// apptime.Location if they haven't shared their location yet.
func (d *Dispatcher) userLocation(userID int64) *time.Location {
	return apptime.Resolve(d.sessions.get(userID).tz)
}

// newMessageContext converts Telegram message into internal context.
func (d *Dispatcher) newMessageContext(msg *tgbotapi.Message) *tgctx.MsgContext {
	ctx := &tgctx.MsgContext{
		Ctx:    d.appCtx,
		ChatID: msg.Chat.ID,
		Text:   msg.Text,
	}

	if msg.From != nil {
		ctx.UserID = int64(msg.From.ID)
	}

	return ctx
}

// handleMessage processes incoming text/command updates.
func (d *Dispatcher) handleMessage(msg *tgbotapi.Message) {
	if msg == nil || msg.From == nil {
		return
	}

	mctx := d.newMessageContext(msg)

	if !d.ensureUser(mctx, msg.Chat.ID, msg.From) {
		return
	}

	// Waiting for a shared location to detect timezone (from ProfileCBEditTimeZone).
	if sess := d.sessions.get(mctx.UserID); sess.waitingLocation {
		if msg.Location != nil {
			sess.waitingLocation = false
			d.profile.ProcessLocationTimeZone(mctx, msg.Location.Latitude, msg.Location.Longitude)
			// Timezone just changed — force the next message to re-read it
			// from DB instead of using the now-stale cached value.
			sess.tzLoaded = false
			return
		}
		if key, ok := i18n.Key(mctx.Language, mctx.Text); ok && key == i18n.KeyCommonCancel {
			sess.waitingLocation = false
			hide := tgbotapi.NewMessage(mctx.ChatID, i18n.T(mctx.Language, i18n.KeyCommonCancelled))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.profile.ShowProfileMenu(mctx)
			return
		}
		_, _ = d.bot.Send(tgbotapi.NewMessage(mctx.ChatID, i18n.T(mctx.Language, i18n.KeyProfileTimezoneInvalidTap)))
		return
	}

	// Handle slash commands first, so they are not treated as plain reply button text.
	if msg.IsCommand() {
		d.handleCommand(msg, mctx)
		return
	}

	// Then handle temporary user states (e.g. waiting for activity name).
	if d.handleUserState(mctx) {
		return
	}

	// Then process reply keyboard buttons.
	if key, ok := i18n.Key(mctx.Language, mctx.Text); ok && key == i18n.KeyEntryButtonTrack {
		d.setScreen(mctx.UserID, screenTrackMain)
	}
	if key, ok := i18n.Key(mctx.Language, mctx.Text); ok && key == i18n.KeyEntryButtonLearning {
		d.setScreen(mctx.UserID, screenLearningMain)
	}
	if key, ok := i18n.Key(mctx.Language, mctx.Text); ok && key == i18n.KeyEntryButtonRoadmap {
		d.setScreen(mctx.UserID, screenRoadmapMain)
	}
	if d.reply != nil && d.reply.HandleReplyButtons(mctx) {
		return
	}

	// Fallback for regular text messages.
	d.handleText(mctx)
}

// handleCallback processes incoming inline callback updates.
func (d *Dispatcher) handleCallback(q *tgbotapi.CallbackQuery) {
	if q == nil || q.Message == nil || q.From == nil {
		return
	}

	ack := tgbotapi.NewCallback(q.ID, "")
	if _, err := d.bot.Request(ack); err != nil {
		log.Error().Err(err).Msg("callback ack failed")
	}

	mctx := &tgctx.MsgContext{
		Ctx:       d.appCtx,
		ChatID:    q.Message.Chat.ID,
		Text:      q.Data,
		UserID:    int64(q.From.ID),
		MessageID: q.Message.MessageID,
	}

	if !d.ensureUser(mctx, q.Message.Chat.ID, q.From) {
		return
	}

	// Shared "back to home" action used by every module's entry inline menu
	// (Track/Profile/Subscription/Learning), not just Track — handled here
	// rather than inside handleTrackCallback.
	if q.Data == "go_home" {
		d.setScreen(mctx.UserID, screenHome)
		d.entry.ShowHomeMenu(mctx)
		return
	}

	if q.Data == profilebtn.ProfileCBEditTimeZone {
		d.sessions.get(mctx.UserID).waitingLocation = true
		d.profile.ShowLocationRequest(mctx)
		return
	}

	if q.Data == profilebtn.ProfileCBEditLanguage {
		d.sessions.get(mctx.UserID).waitingLanguage = true
		d.profile.ShowLanguagePicker(mctx)
		return
	}

	if q.Data == profilebtn.ProfileCBRefresh {
		d.profile.ShowProfileMenu(mctx)
		return
	}

	// "back_to_main"/"noop" are used as raw literals (not "track:"-prefixed)
	// by several inline keyboards in internal/buttons/track/keyboard_build.go.
	// Without this they never reach handleTrackCallback below, so every
	// inline "◀ Back" wired to back_to_main silently did nothing.
	if strings.HasPrefix(q.Data, "track:") || strings.HasPrefix(q.Data, "act_toggle_:") ||
		q.Data == "back_to_main" || q.Data == "noop" {
		d.handleTrackCallback(mctx, q.Data)
		return
	}

	if strings.HasPrefix(q.Data, "learning:") {
		d.handleLearningCallback(mctx, q.Data)
		return
	}

	if strings.HasPrefix(q.Data, "roadmap:") {
		d.handleRoadmapCallback(mctx, q.Data)
		return
	}

	if strings.HasPrefix(q.Data, "challenge:") {
		d.handleChallengeCallback(mctx, q.Data)
		return
	}

	if strings.HasPrefix(q.Data, "onboarding:") {
		d.handleOnboardingCallback(mctx, q.Data)
		return
	}

	if strings.HasPrefix(q.Data, "admin:") {
		if !d.entry.IsAdmin(mctx) {
			return
		}
		d.handleAdminCallback(mctx, q.Data)
		return
	}

	if d.reply != nil && d.reply.HandleReplyButtons(mctx) {
		return
	}
}

// handleAdminCallback routes admin-only inline callbacks: opening the admin
// screen (e.g. from the Profile screen's "👑 Admin" button) and paging the
// users list. Caller must have already checked entry.IsAdmin.
func (d *Dispatcher) handleAdminCallback(ctx *tgctx.MsgContext, data string) {
	switch {
	case data == adminbtn.CBOpen:
		d.setScreen(ctx.UserID, screenAdmin)
		d.entry.ShowAdminMenu(ctx)
	case strings.HasPrefix(data, adminbtn.CBUsersPage):
		offset, ok := parseCallbackID(data, adminbtn.CBUsersPage)
		if !ok {
			return
		}
		d.entry.ShowAdminUsersMenuInPlace(ctx, int(offset))
	case strings.HasPrefix(data, adminbtn.CBUserDeleteAsk):
		id, ok := parseCallbackID(data, adminbtn.CBUserDeleteAsk)
		if !ok {
			return
		}
		d.entry.ShowAdminUserDeleteConfirm(ctx, id)
	case strings.HasPrefix(data, adminbtn.CBUserDeleteConfirm):
		id, ok := parseCallbackID(data, adminbtn.CBUserDeleteConfirm)
		if !ok {
			return
		}
		d.entry.DeleteAdminUser(ctx, id)
	case strings.HasPrefix(data, adminbtn.CBUserDetail):
		id, ok := parseCallbackID(data, adminbtn.CBUserDetail)
		if !ok {
			return
		}
		d.entry.ShowAdminUserDetail(ctx, id, true)
	case data == adminbtn.CBBroadcastConfirm:
		sess := d.sessions.get(ctx.UserID)
		text := sess.pendingBroadcastText
		sess.pendingBroadcastText = ""
		d.entry.SendAdminBroadcast(ctx, text)
	case data == adminbtn.CBBroadcastCancel:
		d.sessions.get(ctx.UserID).pendingBroadcastText = ""
		d.entry.CancelAdminBroadcast(ctx)
	}
}

// handleUserState handles temporary per-user states (FSM-like flow).
func (d *Dispatcher) handleUserState(ctx *tgctx.MsgContext) bool {
	sess := d.sessions.get(ctx.UserID)

	if sess.waitingAdminBroadcastText {
		if ctx.Text == adminbtn.LabelCancel {
			sess.waitingAdminBroadcastText = false
			hide := tgbotapi.NewMessage(ctx.ChatID, "❌ Broadcast cancelled.")
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.entry.ShowAdminMenu(ctx)
			return true
		}
		text := strings.TrimSpace(ctx.Text)
		if text == "" {
			_, _ = d.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Message can't be empty. Try again:"))
			return true
		}
		if d.entry.ShowAdminBroadcastConfirm(ctx, text) {
			sess.waitingAdminBroadcastText = false
			sess.pendingBroadcastText = text
		}
		return true
	}

	if sess.waitingActivityName {
		if d.isTrackButtonText(ctx) {
			// A tap on a real nav/action button (Back, Home, ...) means the
			// user changed their mind about typing a name — cancel the
			// prompt and let it act like a normal button press, rather than
			// blocking it with no way out. screenTrackManage is where
			// "Create activity" was launched from, so Back lands there.
			sess.waitingActivityName = false
			d.setScreen(ctx.UserID, screenTrackManage)
			return false
		}
		done := d.track.ProcessCreateActivity(ctx)
		if done {
			sess.waitingActivityName = false
			d.setScreen(ctx.UserID, screenTrackMain)
		}
		return true
	}
	if sess.waitingCustomTimerMinutes {
		if d.isTrackButtonText(ctx) {
			// Same reasoning as waitingActivityName above — let a real nav
			// button tap cancel the prompt instead of dead-ending on it.
			// Screen is already screenTrackTimer, which Back/Home both
			// already handle correctly.
			sess.waitingCustomTimerMinutes = false
			return false
		}
		if d.track.ProcessCreateCustomTimer(ctx) {
			sess.waitingCustomTimerMinutes = false
		}
		return true
	}
	if sess.waitingLanguage {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancel {
			sess.waitingLanguage = false
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonCancelled))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.profile.ShowProfileMenu(ctx)
			return true
		}
		if d.profile.ProcessLanguageSelection(ctx) {
			sess.waitingLanguage = false
			// Language just changed — force the next message to re-read it
			// from DB instead of using the now-stale cached value.
			sess.langLoaded = false
		}
		return true
	}
	if sess.waitingChallengeName {
		if ctx.Text == challengebtn.ReplyCancel {
			sess.waitingChallengeName = false
			hide := tgbotapi.NewMessage(ctx.ChatID, "❌ Cancelled.")
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.setScreen(ctx.UserID, screenChallengeList)
			d.entry.ShowChallengesMenu(ctx)
			return true
		}
		name, ok := d.entry.ProcessCreateChallengeName(ctx)
		if ok {
			sess.waitingChallengeName = false
			sess.pendingChallengeName = name
			sess.challengeCalMonth = apptime.NowIn(d.userLocation(ctx.UserID))
			sess.challengeCalFrom = time.Time{}
			sess.challengeCalTo = time.Time{}
			d.setScreen(ctx.UserID, screenChallengeCalendar)
			d.entry.ShowCreateChallengeCalendar(ctx, sess.challengeCalMonth, sess.challengeCalFrom, sess.challengeCalTo)
		}
		return true
	}
	if sess.waitingLearningCollectionName {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancelX {
			sess.waitingLearningCollectionName = false
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonCancelled))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.setScreen(ctx.UserID, screenLearningMain)
			d.learning.ShowLearningMenu(ctx)
			return true
		}
		id, done := d.learning.ProcessCreateCollectionName(ctx)
		if done {
			sess.waitingLearningCollectionName = false
			if id > 0 {
				sess.waitingLearningWords = true
				sess.learningCollectionID = id
				d.setScreen(ctx.UserID, screenLearningWordBase)
			} else {
				d.setScreen(ctx.UserID, screenLearningMain)
				d.learning.ShowLearningMenu(ctx)
			}
		}
		return true
	}
	if sess.waitingLearningWords {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonDone {
			sess.waitingLearningWords = false
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningAddWordsDoneNotice))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.learning.ShowCollectionDetail(ctx, sess.learningCollectionID, false)
			return true
		}
		d.learning.ProcessAddWords(ctx, sess.learningCollectionID)
		return true
	}
	if sess.waitingLearningRenameCollection {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancelX {
			sess.waitingLearningRenameCollection = false
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonCancelled))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.learning.ShowCollectionDetail(ctx, sess.learningCollectionID, false)
			return true
		}
		if d.learning.ProcessRenameCollection(ctx, sess.learningCollectionID) {
			sess.waitingLearningRenameCollection = false
		}
		return true
	}
	if sess.waitingRoadmapGoalName {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancelX {
			sess.waitingRoadmapGoalName = false
			d.cancelRoadmapPrompt(ctx)
			d.setScreen(ctx.UserID, screenRoadmapMain)
			d.roadmap.ShowRoadmapMenu(ctx)
			return true
		}
		if id, done := d.roadmap.ProcessCreateGoalName(ctx); done {
			sess.waitingRoadmapGoalName = false
			if id > 0 {
				sess.roadmapGoalID = id
				d.setScreen(ctx.UserID, screenRoadmapGoalDetail)
			} else {
				d.setScreen(ctx.UserID, screenRoadmapMain)
				d.roadmap.ShowRoadmapMenu(ctx)
			}
		}
		return true
	}
	if sess.waitingRoadmapGoalRename {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancelX {
			sess.waitingRoadmapGoalRename = false
			d.cancelRoadmapPrompt(ctx)
			d.roadmap.ShowRoadmapGoalDetail(ctx, sess.roadmapGoalID, false)
			return true
		}
		if d.roadmap.ProcessRenameGoal(ctx, sess.roadmapGoalID) {
			sess.waitingRoadmapGoalRename = false
		}
		return true
	}
	if sess.waitingRoadmapName {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancelX {
			sess.waitingRoadmapName = false
			d.cancelRoadmapPrompt(ctx)
			d.roadmap.ShowRoadmapGoalDetail(ctx, sess.roadmapGoalID, false)
			return true
		}
		id, done := d.roadmap.ProcessCreateRoadmapName(ctx, sess.roadmapGoalID)
		if done {
			sess.waitingRoadmapName = false
			if id > 0 {
				// The name prompt chains straight into the criteria prompt,
				// which then chains into the card paste step.
				sess.waitingRoadmapCriteria = true
				sess.roadmapID = id
				d.setScreen(ctx.UserID, screenRoadmapCards)
			} else {
				d.setScreen(ctx.UserID, screenRoadmapGoalDetail)
				d.roadmap.ShowRoadmapGoalDetail(ctx, sess.roadmapGoalID, false)
			}
		}
		return true
	}
	if sess.waitingRoadmapCriteria {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok {
			switch key {
			case i18n.KeyCommonCancelX:
				// Cancel here abandons the criteria step only — the
				// technology already exists, so land the user on it rather
				// than back at a menu with no sign of what happened.
				sess.waitingRoadmapCriteria = false
				d.cancelRoadmapPrompt(ctx)
				d.roadmap.ShowRoadmapDetail(ctx, sess.roadmapID, false)
				return true
			case i18n.KeyRoadmapButtonSkipCriteria:
				sess.waitingRoadmapCriteria = false
				d.roadmap.NoticeRoadmapCriteriaSkipped(ctx)
				sess.waitingRoadmapCards = true
				d.roadmap.PromptAddRoadmapCards(ctx, sess.roadmapID, true)
				return true
			}
		}
		if d.roadmap.ProcessRoadmapCriteria(ctx, sess.roadmapID) {
			sess.waitingRoadmapCriteria = false
			sess.waitingRoadmapCards = true
			d.roadmap.PromptAddRoadmapCards(ctx, sess.roadmapID, true)
		}
		return true
	}
	if sess.waitingRoadmapCards {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonDone {
			sess.waitingRoadmapCards = false
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAddCardsDoneNotice))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.roadmap.ShowRoadmapDetail(ctx, sess.roadmapID, false)
			return true
		}
		d.roadmap.ProcessAddRoadmapCards(ctx, sess.roadmapID)
		return true
	}
	if sess.waitingRoadmapRename {
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonCancelX {
			sess.waitingRoadmapRename = false
			d.cancelRoadmapPrompt(ctx)
			d.roadmap.ShowRoadmapDetail(ctx, sess.roadmapID, false)
			return true
		}
		if d.roadmap.ProcessRenameRoadmap(ctx, sess.roadmapID) {
			sess.waitingRoadmapRename = false
		}
		return true
	}
	if sess.waitingPeriodRange {
		from, to, err := parseDateRange(ctx.Text)
		if err != nil {
			_, _ = d.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPeriodRangeInvalidFmt)))
			return true
		}
		sess.reportFrom = from
		sess.reportTo = to
		sess.waitingPeriodRange = false

		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPeriodRangeSetConfirm, from.Format("2006-01-02"), to.Format("2006-01-02")))
		_, _ = d.bot.Send(msg)
		return true
	}

	return false
}

// handleCommand routes slash commands.
func (d *Dispatcher) handleCommand(msg *tgbotapi.Message, ctx *tgctx.MsgContext) {
	cmd := msg.Command()

	switch cmd {
	case "start":
		d.setScreen(ctx.UserID, screenHome)
		d.entry.ShowEntryMenu(ctx)
		return

	case "help":
		out := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonHelpText))
		if _, err := d.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("send help failed")
		}
		return

	case "admin":
		if !d.entry.IsAdmin(ctx) {
			out := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonUnknownCommand))
			if _, err := d.bot.Send(out); err != nil {
				log.Error().Err(err).Msg("send unknown command failed")
			}
			return
		}
		d.setScreen(ctx.UserID, screenAdmin)
		d.entry.ShowAdminMenu(ctx)
		return

	default:
		out := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonUnknownCommand))
		if _, err := d.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("send unknown command failed")
		}
		return
	}
}

// handleText routes plain text based on current screen and reply buttons.
func (d *Dispatcher) handleText(ctx *tgctx.MsgContext) {
	// Timer interval buttons carry their minutes in the button text itself
	// ("⏱ 15 min" / "🗑 15 min") rather than through inline callback_data —
	// checked here, gated by screen, before the exact-match switch below.
	if d.isScreen(ctx.UserID, screenTrackTimer) {
		if minutes, ok := trackbtn.ParseTimerButtonMinutes(ctx.Language, ctx.Text, trackbtn.TrackTimerActivatePrefix); ok {
			d.track.ActivateTrackTimer(ctx, minutes)
			d.setScreen(ctx.UserID, screenHome)
			return
		}
	}
	if d.isScreen(ctx.UserID, screenTrackTimerDel) {
		if minutes, ok := trackbtn.ParseTimerButtonMinutes(ctx.Language, ctx.Text, trackbtn.TrackTimerDeletePrefix); ok {
			d.track.DeleteCustomTimer(ctx, minutes)
			d.setScreen(ctx.UserID, screenTrackTimer)
			return
		}
	}

	// Learning review-push interval picker.
	if d.isScreen(ctx.UserID, screenLearningTimer) {
		if minutes, ok := learningbtn.ParsePushIntervalButtonMinutes(ctx.Text); ok {
			d.setScreen(ctx.UserID, screenLearningMain)
			d.learning.ActivateReviews(ctx, minutes)
			return
		}
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonBack {
			d.setScreen(ctx.UserID, screenLearningMain)
			hide := tgbotapi.NewMessage(ctx.ChatID, " ")
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.learning.ShowLearningMenu(ctx)
			return
		}
	}

	// Roadmap reminder-interval picker.
	if d.isScreen(ctx.UserID, screenRoadmapTimer) {
		if minutes, ok := roadmapbtn.ParsePushIntervalButtonMinutes(ctx.Text); ok {
			d.setScreen(ctx.UserID, screenRoadmapMain)
			d.roadmap.ActivateRoadmapReminders(ctx, minutes)
			return
		}
		if key, ok := i18n.Key(ctx.Language, ctx.Text); ok && key == i18n.KeyCommonBack {
			d.setScreen(ctx.UserID, screenRoadmapMain)
			hide := tgbotapi.NewMessage(ctx.ChatID, " ")
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = d.bot.Send(hide)
			d.roadmap.ShowRoadmapMenu(ctx)
			return
		}
	}

	// Buttons whose text is translated (see internal/i18n) can't be matched
	// as literal switch cases below — resolve to a stable key first (see
	// handleTranslatedButton). Falls through to the literal-text switch for
	// buttons not yet converted (Admin's "👥 Users").
	if key, ok := i18n.Key(ctx.Language, ctx.Text); ok {
		if d.handleTranslatedButton(ctx, key) {
			return
		}
	}

	switch ctx.Text {
	case adminbtn.ReplyButtonUsers:
		if !d.entry.IsAdmin(ctx) || !d.isScreen(ctx.UserID, screenAdmin) {
			d.replyUseButtons(ctx)
			return
		}
		d.setScreen(ctx.UserID, screenAdminUsers)
		d.entry.ShowAdminUsersMenu(ctx, 0)
		return
	case adminbtn.ReplyButtonOverview:
		if !d.entry.IsAdmin(ctx) || !d.isScreen(ctx.UserID, screenAdmin) {
			d.replyUseButtons(ctx)
			return
		}
		d.entry.ShowAdminOverview(ctx)
		return
	case adminbtn.ReplyButtonBroadcast:
		if !d.entry.IsAdmin(ctx) || !d.isScreen(ctx.UserID, screenAdmin) {
			d.replyUseButtons(ctx)
			return
		}
		d.sessions.get(ctx.UserID).waitingAdminBroadcastText = true
		d.entry.PromptAdminBroadcast(ctx)
		return
	}

	out := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonFallback))
	if _, err := d.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("send fallback failed")
	}
}

// handleTranslatedButton routes a reply-button tap that resolved to a
// stable i18n key (see i18n.Key) rather than literal Go-constant text.
// Returns false when key isn't one of the buttons handled here, so the
// caller falls through to the literal-text switch for buttons not yet
// localized (Reports, Admin's "👥 Users").
func (d *Dispatcher) handleTranslatedButton(ctx *tgctx.MsgContext, key string) bool {
	switch key {
	case i18n.KeyCommonAdmin:
		if !d.entry.IsAdmin(ctx) {
			d.replyUseButtons(ctx)
			return true
		}
		d.setScreen(ctx.UserID, screenAdmin)
		d.entry.ShowAdminMenu(ctx)
		return true
	case i18n.KeyTrackButtonActivityDelete:
		if !d.isScreen(ctx.UserID, screenTrackManage) {
			d.replyUseButtons(ctx)
			return true
		}
		d.track.DeleteSelectedActivities(ctx)
		return true
	case i18n.KeyTrackButtonActivityActivate:
		if !d.isScreen(ctx.UserID, screenTrackManage, screenTrackMain) {
			d.replyUseButtons(ctx)
			return true
		}
		d.setScreen(ctx.UserID, screenTrackTimer)
		d.track.ShowTrackTimerMenu(ctx)
		return true
	case i18n.KeyTrackButtonViewArchive:
		d.setScreen(ctx.UserID, screenTrackArchive)
		d.track.ShowArchiveMenu(ctx)
		return true
	case i18n.KeyTrackButtonTimerCreate:
		if !d.isScreen(ctx.UserID, screenTrackTimer) {
			d.replyUseButtons(ctx)
			return true
		}
		d.sessions.get(ctx.UserID).waitingCustomTimerMinutes = true
		d.track.PromptCreateCustomTimer(ctx)
		return true
	case i18n.KeyTrackButtonTimerDelete:
		if !d.isScreen(ctx.UserID, screenTrackTimer) {
			d.replyUseButtons(ctx)
			return true
		}
		if d.track.ShowTrackTimerDeleteMenu(ctx) {
			d.setScreen(ctx.UserID, screenTrackTimerDel)
		}
		return true
	case i18n.KeyCommonBack:
		switch {
		case d.isScreen(ctx.UserID, screenTrackReports):
			d.track.ShowReportsHub(ctx, false)
		case d.isScreen(ctx.UserID, screenTrackTimerDel):
			d.setScreen(ctx.UserID, screenTrackTimer)
			d.track.ShowTrackTimerMenu(ctx)
		case d.isScreen(ctx.UserID, screenAdminUsers):
			d.setScreen(ctx.UserID, screenAdmin)
			d.entry.ShowAdminMenu(ctx)
		case d.isScreen(ctx.UserID, screenAdmin):
			d.setScreen(ctx.UserID, screenHome)
			d.entry.ShowHomeMenu(ctx)
		case d.isScreen(ctx.UserID, screenRoadmapGoals, screenRoadmapGoalDetail, screenRoadmapCards, screenRoadmapOrphans, screenRoadmapArchive):
			d.setScreen(ctx.UserID, screenRoadmapMain)
			d.roadmap.ShowRoadmapMenu(ctx)
		case d.isScreen(ctx.UserID, screenTrackManage, screenTrackArchive, screenTrackTimer):
			d.setScreen(ctx.UserID, screenTrackMain)
			d.track.ShowTrackingMenu(ctx)
		default:
			d.replyUseButtons(ctx)
		}
		return true
	case i18n.KeyTrackButtonSelectActivity:
		d.setScreen(ctx.UserID, screenTrackManage)
		d.track.ShowTrackActivitySelectionMenu(ctx)
		return true
	case i18n.KeyCommonHome:
		d.setScreen(ctx.UserID, screenHome)
		d.entry.ShowHomeMenu(ctx)
		return true
	case i18n.KeyTrackButtonToday:
		if !d.isScreen(ctx.UserID, screenTrackReports) {
			d.replyUseButtons(ctx)
			return true
		}
		d.track.ShowTodayReport(ctx)
		return true
	case i18n.KeyTrackButtonHeatmap:
		if !d.isScreen(ctx.UserID, screenTrackReports) {
			d.replyUseButtons(ctx)
			return true
		}
		d.track.ShowHeatmap(ctx, false)
		return true
	case i18n.KeyTrackButtonCalendar:
		if !d.isScreen(ctx.UserID, screenTrackReports) {
			d.replyUseButtons(ctx)
			return true
		}
		d.setScreen(ctx.UserID, screenTrackReports)
		d.ensurePeriodDefaults(ctx.UserID)
		d.showPeriodMenu(ctx)
		return true
	}
	return false
}

// handleTrackCallback routes track-related inline callbacks.
func (d *Dispatcher) handleTrackCallback(ctx *tgctx.MsgContext, data string) {
	sess := d.sessions.get(ctx.UserID)

	switch {
	case data == "noop":
		return
	case data == "back_to_main":
		d.setScreen(ctx.UserID, screenTrackMain)
		d.track.ShowTrackingMenu(ctx)
	case data == trackbtn.TrackCBActivitySelect:
		d.setScreen(ctx.UserID, screenTrackManage)
		d.track.ShowTrackActivitySelectionMenu(ctx)
	case data == trackbtn.TrackCBReportSummary, data == trackbtn.TrackCBReportsHub:
		d.setScreen(ctx.UserID, screenTrackReports)
		d.track.ShowReportsHub(ctx, true)
	case data == trackbtn.TrackCBReportsToday:
		d.setScreen(ctx.UserID, screenTrackReports)
		d.track.ShowTodayReport(ctx)
	case data == trackbtn.TrackCBReportsTodayBySelected:
		d.setScreen(ctx.UserID, screenTrackReports)
		d.track.ShowTodayReportBySelected(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBReportsTodaySelToggle):
		d.setScreen(ctx.UserID, screenTrackReports)
		id, ok := parseCallbackID(data, trackbtn.TrackCBReportsTodaySelToggle)
		if !ok {
			return
		}
		sel := d.getReportSelected(ctx.UserID)
		toggleSelected(sel, id)
		d.track.ShowTodaySelectActivities(ctx, sel)
	case data == trackbtn.TrackCBReportsTodaySelBuild:
		d.setScreen(ctx.UserID, screenTrackReports)
		ids := selectedIDs(d.getReportSelected(ctx.UserID))
		loc := d.userLocation(ctx.UserID)
		today := apptime.NowIn(loc)
		from := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
		to := from
		d.track.ShowPeriodChartReport(ctx, from, to, ids)
	case data == trackbtn.TrackCBReportsPeriodOpen:
		d.setScreen(ctx.UserID, screenTrackReports)
		d.ensurePeriodDefaults(ctx.UserID)
		d.showPeriodMenu(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBReportsPeriodToggle):
		d.setScreen(ctx.UserID, screenTrackReports)
		id, ok := parseCallbackID(data, trackbtn.TrackCBReportsPeriodToggle)
		if !ok {
			return
		}
		sel := d.getReportSelected(ctx.UserID)
		toggleSelected(sel, id)
		d.showPeriodMenu(ctx)
	case data == trackbtn.TrackCBReportsPeriodSetRange:
		if sess.reportCalMonth.IsZero() {
			sess.reportCalMonth = apptime.NowIn(d.userLocation(ctx.UserID))
		}
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalPrev:
		sess.reportCalMonth = d.calendarMonth(ctx.UserID).AddDate(0, -1, 0)
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalNext:
		sess.reportCalMonth = d.calendarMonth(ctx.UserID).AddDate(0, 1, 0)
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalPrevYear:
		sess.reportCalMonth = d.calendarMonth(ctx.UserID).AddDate(-1, 0, 0)
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalNextYear:
		sess.reportCalMonth = d.calendarMonth(ctx.UserID).AddDate(1, 0, 0)
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalThisMonth:
		loc := d.userLocation(ctx.UserID)
		now := apptime.NowIn(loc)
		sess.reportCalMonth = now
		sess.reportCalFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		sess.reportCalTo = time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, loc)
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalThisYear:
		loc := d.userLocation(ctx.UserID)
		now := apptime.NowIn(loc)
		sess.reportCalMonth = now
		sess.reportCalFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		sess.reportCalTo = time.Date(now.Year(), 12, 31, 0, 0, 0, 0, loc)
		d.showPeriodCalendar(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBReportsCalPick):
		raw := strings.TrimPrefix(data, trackbtn.TrackCBReportsCalPick)
		day, err := apptime.ParseDay(raw, d.userLocation(ctx.UserID))
		if err != nil {
			return
		}
		if sess.reportCalFrom.IsZero() || !sess.reportCalTo.IsZero() {
			sess.reportCalFrom = day
			sess.reportCalTo = time.Time{}
		} else {
			sess.reportCalTo = day
			if sess.reportCalTo.Before(sess.reportCalFrom) {
				sess.reportCalFrom, sess.reportCalTo = sess.reportCalTo, sess.reportCalFrom
			}
		}
		d.showPeriodCalendar(ctx)
	case data == trackbtn.TrackCBReportsCalDone:
		if sess.reportCalFrom.IsZero() || sess.reportCalTo.IsZero() {
			_, _ = d.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCalendarPickBothDays)))
			return
		}
		sess.reportFrom = sess.reportCalFrom
		sess.reportTo = sess.reportCalTo
		d.showPeriodMenu(ctx)
	case data == trackbtn.TrackCBReportsCalCancel:
		d.showPeriodMenu(ctx)
	case data == trackbtn.TrackCBReportsPeriodText:
		d.setScreen(ctx.UserID, screenTrackReports)
		ids := selectedIDs(d.getReportSelected(ctx.UserID))
		d.track.ShowPeriodTextReport(ctx, sess.reportFrom, sess.reportTo, ids, true)
	case data == trackbtn.TrackCBReportsPeriodChart:
		d.setScreen(ctx.UserID, screenTrackReports)
		ids := selectedIDs(d.getReportSelected(ctx.UserID))
		d.track.ShowPeriodChartReport(ctx, sess.reportFrom, sess.reportTo, ids)
	case data == trackbtn.TrackCBReportsBackHub:
		d.setScreen(ctx.UserID, screenTrackReports)
		d.track.ShowReportsHub(ctx, true)
	case data == trackbtn.TrackCBActivityCreate:
		sess.waitingActivityName = true
		d.setScreen(ctx.UserID, screenCreateActivity)
		d.track.PromptCreateActivity(ctx)
	case data == trackbtn.TrackCBArchiveOpen:
		d.setScreen(ctx.UserID, screenTrackArchive)
		d.track.ShowArchiveMenu(ctx)
	case data == trackbtn.TrackCBOpenArchive:
		d.setScreen(ctx.UserID, screenTrackArchive)
		d.track.ShowArchiveMenuInPlace(ctx)
	case data == trackbtn.TrackCBOpenActivities:
		d.setScreen(ctx.UserID, screenTrackManage)
		d.track.ShowTrackActivitySelectionMenuInPlace(ctx)
	case data == trackbtn.TrackCBCreateAnother:
		sess.waitingActivityName = true
		d.setScreen(ctx.UserID, screenCreateActivity)
		d.track.PromptCreateActivity(ctx)
	case data == trackbtn.TrackCBArchiveSelected:
		if !d.isScreen(ctx.UserID, screenTrackManage) {
			d.closeInlineMenu(ctx, i18n.T(ctx.Language, i18n.KeyTrackManageMenuClosed))
			return
		}
		d.setScreen(ctx.UserID, screenTrackArchive)
		d.track.ArchiveSelectedActivitiesInPlace(ctx)
	case data == trackbtn.TrackCBArchiveToActive:
		d.setScreen(ctx.UserID, screenTrackManage)
		d.track.ShowTrackActivitySelectionMenuInPlace(ctx)
	case data == trackbtn.TrackCBPromptStopTimer:
		d.track.StopTrackTimer(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBPromptActivity):
		d.track.RecordPromptAnswer(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBArchiveRestore):
		d.track.RestoreArchivedActivity(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBArchiveDelete):
		d.track.DeleteArchivedForever(ctx)
	case strings.HasPrefix(data, "act_toggle_:"):
		if !d.isScreen(ctx.UserID, screenTrackManage) {
			d.closeInlineMenu(ctx, i18n.T(ctx.Language, i18n.KeyTrackManageMenuClosed))
			return
		}
		d.track.HandleTrackToggleCallback(ctx)
	case strings.HasPrefix(data, trackbtn.TrackCBHeatmapDay):
		raw := strings.TrimPrefix(data, trackbtn.TrackCBHeatmapDay)
		day, err := apptime.ParseDay(raw, d.userLocation(ctx.UserID))
		if err != nil {
			return
		}
		d.track.ShowHeatmapDayDetail(ctx, day)
	case data == trackbtn.TrackCBHeatmapBack:
		d.track.ShowHeatmap(ctx, true)
	}
}

// handleLearningCallback routes learning-related inline callbacks.
func (d *Dispatcher) handleLearningCallback(ctx *tgctx.MsgContext, data string) {
	sess := d.sessions.get(ctx.UserID)

	switch {
	case data == learningbtn.LearningCBBackMain:
		d.setScreen(ctx.UserID, screenLearningMain)
		d.learning.ShowLearningMenu(ctx)
	case data == learningbtn.LearningCBCreateCollection:
		sess.waitingLearningCollectionName = true
		d.learning.PromptCreateCollection(ctx)
	case data == learningbtn.LearningCBWordBase:
		d.setScreen(ctx.UserID, screenLearningWordBase)
		d.learning.ShowWordBase(ctx, true)
	case strings.HasPrefix(data, learningbtn.LearningCBCollectionOpen):
		id, ok := parseCallbackID(data, learningbtn.LearningCBCollectionOpen)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenLearningWordBase)
		sess.learningCollectionID = id
		d.learning.ShowCollectionDetail(ctx, id, true)
	case strings.HasPrefix(data, learningbtn.LearningCBCollectionToggle):
		id, ok := parseCallbackID(data, learningbtn.LearningCBCollectionToggle)
		if !ok {
			return
		}
		d.learning.HandleCollectionToggle(ctx, id)
	case strings.HasPrefix(data, learningbtn.LearningCBCollectionAddMore):
		id, ok := parseCallbackID(data, learningbtn.LearningCBCollectionAddMore)
		if !ok {
			return
		}
		sess.waitingLearningWords = true
		sess.learningCollectionID = id
		d.learning.PromptAddWords(ctx, id, false)
	case strings.HasPrefix(data, learningbtn.LearningCBCollectionRename):
		id, ok := parseCallbackID(data, learningbtn.LearningCBCollectionRename)
		if !ok {
			return
		}
		sess.waitingLearningRenameCollection = true
		sess.learningCollectionID = id
		d.learning.PromptRenameCollection(ctx, id)
	case strings.HasPrefix(data, learningbtn.LearningCBCollectionArchive):
		id, ok := parseCallbackID(data, learningbtn.LearningCBCollectionArchive)
		if !ok {
			return
		}
		d.learning.HandleCollectionArchive(ctx, id)
	case strings.HasPrefix(data, learningbtn.LearningCBWordDelete):
		id, ok := parseCallbackID(data, learningbtn.LearningCBWordDelete)
		if !ok {
			return
		}
		d.learning.HandleWordDelete(ctx, id, sess.learningCollectionID)
	case data == learningbtn.LearningCBArchiveOpen:
		d.setScreen(ctx.UserID, screenLearningArchive)
		d.learning.ShowLearningArchiveMenu(ctx, true)
	case strings.HasPrefix(data, learningbtn.LearningCBArchiveRestore):
		id, ok := parseCallbackID(data, learningbtn.LearningCBArchiveRestore)
		if !ok {
			return
		}
		d.learning.RestoreArchivedCollection(ctx, id)
	case strings.HasPrefix(data, learningbtn.LearningCBArchiveDelete):
		id, ok := parseCallbackID(data, learningbtn.LearningCBArchiveDelete)
		if !ok {
			return
		}
		d.learning.DeleteArchivedCollectionForever(ctx, id)
	case data == learningbtn.LearningCBStats:
		d.learning.ShowLearningStatsDetail(ctx, true)
	case data == learningbtn.LearningCBReviewOpen:
		d.setScreen(ctx.UserID, screenLearningReviewPick)
		d.learning.ShowReviewCollectionPicker(ctx, true)
	case strings.HasPrefix(data, learningbtn.LearningCBReviewPickToggle):
		id, ok := parseCallbackID(data, learningbtn.LearningCBReviewPickToggle)
		if !ok {
			return
		}
		d.learning.HandleReviewPickToggle(ctx, id)
	case data == learningbtn.LearningCBReviewContinue:
		if d.learning.HandleReviewContinue(ctx) {
			d.setScreen(ctx.UserID, screenLearningTimer)
		}
	case data == learningbtn.LearningCBReviewStop:
		d.learning.StopReviews(ctx)
	case strings.HasPrefix(data, learningbtn.LearningCBReviewReveal):
		id, ok := parseCallbackID(data, learningbtn.LearningCBReviewReveal)
		if !ok {
			return
		}
		d.learning.ShowReviewReveal(ctx, id)
	case strings.HasPrefix(data, learningbtn.LearningCBReviewAgain):
		id, ok := parseCallbackID(data, learningbtn.LearningCBReviewAgain)
		if !ok {
			return
		}
		d.learning.RecordReviewGrade(ctx, id, models.LearningGradeAgain)
	case strings.HasPrefix(data, learningbtn.LearningCBReviewHard):
		id, ok := parseCallbackID(data, learningbtn.LearningCBReviewHard)
		if !ok {
			return
		}
		d.learning.RecordReviewGrade(ctx, id, models.LearningGradeHard)
	case strings.HasPrefix(data, learningbtn.LearningCBReviewGood):
		id, ok := parseCallbackID(data, learningbtn.LearningCBReviewGood)
		if !ok {
			return
		}
		d.learning.RecordReviewGrade(ctx, id, models.LearningGradeGood)
	case strings.HasPrefix(data, learningbtn.LearningCBReviewEasy):
		id, ok := parseCallbackID(data, learningbtn.LearningCBReviewEasy)
		if !ok {
			return
		}
		d.learning.RecordReviewGrade(ctx, id, models.LearningGradeEasy)
	}
}

// handleRoadmapCallback routes roadmap-related inline callbacks.
//
// Case order matters: prefixes are matched with HasPrefix, so the longer
// "roadmap:archive:this:" / "roadmap:archive:goal*:" forms must not be
// shadowed by a shorter sibling (see
// internal/buttons/roadmap/constants_menu.go).
func (d *Dispatcher) handleRoadmapCallback(ctx *tgctx.MsgContext, data string) {
	sess := d.sessions.get(ctx.UserID)

	switch {
	case data == roadmapbtn.RoadmapCBBackMain:
		d.setScreen(ctx.UserID, screenRoadmapMain)
		d.roadmap.ShowRoadmapMenu(ctx)

	// Goals.
	case data == roadmapbtn.RoadmapCBGoalsOpen:
		d.setScreen(ctx.UserID, screenRoadmapGoals)
		d.roadmap.ShowRoadmapGoals(ctx, true)
	case data == roadmapbtn.RoadmapCBGoalCreate:
		if d.roadmap.PromptCreateGoal(ctx) {
			sess.waitingRoadmapGoalName = true
		}
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBGoalOpen):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBGoalOpen)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenRoadmapGoalDetail)
		sess.roadmapGoalID = id
		d.roadmap.ShowRoadmapGoalDetail(ctx, id, true)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBGoalRename):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBGoalRename)
		if !ok {
			return
		}
		sess.waitingRoadmapGoalRename = true
		sess.roadmapGoalID = id
		d.roadmap.PromptRenameGoal(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBGoalArchive):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBGoalArchive)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenRoadmapGoals)
		d.roadmap.HandleGoalArchive(ctx, id)

	// Technologies.
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBTechCreate):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBTechCreate)
		if !ok {
			return
		}
		sess.roadmapGoalID = id
		if d.roadmap.PromptCreateRoadmap(ctx, id) {
			sess.waitingRoadmapName = true
		}
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBOpen):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBOpen)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenRoadmapCards)
		sess.roadmapID = id
		d.roadmap.ShowRoadmapDetail(ctx, id, true)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBToggle):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBToggle)
		if !ok {
			return
		}
		d.roadmap.HandleRoadmapToggle(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBAddCards):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBAddCards)
		if !ok {
			return
		}
		sess.waitingRoadmapCards = true
		sess.roadmapID = id
		d.roadmap.PromptAddRoadmapCards(ctx, id, false)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBSetCriteria):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBSetCriteria)
		if !ok {
			return
		}
		sess.waitingRoadmapCriteria = true
		sess.roadmapID = id
		d.roadmap.PromptRoadmapCriteria(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBRename):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBRename)
		if !ok {
			return
		}
		sess.waitingRoadmapRename = true
		sess.roadmapID = id
		d.roadmap.PromptRenameRoadmap(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBArchiveThis):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBArchiveThis)
		if !ok {
			return
		}
		d.roadmap.HandleRoadmapArchive(ctx, id)
	case data == roadmapbtn.RoadmapCBOrphansOpen:
		d.setScreen(ctx.UserID, screenRoadmapOrphans)
		d.roadmap.ShowRoadmapOrphans(ctx, true)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBAssignOpen):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBAssignOpen)
		if !ok {
			return
		}
		sess.roadmapID = id
		d.roadmap.ShowRoadmapAssignGoal(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBAssignPick):
		roadmapID, goalID, ok := roadmapbtn.ParseAssignPayload(data)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenRoadmapGoalDetail)
		sess.roadmapGoalID = goalID
		d.roadmap.HandleRoadmapAssign(ctx, roadmapID, goalID)

	// Archive.
	case data == roadmapbtn.RoadmapCBArchiveOpen:
		d.setScreen(ctx.UserID, screenRoadmapArchive)
		d.roadmap.ShowRoadmapArchiveMenu(ctx, true)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBArchiveGoalRestore):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBArchiveGoalRestore)
		if !ok {
			return
		}
		d.roadmap.RestoreArchivedGoal(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBArchiveGoalDelete):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBArchiveGoalDelete)
		if !ok {
			return
		}
		d.roadmap.DeleteArchivedGoalForever(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBArchiveRestore):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBArchiveRestore)
		if !ok {
			return
		}
		d.roadmap.RestoreArchivedRoadmap(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBArchiveDelete):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBArchiveDelete)
		if !ok {
			return
		}
		d.roadmap.DeleteArchivedRoadmapForever(ctx, id)

	// Cards. Standalone on purpose — no screen requirement, since the card id
	// resolves its own technology.
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBCardToggle):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBCardToggle)
		if !ok {
			return
		}
		d.roadmap.HandleRoadmapCardToggle(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBCardDiff):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBCardDiff)
		if !ok {
			return
		}
		d.roadmap.HandleRoadmapCardDifficulty(ctx, id)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBCardDelete):
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBCardDelete)
		if !ok {
			return
		}
		d.roadmap.HandleRoadmapCardDelete(ctx, id, sess.roadmapID)
	case strings.HasPrefix(data, roadmapbtn.RoadmapCBDigestToggle):
		// Same flip, but tapped on a reminder push — re-renders that push
		// instead of a checklist screen.
		id, ok := parseCallbackID(data, roadmapbtn.RoadmapCBDigestToggle)
		if !ok {
			return
		}
		d.roadmap.HandleRoadmapDigestToggle(ctx, id)

	case data == roadmapbtn.RoadmapCBStats:
		d.roadmap.ShowRoadmapStatsDetail(ctx, true)
	case data == roadmapbtn.RoadmapCBPushOpen:
		if d.roadmap.ShowRoadmapIntervalPicker(ctx) {
			d.setScreen(ctx.UserID, screenRoadmapTimer)
		}
	case data == roadmapbtn.RoadmapCBPushStop:
		d.roadmap.StopRoadmapReminders(ctx)
	}
}

// cancelRoadmapPrompt drops the waiting reply keyboard and says so — shared
// by every roadmap text prompt's cancel path.
func (d *Dispatcher) cancelRoadmapPrompt(ctx *tgctx.MsgContext) {
	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonCancelled))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = d.bot.Send(hide)
}

// handleChallengeCallback routes challenge-related inline callbacks.
// handleOnboardingCallback routes the "here's what you can do" tour's
// navigation. Stateless — the step lives entirely in the callback data, no
// session field needed.
func (d *Dispatcher) handleOnboardingCallback(ctx *tgctx.MsgContext, data string) {
	switch {
	case strings.HasPrefix(data, onboardingbtn.CBGoto):
		step, ok := parseCallbackID(data, onboardingbtn.CBGoto)
		if !ok {
			return
		}
		d.entry.ShowOnboardingStep(ctx, int(step), true)
	case data == onboardingbtn.CBSkip:
		d.setScreen(ctx.UserID, screenHome)
		d.entry.ShowHomeMenu(ctx)
	}
}

func (d *Dispatcher) handleChallengeCallback(ctx *tgctx.MsgContext, data string) {
	sess := d.sessions.get(ctx.UserID)

	switch {
	case data == challengebtn.CBBackList:
		d.setScreen(ctx.UserID, screenChallengeList)
		d.entry.ShowChallengesMenu(ctx)
	case data == challengebtn.CBCreate:
		sess.waitingChallengeName = true
		d.setScreen(ctx.UserID, screenChallengeCreate)
		d.entry.PromptCreateChallenge(ctx)
	case strings.HasPrefix(data, challengebtn.CBOpen):
		id, ok := parseCallbackID(data, challengebtn.CBOpen)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenChallengeGrid)
		sess.challengeID = id
		d.entry.ShowChallengeGrid(ctx, id, true)
	case strings.HasPrefix(data, challengebtn.CBArchiveThis):
		id, ok := parseCallbackID(data, challengebtn.CBArchiveThis)
		if !ok {
			return
		}
		d.setScreen(ctx.UserID, screenChallengeList)
		d.entry.ArchiveChallenge(ctx, id)
	case data == challengebtn.CBArchiveOpen:
		d.setScreen(ctx.UserID, screenChallengeArchive)
		d.entry.ShowChallengeArchive(ctx, true)
	case strings.HasPrefix(data, challengebtn.CBArchiveRestore):
		id, ok := parseCallbackID(data, challengebtn.CBArchiveRestore)
		if !ok {
			return
		}
		d.entry.RestoreChallenge(ctx, id)
	case strings.HasPrefix(data, challengebtn.CBArchiveDelete):
		id, ok := parseCallbackID(data, challengebtn.CBArchiveDelete)
		if !ok {
			return
		}
		d.entry.DeleteChallengeForever(ctx, id)
	case strings.HasPrefix(data, challengebtn.CBDayOpen):
		id, day, ok := parseChallengeDayPayload(data, challengebtn.CBDayOpen, time.UTC)
		if !ok {
			return
		}
		d.entry.ShowChallengeDayConfirm(ctx, id, day)
	case strings.HasPrefix(data, challengebtn.CBDayDone):
		id, day, ok := parseChallengeDayPayload(data, challengebtn.CBDayDone, time.UTC)
		if !ok {
			return
		}
		d.entry.MarkChallengeDay(ctx, id, day, true)
	case strings.HasPrefix(data, challengebtn.CBDaySkip):
		id, day, ok := parseChallengeDayPayload(data, challengebtn.CBDaySkip, time.UTC)
		if !ok {
			return
		}
		d.entry.MarkChallengeDay(ctx, id, day, false)
	case data == challengebtn.CBCalPrev:
		sess.challengeCalMonth = sess.challengeCalMonth.AddDate(0, -1, 0)
		d.entry.ShowCreateChallengeCalendarInPlace(ctx, sess.challengeCalMonth, sess.challengeCalFrom, sess.challengeCalTo)
	case data == challengebtn.CBCalNext:
		sess.challengeCalMonth = sess.challengeCalMonth.AddDate(0, 1, 0)
		d.entry.ShowCreateChallengeCalendarInPlace(ctx, sess.challengeCalMonth, sess.challengeCalFrom, sess.challengeCalTo)
	case data == challengebtn.CBCalPrevYear:
		sess.challengeCalMonth = sess.challengeCalMonth.AddDate(-1, 0, 0)
		d.entry.ShowCreateChallengeCalendarInPlace(ctx, sess.challengeCalMonth, sess.challengeCalFrom, sess.challengeCalTo)
	case data == challengebtn.CBCalNextYear:
		sess.challengeCalMonth = sess.challengeCalMonth.AddDate(1, 0, 0)
		d.entry.ShowCreateChallengeCalendarInPlace(ctx, sess.challengeCalMonth, sess.challengeCalFrom, sess.challengeCalTo)
	case strings.HasPrefix(data, challengebtn.CBCalPick):
		raw := strings.TrimPrefix(data, challengebtn.CBCalPick)
		day, err := apptime.ParseDay(raw, time.UTC)
		if err != nil {
			return
		}
		if sess.challengeCalFrom.IsZero() || !sess.challengeCalTo.IsZero() {
			sess.challengeCalFrom = day
			sess.challengeCalTo = time.Time{}
		} else {
			sess.challengeCalTo = day
			if sess.challengeCalTo.Before(sess.challengeCalFrom) {
				sess.challengeCalFrom, sess.challengeCalTo = sess.challengeCalTo, sess.challengeCalFrom
			}
		}
		d.entry.ShowCreateChallengeCalendarInPlace(ctx, sess.challengeCalMonth, sess.challengeCalFrom, sess.challengeCalTo)
	case data == challengebtn.CBCalDone:
		if sess.challengeCalFrom.IsZero() || sess.challengeCalTo.IsZero() {
			return
		}
		name := sess.pendingChallengeName
		from, to := sess.challengeCalFrom, sess.challengeCalTo
		sess.pendingChallengeName = ""
		sess.challengeCalFrom, sess.challengeCalTo, sess.challengeCalMonth = time.Time{}, time.Time{}, time.Time{}
		d.setScreen(ctx.UserID, screenChallengeList)
		d.entry.CreateChallenge(ctx, name, from, to)
	case data == challengebtn.CBCalCancel:
		sess.pendingChallengeName = ""
		sess.challengeCalFrom, sess.challengeCalTo, sess.challengeCalMonth = time.Time{}, time.Time{}, time.Time{}
		d.setScreen(ctx.UserID, screenChallengeList)
		d.entry.ShowChallengesMenu(ctx)
	case strings.HasPrefix(data, challengebtn.CBPushDone):
		id, ok := parseCallbackID(data, challengebtn.CBPushDone)
		if !ok {
			return
		}
		d.entry.RecordChallengePushAnswer(ctx, id, true)
	case strings.HasPrefix(data, challengebtn.CBPushSkip):
		id, ok := parseCallbackID(data, challengebtn.CBPushSkip)
		if !ok {
			return
		}
		d.entry.RecordChallengePushAnswer(ctx, id, false)
	}
}

// parseChallengeDayPayload splits a "<prefix><challengeID>:<2006-01-02>"
// callback payload into its parts.
func parseChallengeDayPayload(data, prefix string, loc *time.Location) (challengeID int64, day time.Time, ok bool) {
	raw := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return 0, time.Time{}, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, time.Time{}, false
	}
	day, err = apptime.ParseDay(parts[1], loc)
	if err != nil {
		return 0, time.Time{}, false
	}
	return id, day, true
}

// replyUseButtons sends a guard message when user is out of current flow.
func (d *Dispatcher) replyUseButtons(ctx *tgctx.MsgContext) {
	_, _ = d.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonUseButtons)))
}

// isTrackButtonText checks if text belongs to a known nav/action button, so
// a "waiting for free text" flow (activity name, custom timer minutes, ...)
// can tell a stray button tap from actual input.
func (d *Dispatcher) isTrackButtonText(ctx *tgctx.MsgContext) bool {
	if ctx.Text == adminbtn.ReplyButtonUsers {
		return true
	}
	// Covers every translated nav/action button (Back/Home, Activate/
	// Delete, Timer Create/Delete, View Archive, Select Activity, Admin,
	// Today, Calendar) — see handleTranslatedButton for the full set this
	// key space spans.
	if _, ok := i18n.Key(ctx.Language, ctx.Text); ok {
		return true
	}
	if _, ok := trackbtn.ParseTimerButtonMinutes(ctx.Language, ctx.Text, trackbtn.TrackTimerActivatePrefix); ok {
		return true
	}
	if _, ok := trackbtn.ParseTimerButtonMinutes(ctx.Language, ctx.Text, trackbtn.TrackTimerDeletePrefix); ok {
		return true
	}
	return false
}

// ensurePeriodDefaults sets initial period report dates for user.
func (d *Dispatcher) ensurePeriodDefaults(userID int64) {
	loc := d.userLocation(userID)
	sess := d.sessions.get(userID)
	if sess.reportFrom.IsZero() {
		sess.reportFrom = apptime.NowIn(loc).AddDate(0, 0, -30)
	}
	if sess.reportTo.IsZero() {
		sess.reportTo = apptime.NowIn(loc)
	}
	d.getReportSelected(userID)
}

// setScreen stores current UI screen for user, in memory and persisted so it
// survives a bot restart.
func (d *Dispatcher) setScreen(userID int64, screen string) {
	sess := d.sessions.get(userID)
	sess.screen = screen
	sess.screenLoaded = true
	if d.uistate == nil || sess.dbID == 0 {
		return
	}
	if err := d.uistate.SetScreen(d.appCtx, sess.dbID, screen); err != nil {
		log.Error().Err(err).Int64("user_id", sess.dbID).Msg("persist screen failed")
	}
}

// isScreen checks whether current screen is one of allowed values.
func (d *Dispatcher) isScreen(userID int64, allowed ...string) bool {
	current := d.sessions.get(userID).screen
	for _, s := range allowed {
		if current == s {
			return true
		}
	}
	return false
}

// calendarMonth returns current calendar month or now if empty.
func (d *Dispatcher) calendarMonth(userID int64) time.Time {
	m := d.sessions.get(userID).reportCalMonth
	if m.IsZero() {
		return apptime.NowIn(d.userLocation(userID))
	}
	return m
}

// showPeriodMenu redraws period report menu.
func (d *Dispatcher) showPeriodMenu(ctx *tgctx.MsgContext) {
	sess := d.sessions.get(ctx.UserID)
	d.track.ShowPeriodMenu(ctx, d.getReportSelected(ctx.UserID), sess.reportCalMonth, sess.reportFrom, sess.reportTo)
}

// showPeriodCalendar redraws period calendar view.
func (d *Dispatcher) showPeriodCalendar(ctx *tgctx.MsgContext) {
	sess := d.sessions.get(ctx.UserID)
	d.track.ShowPeriodCalendar(ctx, sess.reportCalMonth, sess.reportCalFrom, sess.reportCalTo)
}

// getReportSelected returns selected activity map for user.
func (d *Dispatcher) getReportSelected(userID int64) map[int64]bool {
	sess := d.sessions.get(userID)
	if sess.reportSelected == nil {
		sess.reportSelected = make(map[int64]bool)
	}
	return sess.reportSelected
}

// parseDateRange parses "YYYY-MM-DD..YYYY-MM-DD".
func parseDateRange(s string) (time.Time, time.Time, error) {
	parts := strings.Split(strings.TrimSpace(s), "..")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("bad format")
	}
	from, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("from after to")
	}
	return from, to, nil
}

// selectedIDs converts selected map to slice of ids.
func selectedIDs(m map[int64]bool) []int64 {
	out := make([]int64, 0, len(m))
	for id, ok := range m {
		if ok {
			out = append(out, id)
		}
	}
	return out
}

// toggleSelected toggles selected state for one id.
func toggleSelected(m map[int64]bool, id int64) {
	m[id] = !m[id]
	if !m[id] {
		delete(m, id)
	}
}

// parseCallbackID extracts int64 id from callback by prefix.
func parseCallbackID(data, prefix string) (int64, bool) {
	idRaw := strings.TrimPrefix(data, prefix)
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// sameDay checks whether two dates are the same day.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// closeInlineMenu clears inline keyboard and replaces message text.
func (d *Dispatcher) closeInlineMenu(ctx *tgctx.MsgContext, text string) {
	if ctx.MessageID <= 0 {
		return
	}
	empty := tgbotapi.NewInlineKeyboardMarkup()
	edit := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, empty)
	_, _ = d.bot.Send(edit)
}
