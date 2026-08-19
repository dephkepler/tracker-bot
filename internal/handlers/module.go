package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tracker-bot/internal/buttons/admin"
	"tracker-bot/internal/buttons/entry"
	"tracker-bot/internal/buttons/learning"
	"tracker-bot/internal/buttons/profile"
	"tracker-bot/internal/buttons/subscription"
	"tracker-bot/internal/buttons/track"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/internal/service"
	"tracker-bot/internal/utils/tgctx"
	"tracker-bot/pkg/apptime"
	"tracker-bot/pkg/geotz"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

// Module routes UI actions to services and renders bot responses.

type Handler interface {
	Track()
}

type Module struct {
	bot             *tgbotapi.BotAPI
	profilesvc      service.ProfileService
	tracksvc        service.TrackerService
	timersvc        service.TimerService
	learningsvc     service.LearningService
	subscriptionsvc service.SubscriptionService
	entrysvc        service.EntryService
	// adminUsername is the Telegram @handle (no leading "@") allowed to see
	// the admin screen — see IsAdmin. Empty disables the admin feature
	// entirely rather than matching everyone.
	adminUsername string
}

// adminUsersPageSize is how many users are listed per page on the admin
// screen.
const adminUsersPageSize = 15

// New creates handler module with all service dependencies.
func New(bot *tgbotapi.BotAPI, entrysvc service.EntryService, profilesvc service.ProfileService, tracksvc service.TrackerService, timersvc service.TimerService, learningsvc service.LearningService, subscriptionsvc service.SubscriptionService, adminUsername string) *Module {
	return &Module{
		bot:             bot,
		profilesvc:      profilesvc,
		tracksvc:        tracksvc,
		timersvc:        timersvc,
		learningsvc:     learningsvc,
		subscriptionsvc: subscriptionsvc,
		entrysvc:        entrysvc,
		adminUsername:   strings.TrimPrefix(strings.TrimSpace(adminUsername), "@"),
	}
}

// IsAdmin reports whether ctx belongs to the configured admin user, matched
// by Telegram @handle (case-insensitive). False whenever adminUsername is
// unset or the user has no @handle — this is the actual access check, not
// just whether the admin button is shown, so callers must call it directly
// rather than trusting screen/button visibility.
func (m *Module) IsAdmin(ctx *tgctx.MsgContext) bool {
	if m.adminUsername == "" || ctx.Username == "" {
		return false
	}
	return strings.EqualFold(ctx.Username, m.adminUsername)
}

// ShowEntryMenu renders the main entry screen: a one-time welcome for a
// brand-new user, a plain "home" message for anyone returning to it.
func (m *Module) ShowEntryMenu(ctx *tgctx.MsgContext) {
	if ctx.IsNewUser {
		m.ShowWelcome(ctx)
		return
	}
	m.ShowHomeMenu(ctx)
}

// ShowWelcome renders the first-time greeting (only meant for a user's very
// first /start).
func (m *Module) ShowWelcome(ctx *tgctx.MsgContext) {
	m.sendEntryMenu(ctx, entry.WelcomeText(ctx.Language))
}

// ShowHomeMenu renders the entry screen for a user who already knows the
// bot — every other "go home" action (Home button, /start after the first
// time, post-activation redirect, etc.) should use this, not ShowWelcome.
func (m *Module) ShowHomeMenu(ctx *tgctx.MsgContext) {
	m.sendEntryMenu(ctx, entry.HomeMenuText(ctx.Language))
}

func (m *Module) sendEntryMenu(ctx *tgctx.MsgContext, text string) {
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = entry.EntryReplyMenu(ctx.Language, m.IsAdmin(ctx))

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send entry menu failed")
	}
}

// ShowAdminMenu renders the admin "who's using the bot" screen: a total
// count plus one page of users (see admin.UsersInlineMenu for paging).
// Silently does nothing for a non-admin caller — this is the actual access
// check (see IsAdmin), the button/command being hidden from other users is
// only a UI nicety on top of it.
func (m *Module) ShowAdminMenu(ctx *tgctx.MsgContext) {
	if !m.IsAdmin(ctx) {
		return
	}

	total, err := m.entrysvc.CountUsers(ctx.Ctx)
	if err != nil {
		log.Error().Err(err).Msg("count users failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to load users."))
		return
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, fmt.Sprintf("👑 Admin\n\nTotal users: %d", total))
	msg.ReplyMarkup = admin.MenuReplyMenu()
	_, _ = m.bot.Send(msg)
}

// ShowAdminUsersMenu renders the paginated users list as a fresh message:
// a reply-keyboard carrier (Back/Home) plus the inline listing, mirroring
// the dual-message pattern used elsewhere (e.g. ShowTrackActivitySelectionMenu).
// Silently does nothing for a non-admin caller.
func (m *Module) ShowAdminUsersMenu(ctx *tgctx.MsgContext, offset int) {
	if !m.IsAdmin(ctx) {
		return
	}
	reply := tgbotapi.NewMessage(ctx.ChatID, "👥")
	reply.ReplyMarkup = admin.UsersReplyMenu()
	_, _ = m.bot.Send(reply)

	m.sendOrEditAdminUsers(ctx, offset, false)
}

// ShowAdminUsersMenuInPlace re-renders the users list by editing the
// existing inline message (used for Prev/Next paging, so clicking through
// pages doesn't spam a new reply-keyboard message each time). Silently does
// nothing for a non-admin caller.
func (m *Module) ShowAdminUsersMenuInPlace(ctx *tgctx.MsgContext, offset int) {
	if !m.IsAdmin(ctx) {
		return
	}
	m.sendOrEditAdminUsers(ctx, offset, true)
}

func (m *Module) sendOrEditAdminUsers(ctx *tgctx.MsgContext, offset int, inPlace bool) {
	total, err := m.entrysvc.CountUsers(ctx.Ctx)
	if err != nil {
		log.Error().Err(err).Msg("count users failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to load users."))
		return
	}

	if offset < 0 {
		offset = 0
	}
	if total > 0 && offset >= total {
		offset = ((total - 1) / adminUsersPageSize) * adminUsersPageSize
	}

	users, err := m.entrysvc.ListUsersPage(ctx.Ctx, adminUsersPageSize, offset)
	if err != nil {
		log.Error().Err(err).Msg("list users failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to load users."))
		return
	}

	text := fmt.Sprintf("👥 Users\n\nTotal: %d", total)
	markup := admin.UsersInlineMenu(users, offset, adminUsersPageSize, total)

	if inPlace && ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, markup)
		if _, err := m.bot.Send(edit); err != nil {
			log.Error().Err(err).Msg("edit admin users list failed")
		}
		return
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ReplyMarkup = markup
	_, _ = m.bot.Send(msg)
}

// ShowProfileMenu loads profile stats and renders profile screen.
func (m *Module) ShowProfileMenu(ctx *tgctx.MsgContext) {
	stats, err := m.profilesvc.GetProfileStats(ctx.Ctx, ctx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GetProfile failed")
		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLoadFailed))
		_, _ = m.bot.Send(msg)
		return
	}

	text := profile.ProfileMenuText(ctx.Language, stats)

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ReplyMarkup = profile.ProfileEntryInlineMenu(ctx.Language, m.IsAdmin(ctx))

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send profile menu failed")
	}
}

// languageCodeByButton maps a language-picker reply button's text to the
// ISO code stored in users.language (see migration 0001_users_init.up.sql's
// users_allowed_language CHECK constraint: ru/en/de/uk/ar).
var languageCodeByButton = map[string]string{
	profile.ProfileButtonLanguageRussian:   "ru",
	profile.ProfileButtonLanguageEnglish:   "en",
	profile.ProfileButtonLanguageGerman:    "de",
	profile.ProfileButtonLanguageUkrainian: "uk",
	profile.ProfileButtonLanguageArabian:   "ar",
}

// ShowLanguagePicker asks the user to pick their interface language.
func (m *Module) ShowLanguagePicker(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLanguagePrompt))
	msg.ReplyMarkup = profile.ProfileLanguageManageReplyMenu(ctx.Language)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send language picker failed")
	}
}

// ProcessLanguageSelection saves the language behind the reply button the
// user tapped (see languageCodeByButton) and returns to the profile screen.
// Returns false — without changing anything — if ctx.Text isn't one of the
// picker's buttons, so the caller keeps waiting for a valid tap instead of
// silently dropping the selection.
func (m *Module) ProcessLanguageSelection(ctx *tgctx.MsgContext) bool {
	code, ok := languageCodeByButton[ctx.Text]
	if !ok {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLanguageInvalid)))
		return false
	}

	if err := m.profilesvc.ChangeLanguage(ctx.Ctx, ctx.UserID, code); err != nil {
		log.Error().Err(err).Msg("save language failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLanguageSaveFailed)))
		return false
	}

	// Switch ctx over to the newly-picked language right away — both for the
	// confirmation below (so it reads in the language just chosen, not the
	// old one) and for the ShowProfileMenu call after it, so the profile
	// screen already renders in the new language on this same round trip.
	// The dispatcher independently invalidates the cached session language
	// (sess.langLoaded = false) so the NEXT message also picks it up fresh
	// from DB rather than relying on this in-request mutation.
	ctx.Language = i18n.Normalize(code)

	// ctx.Text is the language button itself (e.g. "🇷🇺 Русский") — those are
	// deliberately not translated (see profile.ProfileLanguageManageReplyMenu),
	// so it's already the right thing to substitute in regardless of language.
	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLanguageSaved, ctx.Text))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)

	m.ShowProfileMenu(ctx)
	return true
}

// ShowLocationRequest asks the user to share their location so the bot can
// detect their real timezone instead of assuming one for everyone.
func (m *Module) ShowLocationRequest(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileTimezonePrompt))
	msg.ReplyMarkup = profile.ProfileLocationReplyMenu(ctx.Language)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send location request failed")
	}
}

// ProcessLocationTimeZone resolves the shared location to an IANA timezone,
// saves it, and confirms to the user.
func (m *Module) ProcessLocationTimeZone(ctx *tgctx.MsgContext, lat, lng float64) {
	tzName, err := geotz.Lookup(lat, lng)
	if err != nil {
		log.Error().Err(err).Msg("resolve timezone from location failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileTimezoneLookupFailed)))
		return
	}

	if err := m.profilesvc.ChangeTimeZone(ctx.Ctx, ctx.UserID, tzName); err != nil {
		log.Error().Err(err).Msg("save timezone failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileTimezoneSaveFailed)))
		return
	}

	// tzName is an IANA identifier (e.g. "Europe/Berlin") — a technical
	// value, not translatable text, so it's inserted as-is regardless of
	// language.
	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileTimezoneSaved, tzName))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)

	m.ShowProfileMenu(ctx)
}

// ShowTrackingMenu loads tracking stats and renders tracking home screen.
func (m *Module) ShowTrackingMenu(ctx *tgctx.MsgContext) {
	stats, err := m.tracksvc.GetMainStats(ctx.Ctx, ctx.DBUserID, ctx.Location)
	if err != nil {
		log.Error().Err(err).Msg("GetMainStats failed")
		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackLoadFailed))
		_, _ = m.bot.Send(msg)
		return
	}

	text := track.TrackingMenuText(ctx.Language, stats)

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = track.TrackEntryInlineMenu(ctx.Language)

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send tracking menu failed")
	}
}

// ShowReportsHub renders report type selector.
func (m *Module) ShowReportsHub(ctx *tgctx.MsgContext, inPlace bool) {
	text := i18n.T(ctx.Language, i18n.KeyTrackReportsHubTitle)
	msgReply := tgbotapi.NewMessage(ctx.ChatID, "📈")
	msgReply.ReplyMarkup = track.TrackReportsReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msgReply)

	if inPlace && ctx.MessageID > 0 {
		msg := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
		_, _ = m.bot.Send(msg)
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

// ShowTodayChart renders today's activity distribution as text bars.
func (m *Module) ShowTodayChart(ctx *tgctx.MsgContext) {
	stats, err := m.tracksvc.GetTodayReport(ctx.Ctx, ctx.DBUserID, ctx.Location)
	if err != nil {
		log.Error().Err(err).Msg("today chart failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTodayChartLoadFailed)))
		return
	}
	if len(stats.TopActivities) == 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTodayChartEmpty)))
		return
	}

	maxDur := time.Duration(1)
	for _, a := range stats.TopActivities {
		if a.Duration > maxDur {
			maxDur = a.Duration
		}
	}

	var b strings.Builder
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackTodayChartTitle))
	total := stats.TotalTracked
	for _, a := range stats.TopActivities {
		name := a.Name
		if a.Emoji != "" {
			name = a.Emoji + " " + a.Name
		}
		barLen := int((float64(a.Duration) / float64(maxDur)) * 12.0)
		if barLen < 1 {
			barLen = 1
		}
		if barLen > 12 {
			barLen = 12
		}
		percent := percentOf(a.Duration, total)
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackTodayChartActivityLine, name, strings.Repeat("█", barLen), formatReportDuration(a.Duration), percent))
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, b.String())
	msg.ReplyMarkup = track.TrackReportTodayInlineMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

// ShowPeriodMenu renders period report configuration screen.
func (m *Module) ShowPeriodMenu(ctx *tgctx.MsgContext, selected map[int64]bool, month, from, to time.Time) {
	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPeriodMenuLoadFailed)))
		return
	}
	if month.IsZero() {
		month = apptime.NowIn(ctx.Location)
	}
	rangeLabel := formatDateOrDash(from) + ".." + formatDateOrDash(to)
	text := i18n.T(ctx.Language, i18n.KeyTrackPeriodMenuTitle, len(selected), rangeLabel)
	if ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			ctx.ChatID,
			ctx.MessageID,
			text,
			track.TrackReportPeriodInlineMenu(ctx.Language, items, selected, rangeLabel),
		)
		_, _ = m.bot.Send(edit)
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ReplyMarkup = track.TrackReportPeriodInlineMenu(ctx.Language, items, selected, rangeLabel)
	_, _ = m.bot.Send(msg)
}

// ShowPeriodTextReport builds and sends period report in text form.
func (m *Module) ShowPeriodTextReport(ctx *tgctx.MsgContext, from, to time.Time, activityIDs []int64, selectedOnly bool) {
	stats, err := m.tracksvc.GetPeriodReport(ctx.Ctx, ctx.DBUserID, from, to.Add(24*time.Hour), activityIDs, ctx.Location)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPeriodTextFailed)))
		return
	}
	var b strings.Builder
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodTextTitle))
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodRangeLine, from.Format("2006-01-02"), to.Format("2006-01-02")))
	if selectedOnly {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodScopeSelected))
	} else {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodScopeAll))
	}
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodTotalsLine, formatReportDuration(stats.TotalTracked), stats.TotalSessions))
	total := stats.TotalTracked
	if len(stats.Activities) == 0 {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodNoSessions))
	} else {
		for i, a := range stats.Activities {
			name := a.Name
			if a.Emoji != "" {
				name = a.Emoji + " " + a.Name
			}
			b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodTextActivityLine, i+1, name, formatReportDuration(a.Duration), percentOf(a.Duration, total), a.Sessions))
		}
	}
	m.appendGranularityText(ctx, &b, from, to, activityIDs)
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, b.String()))
}

// ShowPeriodChartReport builds and sends period report in chart-like form.
func (m *Module) ShowPeriodChartReport(ctx *tgctx.MsgContext, from, to time.Time, activityIDs []int64) {
	stats, err := m.tracksvc.GetPeriodReport(ctx.Ctx, ctx.DBUserID, from, to.Add(24*time.Hour), activityIDs, ctx.Location)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPeriodChartFailed)))
		return
	}
	if len(stats.Activities) == 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPeriodChartEmpty)))
		return
	}
	maxDur := time.Duration(1)
	for _, a := range stats.Activities {
		if a.Duration > maxDur {
			maxDur = a.Duration
		}
	}
	var b strings.Builder
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodChartTitle))
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodRangeLine, from.Format("2006-01-02"), to.Format("2006-01-02")))
	b.WriteString("\n")
	total := stats.TotalTracked
	for _, a := range stats.Activities {
		name := a.Name
		if a.Emoji != "" {
			name = a.Emoji + " " + a.Name
		}
		barLen := int((float64(a.Duration) / float64(maxDur)) * 12)
		if barLen < 1 {
			barLen = 1
		}
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackPeriodChartActivityLine, name, strings.Repeat("█", barLen), formatReportDuration(a.Duration), percentOf(a.Duration, total), a.Sessions))
	}
	m.appendGranularityText(ctx, &b, from, to, activityIDs)
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, b.String()))
}

// ShowPeriodCalendar renders inline calendar for period selection.
func (m *Module) ShowPeriodCalendar(ctx *tgctx.MsgContext, month, from, to time.Time) {
	text := i18n.T(ctx.Language, i18n.KeyTrackCalendarPickTitle, formatDateOrDash(from), formatDateOrDash(to))
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID,
		ctx.MessageID,
		text,
		track.TrackReportPeriodCalendarInlineMenu(ctx.Language, month, from, to),
	)
	_, _ = m.bot.Send(edit)
}

// appendGranularityText appends bucketed totals (hour/day/month) to report.
func (m *Module) appendGranularityText(ctx *tgctx.MsgContext, b *strings.Builder, from, to time.Time, activityIDs []int64) {
	if len(activityIDs) == 0 {
		return
	}

	granularity := "day"
	labelFmt := "2006-01-02"
	if from.Year() != to.Year() {
		granularity = "month"
		labelFmt = "2006-01"
	} else if from.Year() == to.Year() && from.Month() == to.Month() && from.Day() == to.Day() {
		granularity = "hour"
		labelFmt = "15:00"
	}

	buckets, durs, err := m.tracksvc.GetPeriodBuckets(ctx.Ctx, ctx.DBUserID, from, to.Add(24*time.Hour), activityIDs, granularity, ctx.Location)
	if err != nil || len(buckets) == 0 {
		return
	}

	switch granularity {
	case "month":
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityByMonths))
	case "day":
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityByDays))
	case "hour":
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityByHours))
	}

	for i := range buckets {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityBucketLine, buckets[i].Format(labelFmt), formatReportDuration(durs[i])))
	}
}

// ShowTodayReport is an alias to chart-style today report.
func (m *Module) ShowTodayReport(ctx *tgctx.MsgContext) {
	m.ShowTodayChart(ctx)
}

// ShowTodayReportBySelected opens selected-activities screen for today report.
func (m *Module) ShowTodayReportBySelected(ctx *tgctx.MsgContext) {
	m.ShowTodaySelectActivities(ctx, map[int64]bool{})
}

// ShowTodaySelectActivities renders multi-select activities for today chart.
func (m *Module) ShowTodaySelectActivities(ctx *tgctx.MsgContext, selected map[int64]bool) {
	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageLoadFailed)))
		return
	}
	text := i18n.T(ctx.Language, i18n.KeyTrackTodaySelectTitle)
	if ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			ctx.ChatID,
			ctx.MessageID,
			text,
			track.TrackTodaySelectActivitiesInlineMenu(ctx.Language, items, selected),
		)
		_, _ = m.bot.Send(edit)
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ReplyMarkup = track.TrackTodaySelectActivitiesInlineMenu(ctx.Language, items, selected)
	_, _ = m.bot.Send(msg)
}

// PromptCreateActivity asks user to type a new activity name.
func (m *Module) PromptCreateActivity(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCreatePrompt))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = track.TrackActivityManageReplyMenu(ctx.Language)

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send create activity prompt failed")
	}
}

// ProcessCreateActivity validates and creates activity from plain text input.
func (m *Module) ProcessCreateActivity(ctx *tgctx.MsgContext) bool {
	name := strings.TrimSpace(ctx.Text)
	if name == "" {
		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCreateEmptyName))
		_, _ = m.bot.Send(msg)
		return false
	}

	activity, err := m.tracksvc.CreateActivity(ctx.Ctx, ctx.DBUserID, name, "")
	if err != nil {
		if err == models.ErrActivityExists {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCreateAlreadyExists)))
			return false
		}
		log.Error().Err(err).Msg("create activity failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCreateFailed)))
		return false
	}

	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCreateConfirmed, activity.Name))
	confirm.ReplyMarkup = track.TrackCreateSuccessInlineMenu(ctx.Language)
	_, _ = m.bot.Send(confirm)
	return true
}

// ShowTrackActivitySelectionMenu renders active activities and selection state.
func (m *Module) ShowTrackActivitySelectionMenu(ctx *tgctx.MsgContext) {
	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageLoadFailed)))
		return
	}

	if len(items) == 0 {
		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageEmpty))
		msg.ReplyMarkup = track.TrackActivityManageReplyMenu(ctx.Language)
		_, _ = m.bot.Send(msg)
		return
	}

	selectedCount := countSelectedActivities(items)

	msgReply := tgbotapi.NewMessage(ctx.ChatID, "🗂")
	msgReply.ReplyMarkup = track.TrackActivityManageReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msgReply)

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageSelectTitle, selectedCount, len(items)))
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = track.TrackActivitiesInlineMenu(ctx.Language, items)
	_, _ = m.bot.Send(msg)
}

// HandleTrackToggleCallback toggles one activity in selected set.
func (m *Module) HandleTrackToggleCallback(ctx *tgctx.MsgContext) {
	payload := strings.TrimPrefix(ctx.Text, "act_toggle_:")
	activityID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackInvalidActivityID)))
		return
	}

	if err := m.tracksvc.ToggleSelectedActivity(ctx.Ctx, ctx.DBUserID, activityID); err != nil {
		log.Error().Err(err).Msg("toggle activity failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageToggleFailed)))
		return
	}

	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("reload activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageRefreshFailed)))
		return
	}

	selectedCount := countSelectedActivities(items)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID,
		ctx.MessageID,
		i18n.T(ctx.Language, i18n.KeyTrackManageSelectTitle, selectedCount, len(items)),
		track.TrackActivitiesInlineMenu(ctx.Language, items),
	)
	edit.ParseMode = "HTML"
	if _, err := m.bot.Send(edit); err != nil {
		log.Error().Err(err).Msg("edit activity list failed")
	}
}

// DeleteSelectedActivities removes all currently selected activities.
func (m *Module) DeleteSelectedActivities(ctx *tgctx.MsgContext) {
	deleted, err := m.tracksvc.DeleteSelectedActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("delete selected activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageDeleteFailed)))
		return
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageDeleted, deleted)))
	m.ShowTrackActivitySelectionMenu(ctx)
}

// ArchiveSelectedActivities moves selected activities to archive.
func (m *Module) ArchiveSelectedActivities(ctx *tgctx.MsgContext) {
	archived, err := m.tracksvc.ArchiveSelectedActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("archive selected activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveFailed)))
		return
	}

	if archived == 0 {
		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveNoneSelected))
		msg.ReplyMarkup = track.TrackArchiveSuccessInlineMenu(ctx.Language)
		_, _ = m.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchived, archived))
	msg.ReplyMarkup = track.TrackArchiveSuccessInlineMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

// ArchiveSelectedActivitiesInPlace archives selected activities and edits current message.
func (m *Module) ArchiveSelectedActivitiesInPlace(ctx *tgctx.MsgContext) {
	archived, err := m.tracksvc.ArchiveSelectedActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("archive selected activities failed")
		edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackArchiveFailed))
		_, _ = m.bot.Send(edit)
		return
	}

	if archived == 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			ctx.ChatID,
			ctx.MessageID,
			i18n.T(ctx.Language, i18n.KeyTrackArchiveNoneSelected),
			track.TrackArchiveSuccessInlineMenu(ctx.Language),
		)
		_, _ = m.bot.Send(edit)
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID,
		ctx.MessageID,
		i18n.T(ctx.Language, i18n.KeyTrackArchived, archived),
		track.TrackArchiveSuccessInlineMenu(ctx.Language),
	)
	_, _ = m.bot.Send(edit)
}

// ShowArchiveMenu opens archive as a new message.
func (m *Module) ShowArchiveMenu(ctx *tgctx.MsgContext) {
	m.renderArchiveMenu(ctx, false)
}

// ShowArchiveMenuInPlace opens archive by editing current message.
func (m *Module) ShowArchiveMenuInPlace(ctx *tgctx.MsgContext) {
	m.renderArchiveMenu(ctx, true)
}

// renderArchiveMenu renders archived activities list in normal or in-place mode.
func (m *Module) renderArchiveMenu(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.tracksvc.ListArchivedActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list archive failed")
		if edit && ctx.MessageID > 0 {
			msg := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackArchiveLoadFailed))
			_, _ = m.bot.Send(msg)
		} else {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveLoadFailed)))
		}
		return
	}

	if len(items) == 0 {
		text := i18n.T(ctx.Language, i18n.KeyTrackArchiveEmpty)
		if edit && ctx.MessageID > 0 {
			msg := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
			_, _ = m.bot.Send(msg)
		} else {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
		}
		return
	}

	text := i18n.T(ctx.Language, i18n.KeyTrackArchiveTitle, len(items))
	if edit && ctx.MessageID > 0 {
		msgReply := tgbotapi.NewMessage(ctx.ChatID, "🗄")
		msgReply.ReplyMarkup = track.TrackArchiveReplyMenu(ctx.Language)
		_, _ = m.bot.Send(msgReply)

		msg := tgbotapi.NewEditMessageTextAndMarkup(
			ctx.ChatID,
			ctx.MessageID,
			text,
			track.TrackArchiveInlineMenu(ctx.Language, items),
		)
		_, _ = m.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ReplyMarkup = track.TrackArchiveInlineMenu(ctx.Language, items)
	msgReply := tgbotapi.NewMessage(ctx.ChatID, "🗄")
	msgReply.ReplyMarkup = track.TrackArchiveReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msgReply)
	_, _ = m.bot.Send(msg)
}

// ShowTrackActivitySelectionMenuInPlace edits current message with activities list.
func (m *Module) ShowTrackActivitySelectionMenuInPlace(ctx *tgctx.MsgContext) {
	msgReply := tgbotapi.NewMessage(ctx.ChatID, "🗂")
	msgReply.ReplyMarkup = track.TrackActivityManageReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msgReply)

	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list activities failed")
		edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackManageLoadFailed))
		_, _ = m.bot.Send(edit)
		return
	}

	if len(items) == 0 {
		edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackManageEmpty))
		_, _ = m.bot.Send(edit)
		return
	}

	selectedCount := countSelectedActivities(items)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID,
		ctx.MessageID,
		i18n.T(ctx.Language, i18n.KeyTrackManageSelectTitle, selectedCount, len(items)),
		track.TrackActivitiesInlineMenu(ctx.Language, items),
	)
	edit.ParseMode = "HTML"
	_, _ = m.bot.Send(edit)
}

// RestoreArchivedActivity restores one activity from archive to active list.
func (m *Module) RestoreArchivedActivity(ctx *tgctx.MsgContext) {
	idRaw := strings.TrimPrefix(ctx.Text, track.TrackCBArchiveRestore)
	activityID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveInvalidItem)))
		return
	}
	activityName := m.findArchivedActivityName(ctx, activityID)

	if err := m.tracksvc.RestoreArchivedActivity(ctx.Ctx, ctx.DBUserID, activityID); err != nil {
		log.Error().Err(err).Msg("restore archived activity failed")
		edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackArchiveRestoreFailed))
		_, _ = m.bot.Send(edit)
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveRestored, activityName)))
	m.ShowArchiveMenuInPlace(ctx)
}

// DeleteArchivedForever permanently removes one archived activity.
func (m *Module) DeleteArchivedForever(ctx *tgctx.MsgContext) {
	idRaw := strings.TrimPrefix(ctx.Text, track.TrackCBArchiveDelete)
	activityID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveInvalidItem)))
		return
	}
	activityName := m.findArchivedActivityName(ctx, activityID)

	if err := m.tracksvc.DeleteArchivedForever(ctx.Ctx, ctx.DBUserID, activityID); err != nil {
		log.Error().Err(err).Msg("delete archived forever failed")
		edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackArchiveDeleteForeverFailed))
		_, _ = m.bot.Send(edit)
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackArchiveDeletedForever, activityName)))
	m.ShowArchiveMenuInPlace(ctx)
}

// findArchivedActivityName resolves archived activity label for confirmations.
func (m *Module) findArchivedActivityName(ctx *tgctx.MsgContext, activityID int64) string {
	items, err := m.tracksvc.ListArchivedActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		return fmt.Sprintf("#%d", activityID)
	}
	for _, item := range items {
		if item.ID == activityID {
			return activityDisplayName(item)
		}
	}
	return fmt.Sprintf("#%d", activityID)
}

// ShowTrackTimerMenu renders timer interval selector: built-in 15/30 min
// choices plus any custom intervals the user has added, all as reply
// buttons (tap = activate; see TrackButtonTimerDelete for removal).
func (m *Module) ShowTrackTimerMenu(ctx *tgctx.MsgContext) {
	custom, err := m.timersvc.ListCustomIntervals(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list custom timers failed")
		custom = nil
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerPickerTitle))
	msg.ReplyMarkup = track.TrackTimerReplyMenu(ctx.Language, track.BuiltInTimerIntervals, custom)
	_, _ = m.bot.Send(msg)
}

// ShowTrackTimerDeleteMenu lists custom intervals as reply buttons; tapping
// one deletes it (see DeleteCustomTimer). If there are none, it just
// re-shows the main timer picker instead of an empty delete screen.
func (m *Module) ShowTrackTimerDeleteMenu(ctx *tgctx.MsgContext) bool {
	custom, err := m.timersvc.ListCustomIntervals(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list custom timers failed")
		custom = nil
	}
	if len(custom) == 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerNoneToDelete)))
		m.ShowTrackTimerMenu(ctx)
		return false
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerPickToDelete))
	msg.ReplyMarkup = track.TrackTimerDeleteReplyMenu(ctx.Language, custom)
	_, _ = m.bot.Send(msg)
	return true
}

// PromptCreateCustomTimer asks user to type a custom interval in minutes.
func (m *Module) PromptCreateCustomTimer(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerCustomPrompt))
	_, _ = m.bot.Send(msg)
}

// ProcessCreateCustomTimer validates and saves a custom interval typed by
// the user, then re-renders the timer picker with it included. Returns true
// once the "waiting for input" state should be cleared.
func (m *Module) ProcessCreateCustomTimer(ctx *tgctx.MsgContext) bool {
	minutes, err := strconv.Atoi(strings.TrimSpace(ctx.Text))
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerNotANumber)))
		return false
	}

	if err := m.timersvc.AddCustomInterval(ctx.Ctx, ctx.DBUserID, minutes); err != nil {
		switch {
		case errors.Is(err, models.ErrCustomTimerInvalidInterval):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(
				ctx.Language, i18n.KeyTrackTimerOutOfRange,
				models.MinCustomTimerMinutes, models.MaxCustomTimerMinutes,
			)))
		case errors.Is(err, models.ErrCustomTimerLimitReached):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(
				ctx.Language, i18n.KeyTrackTimerLimitReached, models.MaxCustomTimersPerUser,
			)))
		default:
			log.Error().Err(err).Msg("add custom timer failed")
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerSaveFailed)))
		}
		return false
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerAdded, minutes)))
	m.ShowTrackTimerMenu(ctx)
	return true
}

// DeleteCustomTimer removes a custom interval and returns to the main timer
// picker with it gone.
func (m *Module) DeleteCustomTimer(ctx *tgctx.MsgContext, intervalMin int) {
	if err := m.timersvc.RemoveCustomInterval(ctx.Ctx, ctx.DBUserID, intervalMin); err != nil && !errors.Is(err, models.ErrCustomTimerNotFound) {
		log.Error().Err(err).Msg("delete custom timer failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerDeleteFailed)))
		return
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerRemoved, intervalMin)))
	m.ShowTrackTimerMenu(ctx)
}

// ActivateTrackTimer enables periodic prompts for selected activities.
func (m *Module) ActivateTrackTimer(ctx *tgctx.MsgContext, intervalMin int) {
	items, err := m.tracksvc.ListSelectedActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("load selected activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerLoadSelectedFailed)))
		return
	}
	if len(items) == 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerNoneSelected)))
		return
	}

	if err := m.timersvc.Activate(ctx.Ctx, ctx.DBUserID, intervalMin); err != nil {
		log.Error().Err(err).Msg("activate timer failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerActivateFailed)))
		return
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerActivated, intervalMin)))
	hide := tgbotapi.NewMessage(ctx.ChatID, " ")
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	m.ShowHomeMenu(ctx)
}

// StopTrackTimer disables active tracking timer.
func (m *Module) StopTrackTimer(ctx *tgctx.MsgContext) {
	if err := m.timersvc.Stop(ctx.Ctx, ctx.DBUserID); err != nil {
		log.Error().Err(err).Msg("stop timer failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerStopFailed)))
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerStopped)))
}

// SendPromptMessage sends periodic "what are you doing now?" prompt.
func (m *Module) SendPromptMessage(ctx context.Context, chatID int64, userID int64, intervalMin int) error {
	items, err := m.tracksvc.ListSelectedActivities(ctx, userID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	// This runs off the scheduler, not a live user request, so there's no
	// tgctx.MsgContext with a pre-loaded Language — look it up directly.
	// chatID is the user's Telegram id (private-chat DMs only, no groups),
	// which is what GetProfileStats expects.
	lang := i18n.Default
	if stats, err := m.profilesvc.GetProfileStats(ctx, chatID); err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("load language for prompt failed")
	} else if stats.Language != nil {
		lang = i18n.Normalize(*stats.Language)
	}

	msg := tgbotapi.NewMessage(chatID, i18n.T(lang, i18n.KeyTrackPromptQuestion))
	msg.ReplyMarkup = track.TrackPromptInlineMenu(lang, items, intervalMin)
	_, err = m.bot.Send(msg)
	return err
}

// RecordPromptAnswer stores one prompt response as tracked time interval.
func (m *Module) RecordPromptAnswer(ctx *tgctx.MsgContext) {
	payload := strings.TrimPrefix(ctx.Text, track.TrackCBPromptActivity)
	parts := strings.Split(payload, ":")
	if len(parts) != 2 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPromptInvalidPayload)))
		return
	}

	activityID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackInvalidActivityID)))
		return
	}

	intervalMin, err := strconv.Atoi(parts[1])
	if err != nil || intervalMin <= 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPromptInvalidInterval)))
		return
	}

	if err := m.timersvc.RecordPromptAnswerWithInterval(ctx.Ctx, ctx.DBUserID, activityID, intervalMin); err != nil {
		log.Error().Err(err).Msg("record prompt answer failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPromptSaveFailed)))
		return
	}

	if ctx.MessageID > 0 {
		del := tgbotapi.NewDeleteMessage(ctx.ChatID, ctx.MessageID)
		_, _ = m.bot.Request(del)
	}

	endAt := apptime.NowIn(ctx.Location)
	startAt := endAt.Add(-time.Duration(intervalMin) * time.Minute)
	activityName := m.findActivityName(ctx, activityID)

	text := i18n.T(
		ctx.Language, i18n.KeyTrackPromptSaved,
		activityName,
		startAt.Format("15:04"),
		endAt.Format("15:04"),
		intervalMin,
	)
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

// findActivityName resolves active activity label for confirmations.
func (m *Module) findActivityName(ctx *tgctx.MsgContext, activityID int64) string {
	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		return fmt.Sprintf("#%d", activityID)
	}
	for _, item := range items {
		if item.ID == activityID {
			return activityDisplayName(item)
		}
	}
	return fmt.Sprintf("#%d", activityID)
}

// formatReportDuration formats duration as "Xh Ym".
func formatReportDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}

// formatDateOrDash returns date in YYYY-MM-DD or dash for empty time.
func formatDateOrDash(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2006-01-02")
}

// percentOf returns percentage string for part/total durations.
func percentOf(part, total time.Duration) string {
	if total <= 0 || part <= 0 {
		return "0%"
	}
	p := (float64(part) / float64(total)) * 100.0
	return fmt.Sprintf("%.1f%%", p)
}

// ShowLearningMenu loads learning stats and renders learning screen.
func (m *Module) ShowLearningMenu(ctx *tgctx.MsgContext) {
	stats, err := m.learningsvc.GetLearningStats(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("GetLearningStats failed")
		msg := tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to load learning data. Please try again.")
		_, _ = m.bot.Send(msg)
		return
	}

	text := learning.LearningMenuText(stats)

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = learning.LearningEntryInlineMenu(stats.TimerActive)

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send learning menu failed")
	}
}

// PromptCreateCollection asks the user to type a name for a new collection.
func (m *Module) PromptCreateCollection(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, "✏️ Send a short one-line name for the new collection (e.g. \"Travel words\"). You'll paste the actual word list on the next step.")
	msg.ReplyMarkup = learning.LearningWaitingReplyMenu()
	_, _ = m.bot.Send(msg)
}

// ProcessCreateCollectionName validates and creates a collection from typed
// text, then immediately prompts for words. done is true once the "waiting
// for a name" state should be cleared (on both success and unrecoverable
// input, matching ProcessCreateActivity's contract).
func (m *Module) ProcessCreateCollectionName(ctx *tgctx.MsgContext) (collectionID int64, done bool) {
	name := strings.TrimSpace(ctx.Text)
	// A pasted "word - translation" list lands here too, since this and the
	// word-entry step share the same plain reply keyboard — catch it before
	// it becomes a garbage multi-line collection name (see the bug report
	// this guards against: a whole word list swallowed as one giant name).
	if strings.Contains(name, "\n") {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ That looks like a word list, not a name. Send a short one-line name first (e.g. \"Travel words\") — you'll paste the word list on the next step."))
		return 0, false
	}
	if len(name) < 2 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Name must be at least 2 characters. Try again:"))
		return 0, false
	}
	if len(name) > 60 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Name is too long (max 60 characters). Try a shorter one:"))
		return 0, false
	}

	id, err := m.learningsvc.CreateCollection(ctx.Ctx, ctx.DBUserID, name)
	if err != nil {
		if errors.Is(err, models.ErrLearningCollectionExists) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ You already have a collection with that name. Try another:"))
			return 0, false
		}
		log.Error().Err(err).Msg("create collection failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to create collection. Please try again later.")
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		return 0, true
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, fmt.Sprintf("📚 Collection *%s* created!", name)))
	m.PromptAddWords(ctx, id, true)
	return id, true
}

// PromptAddWords asks the user to paste word lines for a collection. First
// is true right after collection creation (slightly different copy).
func (m *Module) PromptAddWords(ctx *tgctx.MsgContext, collectionID int64, first bool) {
	text := "➕ Now send words as \"word - translation\", one per line. You can paste several at once. Tap ✅ Done when finished."
	if !first {
		text = "➕ Send more words as \"word - translation\", one per line. Tap ✅ Done when finished."
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ReplyMarkup = learning.LearningAddWordsReplyMenu()
	_, _ = m.bot.Send(msg)
}

// ProcessAddWords parses pasted "word - translation" lines and appends them
// to a collection. The caller keeps the "waiting for words" state active
// afterward — the user may paste more, and only exits via the "Done" button.
func (m *Module) ProcessAddWords(ctx *tgctx.MsgContext, collectionID int64) {
	added, skipped, err := m.learningsvc.AddWordsFromText(ctx.Ctx, ctx.DBUserID, collectionID, ctx.Text)
	if err != nil {
		if errors.Is(err, models.ErrLearningNoWordsParsed) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Couldn't find any \"word - translation\" lines. Try again, e.g.:\napple - яблоко"))
			return
		}
		log.Error().Err(err).Msg("add words failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to save words. Please try again."))
		return
	}

	text := fmt.Sprintf("✅ Added %d word(s).", added)
	if skipped > 0 {
		text += fmt.Sprintf(" (%d line(s) skipped — couldn't parse.)", skipped)
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

// ShowWordBase lists a user's active collections; edit renders it by
// replacing an existing inline message instead of sending a new one, used
// when reached from another inline screen.
func (m *Module) ShowWordBase(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.learningsvc.ListCollections(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list learning collections failed")
		m.sendOrEditLearning(ctx, edit, "⚠️ Failed to load collections.", nil)
		return
	}
	if len(items) == 0 {
		m.sendOrEditLearning(ctx, edit, "🗂 No collections yet. Create one from the Learning menu.", nil)
		return
	}
	menu := learning.LearningWordBaseInlineMenu(items)
	m.sendOrEditLearning(ctx, edit, learning.LearningWordBaseTitle(len(items)), &menu)
}

// ShowCollectionDetail renders one collection's words and actions.
func (m *Module) ShowCollectionDetail(ctx *tgctx.MsgContext, collectionID int64, edit bool) {
	name, err := m.learningsvc.CollectionName(ctx.Ctx, ctx.DBUserID, collectionID)
	if err != nil {
		m.sendOrEditLearning(ctx, edit, "⚠️ Collection not found.", nil)
		return
	}
	words, err := m.learningsvc.ListWords(ctx.Ctx, ctx.DBUserID, collectionID)
	if err != nil {
		log.Error().Err(err).Msg("list words failed")
		m.sendOrEditLearning(ctx, edit, "⚠️ Failed to load words.", nil)
		return
	}

	active := true
	items, err := m.learningsvc.ListCollections(ctx.Ctx, ctx.DBUserID)
	if err == nil {
		for _, it := range items {
			if it.ID == collectionID {
				active = it.Active
			}
		}
	}

	menu := learning.LearningCollectionDetailInlineMenu(collectionID, active, words)
	m.sendOrEditLearning(ctx, edit, learning.LearningCollectionDetailTitle(name, len(words)), &menu)
}

// HandleCollectionToggle flips whether a collection is included in review
// pushes and re-renders its detail view.
func (m *Module) HandleCollectionToggle(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.ToggleCollectionActive(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("toggle collection active failed")
	}
	m.ShowCollectionDetail(ctx, collectionID, true)
}

// HandleWordDelete removes one word and re-renders its collection's detail view.
func (m *Module) HandleWordDelete(ctx *tgctx.MsgContext, wordID, collectionID int64) {
	if err := m.learningsvc.DeleteWord(ctx.Ctx, ctx.DBUserID, wordID); err != nil {
		log.Error().Err(err).Msg("delete word failed")
	}
	m.ShowCollectionDetail(ctx, collectionID, true)
}

// HandleCollectionArchive archives a collection and returns to the word base.
func (m *Module) HandleCollectionArchive(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.ArchiveCollection(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("archive collection failed")
	}
	m.ShowWordBase(ctx, true)
}

// ShowLearningArchiveMenu lists archived collections.
func (m *Module) ShowLearningArchiveMenu(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.learningsvc.ListArchivedCollections(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list archived collections failed")
		m.sendOrEditLearning(ctx, edit, "⚠️ Failed to load archive.", nil)
		return
	}
	if len(items) == 0 {
		m.sendOrEditLearning(ctx, edit, "🔁 No archived collections.", nil)
		return
	}
	menu := learning.LearningArchiveInlineMenu(items)
	m.sendOrEditLearning(ctx, edit, learning.LearningArchiveTitle(len(items)), &menu)
}

// RestoreArchivedCollection moves a collection back to the active list.
func (m *Module) RestoreArchivedCollection(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.RestoreCollection(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("restore collection failed")
	}
	m.ShowLearningArchiveMenu(ctx, true)
}

// DeleteArchivedCollectionForever permanently removes an archived collection.
func (m *Module) DeleteArchivedCollectionForever(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.DeleteCollectionForever(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("delete collection forever failed")
	}
	m.ShowLearningArchiveMenu(ctx, true)
}

// sendOrEditLearning sends a fresh message or edits the current inline
// message with Markdown text and an optional inline keyboard — shared by
// every Learning sub-screen.
func (m *Module) sendOrEditLearning(ctx *tgctx.MsgContext, edit bool, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	if edit && ctx.MessageID > 0 {
		if menu != nil {
			out := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, *menu)
			out.ParseMode = "Markdown"
			_, _ = m.bot.Send(out)
			return
		}
		out := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
		out.ParseMode = "Markdown"
		_, _ = m.bot.Send(out)
		return
	}
	out := tgbotapi.NewMessage(ctx.ChatID, text)
	out.ParseMode = "Markdown"
	if menu != nil {
		out.ReplyMarkup = *menu
	}
	_, _ = m.bot.Send(out)
}

// ShowReviewIntervalPicker shows the review-push interval picker.
func (m *Module) ShowReviewIntervalPicker(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, "🎲 How often should the bot send you a word to review?")
	msg.ReplyMarkup = learning.LearningPushIntervalReplyMenu(learning.BuiltInPushIntervals)
	_, _ = m.bot.Send(msg)
}

// ActivateReviews enables periodic review pushes at the given interval.
func (m *Module) ActivateReviews(ctx *tgctx.MsgContext, intervalMin int) {
	if err := m.learningsvc.Activate(ctx.Ctx, ctx.DBUserID, intervalMin); err != nil {
		log.Error().Err(err).Msg("activate learning reviews failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to activate reviews. Please try again."))
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, fmt.Sprintf("🎲 Reviews activated — a word every %d min.", intervalMin)))
	hide := tgbotapi.NewMessage(ctx.ChatID, " ")
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	m.ShowLearningMenu(ctx)
}

// StopReviews disables periodic review pushes.
func (m *Module) StopReviews(ctx *tgctx.MsgContext) {
	if err := m.learningsvc.Stop(ctx.Ctx, ctx.DBUserID); err != nil {
		log.Error().Err(err).Msg("stop learning reviews failed")
	}
	m.ShowLearningMenu(ctx)
}

// SendLearningPromptMessage sends one review card (term only) for the most
// overdue due word, if any. Called off the scheduler, not a live user
// request — mirrors SendPromptMessage's language lookup.
func (m *Module) SendLearningPromptMessage(ctx context.Context, chatID int64, userID int64) error {
	due, err := m.learningsvc.PickDueWord(ctx, userID)
	if err != nil {
		return err
	}
	if due == nil {
		return nil
	}

	msg := tgbotapi.NewMessage(chatID, learning.LearningReviewCardText(due.CollectionName, due.Term))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = learning.LearningReviewRevealInlineMenu(due.ID)
	_, err = m.bot.Send(msg)
	return err
}

// ShowReviewReveal reveals a review card's translation and grading buttons.
func (m *Module) ShowReviewReveal(ctx *tgctx.MsgContext, wordID int64) {
	collectionName, term, translation, err := m.learningsvc.PeekWord(ctx.Ctx, ctx.DBUserID, wordID)
	if err != nil {
		log.Error().Err(err).Msg("peek word failed")
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID,
		ctx.MessageID,
		learning.LearningReviewRevealedText(collectionName, term, translation),
		learning.LearningReviewGradeInlineMenu(wordID),
	)
	edit.ParseMode = "Markdown"
	if _, err := m.bot.Send(edit); err != nil {
		log.Error().Err(err).Msg("reveal review card failed")
	}
}

// RecordReviewGrade applies the user's answer to a word's SRS schedule and
// replaces the review card with a confirmation.
func (m *Module) RecordReviewGrade(ctx *tgctx.MsgContext, wordID int64, correct bool) {
	_, term, _, err := m.learningsvc.PeekWord(ctx.Ctx, ctx.DBUserID, wordID)
	if err != nil {
		log.Error().Err(err).Msg("peek word before grading failed")
		return
	}

	nextIntervalDays, learned, err := m.learningsvc.GradeAnswer(ctx.Ctx, ctx.DBUserID, wordID, correct)
	if err != nil {
		log.Error().Err(err).Msg("grade answer failed")
		return
	}

	edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, learning.LearningReviewGradedText(term, correct, nextIntervalDays, learned))
	edit.ParseMode = "Markdown"
	if _, err := m.bot.Send(edit); err != nil {
		log.Error().Err(err).Msg("record review grade failed")
	}
}

// ShowSubscriptionMenu loads subscription stats and renders subscription screen.
func (m *Module) ShowSubscriptionMenu(ctx *tgctx.MsgContext) {
	stats, err := m.subscriptionsvc.GetSubscriptionStats(ctx.Ctx, ctx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GetSubscriptionStats failed")
		msg := tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to load subscription data. Please try again.")
		_, _ = m.bot.Send(msg)
		return
	}

	text := subscription.SubscriptionMenuText(stats)

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = subscription.SubscriptionEntryInlineMenu()

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send subscription menu failed")
	}
}

func activityDisplayName(item models.TrackActivityItem) string {
	if item.Emoji != "" {
		return item.Emoji + " " + item.Name
	}
	return item.Name
}

func countSelectedActivities(items []models.TrackActivityItem) int {
	count := 0
	for _, item := range items {
		if item.Selected {
			count++
		}
	}
	return count
}
