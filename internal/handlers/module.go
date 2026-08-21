package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tracker-bot/internal/buttons/admin"
	"tracker-bot/internal/buttons/challenge"
	"tracker-bot/internal/buttons/entry"
	"tracker-bot/internal/buttons/learning"
	"tracker-bot/internal/buttons/onboarding"
	"tracker-bot/internal/buttons/profile"
	"tracker-bot/internal/buttons/roadmap"
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
	adminsvc        service.AdminService
	challengesvc    service.ChallengeService
	roadmapsvc      service.RoadmapService
	roadmapaisvc    service.RoadmapAIService
	// empty miniAppURL means the Mini App is not registered yet, which hides
	// the dashboard button rather than showing a dead one.
	miniAppURL string
	// empty adminUsername disables the admin feature entirely, not "matches everyone" — see IsAdmin.
	adminUsername string
}

const adminUsersPageSize = 15

func New(bot *tgbotapi.BotAPI, entrysvc service.EntryService, profilesvc service.ProfileService, tracksvc service.TrackerService, timersvc service.TimerService, learningsvc service.LearningService, subscriptionsvc service.SubscriptionService, adminsvc service.AdminService, challengesvc service.ChallengeService, roadmapsvc service.RoadmapService, roadmapaisvc service.RoadmapAIService, miniAppURL, adminUsername string) *Module {
	return &Module{
		bot:             bot,
		profilesvc:      profilesvc,
		tracksvc:        tracksvc,
		timersvc:        timersvc,
		learningsvc:     learningsvc,
		subscriptionsvc: subscriptionsvc,
		entrysvc:        entrysvc,
		adminsvc:        adminsvc,
		challengesvc:    challengesvc,
		roadmapsvc:      roadmapsvc,
		roadmapaisvc:    roadmapaisvc,
		miniAppURL:      strings.TrimSpace(miniAppURL),
		adminUsername:   strings.TrimPrefix(strings.TrimSpace(adminUsername), "@"),
	}
}

// the real access check, not just whether the admin button is shown — callers must call this directly.
func (m *Module) IsAdmin(ctx *tgctx.MsgContext) bool {
	if m.adminUsername == "" || ctx.Username == "" {
		return false
	}
	return strings.EqualFold(ctx.Username, m.adminUsername)
}

func (m *Module) ShowEntryMenu(ctx *tgctx.MsgContext) {
	if ctx.IsNewUser {
		m.ShowWelcome(ctx)
		return
	}
	m.ShowHomeMenu(ctx)
}

// call only for a user's actual first /start; every other case should use ShowHomeMenu.
func (m *Module) ShowWelcome(ctx *tgctx.MsgContext) {
	m.sendEntryMenu(ctx, entry.WelcomeText(ctx.Language))
	m.ShowOnboardingStep(ctx, 0, false)
}

func (m *Module) ShowOnboardingStep(ctx *tgctx.MsgContext, step int, edit bool) {
	text := onboarding.StepText(step)
	menu := onboarding.StepInlineMenu(step)

	if edit && ctx.MessageID > 0 {
		out := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, menu)
		out.ParseMode = "Markdown"
		if _, err := m.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("edit onboarding step failed, sending fresh message instead")
			m.ShowOnboardingStep(ctx, step, false)
		}
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = menu
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send onboarding step failed")
	}
}

// every "go home" action except a user's actual first /start (that's ShowWelcome) should use this.
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

func (m *Module) ShowAdminUsersMenu(ctx *tgctx.MsgContext, offset int) {
	if !m.IsAdmin(ctx) {
		return
	}
	reply := tgbotapi.NewMessage(ctx.ChatID, "👥")
	reply.ReplyMarkup = admin.UsersReplyMenu()
	_, _ = m.bot.Send(reply)

	m.sendOrEditAdminUsers(ctx, offset, false)
}

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

func (m *Module) ShowAdminOverview(ctx *tgctx.MsgContext) {
	if !m.IsAdmin(ctx) {
		return
	}
	stats, err := m.adminsvc.GetOverviewStats(ctx.Ctx)
	if err != nil {
		log.Error().Err(err).Msg("get overview stats failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to load overview."))
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, admin.OverviewStatsText(stats)))
}

func (m *Module) ShowAdminUserDetail(ctx *tgctx.MsgContext, dbID int64, edit bool) {
	if !m.IsAdmin(ctx) {
		return
	}
	detail, err := m.adminsvc.GetUserDetail(ctx.Ctx, dbID)
	if err != nil {
		log.Error().Err(err).Msg("get user detail failed")
		m.sendOrEditAdmin(ctx, edit, "⚠️ User not found.", nil)
		return
	}
	menu := admin.UserDetailInlineMenu(dbID, dbID == ctx.DBUserID)
	m.sendOrEditAdmin(ctx, edit, admin.UserDetailText(detail), &menu)
}

func (m *Module) ShowAdminUserDeleteConfirm(ctx *tgctx.MsgContext, dbID int64) {
	if !m.IsAdmin(ctx) {
		return
	}
	if dbID == ctx.DBUserID {
		m.sendOrEditAdmin(ctx, true, "⚠️ You can't delete your own admin account.", nil)
		return
	}
	detail, err := m.adminsvc.GetUserDetail(ctx.Ctx, dbID)
	if err != nil {
		log.Error().Err(err).Msg("get user detail before delete failed")
		m.sendOrEditAdmin(ctx, true, "⚠️ User not found.", nil)
		return
	}
	menu := admin.UserDeleteConfirmInlineMenu(dbID)
	m.sendOrEditAdmin(ctx, true, admin.UserDeleteConfirmText(detail), &menu)
}

func (m *Module) DeleteAdminUser(ctx *tgctx.MsgContext, dbID int64) {
	if !m.IsAdmin(ctx) {
		return
	}
	if dbID == ctx.DBUserID {
		m.sendOrEditAdmin(ctx, true, "⚠️ You can't delete your own admin account.", nil)
		return
	}
	if err := m.entrysvc.DeleteUser(ctx.Ctx, dbID); err != nil {
		log.Error().Err(err).Msg("delete user failed")
		m.sendOrEditAdmin(ctx, true, "⚠️ Failed to delete user.", nil)
		return
	}
	menu := admin.BackToUsersInlineMenu()
	m.sendOrEditAdmin(ctx, true, admin.UserDeletedText(dbID), &menu)
}

func (m *Module) sendOrEditAdmin(ctx *tgctx.MsgContext, edit bool, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	if edit && ctx.MessageID > 0 {
		var out tgbotapi.Chattable
		if menu != nil {
			out = tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, *menu)
		} else {
			out = tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
		}
		if _, err := m.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("edit admin screen failed, sending fresh message instead")
			m.sendOrEditAdmin(ctx, false, text, menu)
		}
		return
	}
	out := tgbotapi.NewMessage(ctx.ChatID, text)
	if menu != nil {
		out.ReplyMarkup = *menu
	}
	if _, err := m.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("send admin screen failed")
	}
}

func (m *Module) PromptAdminBroadcast(ctx *tgctx.MsgContext) {
	if !m.IsAdmin(ctx) {
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, admin.BroadcastPromptText)
	msg.ReplyMarkup = admin.BroadcastWaitingReplyMenu()
	_, _ = m.bot.Send(msg)
}

// returns false if recipients failed to load; caller must not proceed to the confirm step then.
func (m *Module) ShowAdminBroadcastConfirm(ctx *tgctx.MsgContext, text string) bool {
	ids, err := m.entrysvc.ListAllTelegramIDs(ctx.Ctx)
	if err != nil {
		log.Error().Err(err).Msg("list telegram ids for broadcast confirm failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, "⚠️ Failed to prepare broadcast. Please try again."))
		return false
	}
	hide := tgbotapi.NewMessage(ctx.ChatID, admin.BroadcastConfirmText(text, len(ids)))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	msg := tgbotapi.NewMessage(ctx.ChatID, "👆")
	msg.ReplyMarkup = admin.BroadcastConfirmInlineMenu()
	_, _ = m.bot.Send(msg)
	return true
}

func (m *Module) SendAdminBroadcast(ctx *tgctx.MsgContext, text string) {
	ids, err := m.entrysvc.ListAllTelegramIDs(ctx.Ctx)
	if err != nil {
		log.Error().Err(err).Msg("list telegram ids for broadcast failed")
		m.sendOrEditAdmin(ctx, true, "⚠️ Failed to load recipients.", nil)
		return
	}

	sent, failed := 0, 0
	for _, tgID := range ids {
		if _, err := m.bot.Send(tgbotapi.NewMessage(tgID, text)); err != nil {
			failed++
			continue
		}
		sent++
	}
	m.sendOrEditAdmin(ctx, true, admin.BroadcastResultText(sent, failed), nil)
}

func (m *Module) CancelAdminBroadcast(ctx *tgctx.MsgContext) {
	m.sendOrEditAdmin(ctx, true, "❌ Broadcast cancelled.", nil)
}

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
	msg.ReplyMarkup = profile.ProfileEntryInlineMenu(ctx.Language, m.IsAdmin(ctx), m.miniAppURL)

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send profile menu failed")
	}
}

// keys must match the users_allowed_language CHECK constraint (migration 0001_users_init.up.sql).
var languageCodeByButton = map[string]string{
	profile.ProfileButtonLanguageRussian:   "ru",
	profile.ProfileButtonLanguageEnglish:   "en",
	profile.ProfileButtonLanguageGerman:    "de",
	profile.ProfileButtonLanguageUkrainian: "uk",
	profile.ProfileButtonLanguageArabian:   "ar",
}

func (m *Module) ShowLanguagePicker(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLanguagePrompt))
	msg.ReplyMarkup = profile.ProfileLanguageManageReplyMenu(ctx.Language)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send language picker failed")
	}
}

// returns false (no state change) on an unrecognized button, so the caller keeps waiting for a valid tap.
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

	// updates ctx now so this reply + ShowProfileMenu render in the new language; dispatcher invalidates the cache separately.
	ctx.Language = i18n.Normalize(code)

	// ctx.Text is the tapped button's own label, intentionally left untranslated, so it's safe to reuse verbatim.
	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileLanguageSaved, ctx.Text))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)

	m.ShowProfileMenu(ctx)
	return true
}

func (m *Module) ShowLocationRequest(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileTimezonePrompt))
	msg.ReplyMarkup = profile.ProfileLocationReplyMenu(ctx.Language)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send location request failed")
	}
}

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

	// tzName is an IANA identifier (e.g. "Europe/Berlin"), not translatable text, so it's inserted as-is.
	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyProfileTimezoneSaved, tzName))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)

	m.ShowProfileMenu(ctx)
}

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

// ShowTrackTimerStatus renders which activities are currently activated for
// reminders and how long until the next one — a read-only view reached
// from the Track main screen's 🔔 button.
// RemoveActivityFromReminders takes one activity out of the reminder set —
// the only way to shrink it now that checking/unchecking on the
// Select-Activity screen no longer doubles as live reminder membership.
func (m *Module) RemoveActivityFromReminders(ctx *tgctx.MsgContext, activityID int64) {
	if err := m.tracksvc.RemoveFromReminders(ctx.Ctx, ctx.DBUserID, activityID); err != nil && !errors.Is(err, models.ErrActivityNotFound) {
		log.Error().Err(err).Msg("remove from reminders failed")
	}
	m.ShowTrackTimerStatus(ctx)
}

func (m *Module) ShowTrackTimerStatus(ctx *tgctx.MsgContext) {
	intervalMin, nextPingAt, enabled, err := m.timersvc.GetSettings(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("get timer settings failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerStatusLoadFailed)))
		return
	}
	selected, err := m.tracksvc.ListReminderActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list selected activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerStatusLoadFailed)))
		return
	}

	text := track.TrackTimerStatusText(ctx.Language, enabled, intervalMin, nextPingAt, time.Now().UTC(), selected)
	menu := track.TrackTimerStatusInlineMenu(ctx.Language, selected)

	if ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, menu)
		edit.ParseMode = "Markdown"
		if _, err := m.bot.Send(edit); err != nil {
			log.Error().Err(err).Msg("edit timer status failed")
		}
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = menu
	_, _ = m.bot.Send(msg)
}

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

const heatmapWeeks = 8

func heatmapWindow(loc *time.Location) (gridStart, todayMidnight time.Time) {
	today := apptime.NowIn(loc)
	todayMidnight = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	// Days since the most recent Monday (time.Weekday: Sunday=0..Saturday=6).
	mondayOffset := (int(todayMidnight.Weekday()) + 6) % 7
	thisMonday := todayMidnight.AddDate(0, 0, -mondayOffset)
	gridStart = thisMonday.AddDate(0, 0, -7*(heatmapWeeks-1))
	return gridStart, todayMidnight
}

// marks a day if ANY activity had tracked time, not per-activity; see ShowHeatmapDayDetail for the breakdown.
func (m *Module) ShowHeatmap(ctx *tgctx.MsgContext, edit bool) {
	loc := ctx.Location
	gridStart, todayMidnight := heatmapWindow(loc)
	gridEnd := todayMidnight.AddDate(0, 0, 1) // exclusive

	days, err := m.tracksvc.GetTrackedDaysInRange(ctx.Ctx, ctx.DBUserID, gridStart, gridEnd, loc)
	if err != nil {
		log.Error().Err(err).Msg("get tracked days in range failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackLoadFailed)))
		return
	}

	tracked := make(map[string]bool, len(days))
	for _, d := range days {
		tracked[d.Format("2006-01-02")] = true
	}

	text := track.TrackHeatmapText(ctx.Language, gridStart, todayMidnight, tracked, heatmapWeeks)
	menu := track.TrackHeatmapInlineMenu(ctx.Language, gridStart, todayMidnight, tracked, heatmapWeeks)

	if edit && ctx.MessageID > 0 {
		out := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, menu)
		if _, err := m.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("edit heatmap failed, sending fresh message instead")
			m.ShowHeatmap(ctx, false)
		}
		return
	}
	out := tgbotapi.NewMessage(ctx.ChatID, text)
	out.ReplyMarkup = menu
	if _, err := m.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("send heatmap failed")
	}
}

func (m *Module) ShowHeatmapDayDetail(ctx *tgctx.MsgContext, day time.Time) {
	loc := ctx.Location
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.AddDate(0, 0, 1)

	var b strings.Builder
	b.WriteString(track.TrackHeatmapDayTitle(ctx.Language, dayStart))
	b.WriteString("\n\n")

	activities, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list activities for heatmap day detail failed")
	}
	activityIDs := make([]int64, 0, len(activities))
	for _, a := range activities {
		activityIDs = append(activityIDs, a.ID)
	}

	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayTrackedHeader))
	b.WriteString("\n")
	if len(activityIDs) == 0 {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayNoActivity))
		b.WriteString("\n")
	} else {
		items, err := m.tracksvc.GetPeriodReport(ctx.Ctx, ctx.DBUserID, dayStart, dayEnd, activityIDs, loc)
		if err != nil {
			log.Error().Err(err).Msg("get period report for heatmap day detail failed")
		}
		if len(items.Activities) == 0 {
			b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayNoActivity))
			b.WriteString("\n")
		} else {
			for _, a := range items.Activities {
				b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayActivityLine, activityDisplayNameFromParts(a.Emoji, a.Name), formatReportDuration(a.Duration), a.Sessions))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayReviewsHeader))
	b.WriteString("\n")
	reviews, err := m.learningsvc.ListReviewsOnDay(ctx.Ctx, ctx.DBUserID, dayStart, dayEnd)
	if err != nil {
		log.Error().Err(err).Msg("list reviews on day for heatmap detail failed")
	}
	if len(reviews) == 0 {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayNoReviews))
		b.WriteString("\n")
	} else {
		for _, r := range reviews {
			mark := "🔁"
			if r.Correct {
				mark = "✅"
			}
			b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackHeatmapDayReviewLine, mark, r.Term, r.Translation))
		}
	}

	menu := track.TrackHeatmapDayDetailInlineMenu(ctx.Language)
	if ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, b.String(), menu)
		if _, err := m.bot.Send(edit); err == nil {
			return
		}
		log.Error().Msg("edit heatmap day detail failed, sending fresh message instead")
	}
	out := tgbotapi.NewMessage(ctx.ChatID, b.String())
	out.ReplyMarkup = menu
	if _, err := m.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("send heatmap day detail failed")
	}
}

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

	if granularity == "hour" {
		rows, err := m.tracksvc.GetHourlyBucketsByActivity(ctx.Ctx, ctx.DBUserID, from, to.Add(24*time.Hour), activityIDs, ctx.Location)
		if err != nil || len(rows) == 0 {
			return
		}
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityByHours))
		appendHourlyByActivityLines(ctx, b, rows, labelFmt)
		return
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
	}

	for i := range buckets {
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityBucketLine, buckets[i].Format(labelFmt), formatReportDuration(durs[i])))
	}
}

// assumes rows are pre-sorted by hour, then duration descending — grouping breaks silently otherwise.
func appendHourlyByActivityLines(ctx *tgctx.MsgContext, b *strings.Builder, rows []models.HourActivityDuration, labelFmt string) {
	i := 0
	for i < len(rows) {
		bucket := rows[i].BucketStart
		parts := make([]string, 0, 4)
		for i < len(rows) && rows[i].BucketStart.Equal(bucket) {
			parts = append(parts, fmt.Sprintf("%s %s", activityDisplayNameFromParts(rows[i].Emoji, rows[i].Name), formatReportDuration(rows[i].Duration)))
			i++
		}
		b.WriteString(i18n.T(ctx.Language, i18n.KeyTrackGranularityBucketLine, bucket.Format(labelFmt), strings.Join(parts, ", ")))
	}
}

func activityDisplayNameFromParts(emoji, name string) string {
	if emoji != "" {
		return emoji + " " + name
	}
	return name
}

func (m *Module) ShowTodayReport(ctx *tgctx.MsgContext) {
	m.ShowTodayChart(ctx)
}

func (m *Module) ShowTodayReportBySelected(ctx *tgctx.MsgContext) {
	m.ShowTodaySelectActivities(ctx, map[int64]bool{})
}

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

func (m *Module) PromptCreateActivity(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackCreatePrompt))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = track.TrackActivityManageReplyMenu(ctx.Language)

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send create activity prompt failed")
	}
}

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

// PromptSetActivityTarget asks for a daily time target (in minutes) for one
// activity — the 🎯 button on the activity list.
func (m *Module) PromptSetActivityTarget(ctx *tgctx.MsgContext, activityID int64) {
	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list activities failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackManageLoadFailed)))
		return
	}
	name := activityDisplayName(models.TrackActivityItem{})
	for _, it := range items {
		if it.ID == activityID {
			name = activityDisplayName(it)
			break
		}
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, track.TrackActivityTargetPromptText(ctx.Language, name))
	msg.ParseMode = "Markdown"
	_, _ = m.bot.Send(msg)
}

// ProcessSetActivityTarget parses the typed minutes and saves the
// activity's target. Same contract as ProcessCreateCustomTimer: every
// failure path returns false so the caller keeps waiting for a retry.
func (m *Module) ProcessSetActivityTarget(ctx *tgctx.MsgContext, activityID int64) bool {
	minutes, err := strconv.Atoi(strings.TrimSpace(ctx.Text))
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackActivityTargetInvalid)))
		return false
	}

	if err := m.tracksvc.SetActivityTarget(ctx.Ctx, ctx.DBUserID, activityID, minutes); err != nil {
		if errors.Is(err, models.ErrActivityTargetInvalid) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackActivityTargetInvalid)))
			return false
		}
		log.Error().Err(err).Msg("set activity target failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackActivityTargetSaveFailed)))
		return false
	}

	items, err := m.tracksvc.ListActivities(ctx.Ctx, ctx.DBUserID)
	name := ""
	if err == nil {
		for _, it := range items {
			if it.ID == activityID {
				name = activityDisplayName(it)
				break
			}
		}
	}
	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackActivityTargetSavedFmt, minutes, name))
	confirm.ParseMode = "Markdown"
	_, _ = m.bot.Send(confirm)
	m.ShowTrackActivitySelectionMenu(ctx)
	return true
}

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

func (m *Module) ShowArchiveMenu(ctx *tgctx.MsgContext) {
	m.renderArchiveMenu(ctx, false)
}

func (m *Module) ShowArchiveMenuInPlace(ctx *tgctx.MsgContext) {
	m.renderArchiveMenu(ctx, true)
}

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

func (m *Module) PromptCreateCustomTimer(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerCustomPrompt))
	_, _ = m.bot.Send(msg)
}

// unlike ProcessCreateCollectionName/ProcessCreateRoadmapName, every failure path returns false so the caller keeps waiting for a retry.
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

func (m *Module) DeleteCustomTimer(ctx *tgctx.MsgContext, intervalMin int) {
	if err := m.timersvc.RemoveCustomInterval(ctx.Ctx, ctx.DBUserID, intervalMin); err != nil && !errors.Is(err, models.ErrCustomTimerNotFound) {
		log.Error().Err(err).Msg("delete custom timer failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerDeleteFailed)))
		return
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerRemoved, intervalMin)))
	m.ShowTrackTimerMenu(ctx)
}

// ActivateTrackTimer adds whatever is currently checked on the
// Select-Activity screen to the (persistent, additive) reminder set —
// existing members are untouched, not replaced — then starts/updates the
// schedule. See TrackerService.AddSelectedToReminders for why this doesn't
// just reuse the checkboxes as live reminder truth.
func (m *Module) ActivateTrackTimer(ctx *tgctx.MsgContext, intervalMin int) {
	added, err := m.tracksvc.AddSelectedToReminders(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("add selected to reminders failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerLoadSelectedFailed)))
		return
	}

	items, err := m.tracksvc.ListReminderActivities(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list reminder activities failed")
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

	text := i18n.T(ctx.Language, i18n.KeyTrackTimerActivated, intervalMin)
	if added > 0 {
		text = i18n.T(ctx.Language, i18n.KeyTrackTimerActivatedAddedFmt, added, intervalMin)
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
	hide := tgbotapi.NewMessage(ctx.ChatID, " ")
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	m.ShowHomeMenu(ctx)
}

// PromptStopTrackTimer asks for confirmation before disabling reminders —
// the reminder push's 🛑 button used to stop immediately, which sat right
// below the (sometimes single) activity button and was too easy to hit by
// accident, silently killing reminders with no way back except re-Activate.
func (m *Module) PromptStopTrackTimer(ctx *tgctx.MsgContext) {
	menu := track.TrackStopTimerConfirmInlineMenu(ctx.Language)
	if ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackTimerStopConfirmPrompt), menu)
		if _, err := m.bot.Send(edit); err != nil {
			log.Error().Err(err).Msg("edit stop-timer confirm failed")
		}
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerStopConfirmPrompt))
	msg.ReplyMarkup = menu
	_, _ = m.bot.Send(msg)
}

// ConfirmStopTrackTimer actually disables reminders, after PromptStopTrackTimer's confirm step.
func (m *Module) ConfirmStopTrackTimer(ctx *tgctx.MsgContext) {
	if err := m.timersvc.Stop(ctx.Ctx, ctx.DBUserID); err != nil {
		log.Error().Err(err).Msg("stop timer failed")
		if ctx.MessageID > 0 {
			_, _ = m.bot.Send(tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, i18n.KeyTrackTimerStopFailed)))
		} else {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackTimerStopFailed)))
		}
		return
	}
	text := i18n.T(ctx.Language, i18n.KeyTrackTimerStopped)
	if ctx.MessageID > 0 {
		_, _ = m.bot.Send(tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text))
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

// CancelStopTrackTimer backs out of the confirm step without touching the timer.
func (m *Module) CancelStopTrackTimer(ctx *tgctx.MsgContext) {
	text := i18n.T(ctx.Language, i18n.KeyTrackTimerStopCancelled)
	if ctx.MessageID > 0 {
		_, _ = m.bot.Send(tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text))
		return
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

// dueAt is embedded in the buttons so a late answer still credits the originally scheduled window.
func (m *Module) SendPromptMessage(ctx context.Context, chatID int64, userID int64, intervalMin int, dueAt time.Time) error {
	items, err := m.tracksvc.ListReminderActivities(ctx, userID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	// scheduler call, no MsgContext to read language from — look it up directly; chatID is the DM chat id.
	lang := i18n.Default
	if stats, err := m.profilesvc.GetProfileStats(ctx, chatID); err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("load language for prompt failed")
	} else if stats.Language != nil {
		lang = i18n.Normalize(*stats.Language)
	}

	msg := tgbotapi.NewMessage(chatID, i18n.T(lang, i18n.KeyTrackPromptQuestion))
	msg.ReplyMarkup = track.TrackPromptInlineMenu(lang, items, intervalMin, dueAt)
	_, err = m.bot.Send(msg)
	return err
}

func (m *Module) RecordPromptAnswer(ctx *tgctx.MsgContext) {
	payload := strings.TrimPrefix(ctx.Text, track.TrackCBPromptActivity)
	parts := strings.Split(payload, ":")
	if len(parts) != 3 {
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

	dueAtUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || dueAtUnix <= 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPromptInvalidPayload)))
		return
	}
	// dueAt (not "now") is credited so answering several stacked prompts together doesn't double-credit one moment.
	dueAt := time.Unix(dueAtUnix, 0).UTC()

	if err := m.timersvc.RecordPromptAnswerWithInterval(ctx.Ctx, ctx.DBUserID, activityID, intervalMin, dueAt); err != nil {
		log.Error().Err(err).Msg("record prompt answer failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyTrackPromptSaveFailed)))
		return
	}

	if ctx.MessageID > 0 {
		del := tgbotapi.NewDeleteMessage(ctx.ChatID, ctx.MessageID)
		_, _ = m.bot.Request(del)
	}

	endAt := dueAt.In(ctx.Location)
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

func formatDateOrDash(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2006-01-02")
}

func percentOf(part, total time.Duration) string {
	if total <= 0 || part <= 0 {
		return "0%"
	}
	p := (float64(part) / float64(total)) * 100.0
	return fmt.Sprintf("%.1f%%", p)
}

func (m *Module) ShowLearningMenu(ctx *tgctx.MsgContext) {
	stats, err := m.learningsvc.GetLearningStats(ctx.Ctx, ctx.DBUserID, ctx.Location)
	if err != nil {
		log.Error().Err(err).Msg("GetLearningStats failed")
		msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningLoadFailed))
		_, _ = m.bot.Send(msg)
		return
	}

	text := learning.LearningMenuText(ctx.Language, stats)

	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = learning.LearningEntryInlineMenu(ctx.Language, stats.TimerActive)

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send learning menu failed")
	}
}

func (m *Module) ShowLearningStatsDetail(ctx *tgctx.MsgContext, edit bool) {
	detail, err := m.learningsvc.GetStatsDetail(ctx.Ctx, ctx.DBUserID, ctx.Location)
	if err != nil {
		log.Error().Err(err).Msg("get learning stats detail failed")
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningStatsLoadFailed), nil)
		return
	}
	menu := learning.LearningBackToMainInlineMenu(ctx.Language)
	m.sendOrEditLearning(ctx, edit, learning.LearningStatsDetailText(ctx.Language, detail), &menu)
}

func (m *Module) PromptCreateCollection(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreatePrompt))
	msg.ReplyMarkup = learning.LearningWaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

// done is true on success and on unrecoverable failure, not just success — caller must stop waiting either way.
func (m *Module) ProcessCreateCollectionName(ctx *tgctx.MsgContext) (collectionID int64, done bool) {
	name := strings.TrimSpace(ctx.Text)
	// catches a pasted word list before it becomes one giant garbage collection name (a real bug this guards against).
	if strings.Contains(name, "\n") {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateNotAList)))
		return 0, false
	}
	if len(name) < 2 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateTooShort)))
		return 0, false
	}
	if len(name) > 60 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateTooLong)))
		return 0, false
	}

	id, err := m.learningsvc.CreateCollection(ctx.Ctx, ctx.DBUserID, name)
	if err != nil {
		if errors.Is(err, models.ErrLearningCollectionExists) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateExists)))
			return 0, false
		}
		log.Error().Err(err).Msg("create collection failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		return 0, true
	}

	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateConfirmed, name)))
	m.PromptAddWords(ctx, id, true)
	return id, true
}

func (m *Module) PromptAddWords(ctx *tgctx.MsgContext, collectionID int64, first bool) {
	key := i18n.KeyLearningAddWordsPromptMore
	if first {
		key = i18n.KeyLearningAddWordsPromptFirst
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, key))
	msg.ReplyMarkup = learning.LearningAddWordsReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ProcessAddWords(ctx *tgctx.MsgContext, collectionID int64) {
	added, skipped, err := m.learningsvc.AddWordsFromText(ctx.Ctx, ctx.DBUserID, collectionID, ctx.Text)
	if err != nil {
		if errors.Is(err, models.ErrLearningNoWordsParsed) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningAddWordsNoneParsed)))
			return
		}
		log.Error().Err(err).Msg("add words failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningAddWordsFailed)))
		return
	}

	text := i18n.T(ctx.Language, i18n.KeyLearningAddWordsAdded, added)
	if skipped > 0 {
		text += i18n.T(ctx.Language, i18n.KeyLearningAddWordsSkipped, skipped)
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

func (m *Module) ShowWordBase(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.learningsvc.ListCollections(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list learning collections failed")
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningCollectionsFailed), nil)
		return
	}
	if len(items) == 0 {
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningWordBaseEmpty), nil)
		return
	}
	menu := learning.LearningWordBaseInlineMenu(ctx.Language, items)
	m.sendOrEditLearning(ctx, edit, learning.LearningWordBaseTitle(ctx.Language, len(items)), &menu)
}

// toggling here applies immediately regardless of whether reviews are already running; ShowReviewIntervalPicker is the separate "not yet activated" step.
func (m *Module) ShowReviewCollectionPicker(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.learningsvc.ListCollections(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list learning collections for review picker failed")
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningCollectionsFailed), nil)
		return
	}
	if len(items) == 0 {
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningReviewPickEmptyTitle), nil)
		return
	}

	active := 0
	for _, it := range items {
		if it.Active {
			active++
		}
	}

	stats, err := m.learningsvc.GetLearningStats(ctx.Ctx, ctx.DBUserID, ctx.Location)
	if err != nil {
		log.Error().Err(err).Msg("get learning stats for review picker failed")
	}

	menu := learning.LearningReviewPickInlineMenu(ctx.Language, items, stats.TimerActive)
	m.sendOrEditLearning(ctx, edit, learning.LearningReviewPickTitle(ctx.Language, active, stats.TimerActive, stats.TimerInterval), &menu)
}

func (m *Module) HandleReviewPickToggle(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.ToggleCollectionActive(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("toggle collection active from review picker failed")
	}
	m.ShowReviewCollectionPicker(ctx, true)
}

// ok is false when nothing is selected yet — caller must not advance past the picker.
func (m *Module) HandleReviewContinue(ctx *tgctx.MsgContext) (ok bool) {
	items, err := m.learningsvc.ListCollections(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list learning collections before review continue failed")
		return false
	}
	for _, it := range items {
		if it.Active {
			m.ShowReviewIntervalPicker(ctx)
			return true
		}
	}
	menu := learning.LearningReviewPickInlineMenu(ctx.Language, items, false)
	text := i18n.T(ctx.Language, i18n.KeyLearningReviewPickNeedOne) + "\n\n" + learning.LearningReviewPickTitle(ctx.Language, 0, false, 0)
	m.sendOrEditLearning(ctx, true, text, &menu)
	return false
}

func (m *Module) ShowCollectionDetail(ctx *tgctx.MsgContext, collectionID int64, edit bool) {
	name, err := m.learningsvc.CollectionName(ctx.Ctx, ctx.DBUserID, collectionID)
	if err != nil {
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningCollectionNotFound), nil)
		return
	}
	words, err := m.learningsvc.ListWords(ctx.Ctx, ctx.DBUserID, collectionID)
	if err != nil {
		log.Error().Err(err).Msg("list words failed")
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningWordsLoadFailed), nil)
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

	menu := learning.LearningCollectionDetailInlineMenu(ctx.Language, collectionID, active, words)
	m.sendOrEditLearning(ctx, edit, learning.LearningCollectionDetailTitle(ctx.Language, name, len(words)), &menu)
}

func (m *Module) PromptRenameCollection(ctx *tgctx.MsgContext, collectionID int64) {
	name, err := m.learningsvc.CollectionName(ctx.Ctx, ctx.DBUserID, collectionID)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCollectionNotFound)))
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, learning.LearningRenamePromptText(ctx.Language, name))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = learning.LearningWaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

// done is false only while the name itself is invalid/duplicate; true even on an unrecoverable save failure.
func (m *Module) ProcessRenameCollection(ctx *tgctx.MsgContext, collectionID int64) (done bool) {
	name := strings.TrimSpace(ctx.Text)
	if strings.Contains(name, "\n") || len(name) < 2 || len(name) > 60 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonNameSingleLineInvalid)))
		return false
	}

	if err := m.learningsvc.RenameCollection(ctx.Ctx, ctx.DBUserID, collectionID, name); err != nil {
		if errors.Is(err, models.ErrLearningCollectionExists) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningCreateExists)))
			return false
		}
		log.Error().Err(err).Msg("rename collection failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningRenameFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		m.ShowCollectionDetail(ctx, collectionID, false)
		return true
	}

	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningRenamed, name))
	hide.ParseMode = "Markdown"
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	m.ShowCollectionDetail(ctx, collectionID, false)
	return true
}

func (m *Module) HandleCollectionToggle(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.ToggleCollectionActive(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("toggle collection active failed")
	}
	m.ShowCollectionDetail(ctx, collectionID, true)
}

func (m *Module) HandleWordDelete(ctx *tgctx.MsgContext, wordID, collectionID int64) {
	if err := m.learningsvc.DeleteWord(ctx.Ctx, ctx.DBUserID, wordID); err != nil {
		log.Error().Err(err).Msg("delete word failed")
	}
	m.ShowCollectionDetail(ctx, collectionID, true)
}

func (m *Module) HandleCollectionArchive(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.ArchiveCollection(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("archive collection failed")
	}
	m.ShowWordBase(ctx, true)
}

func (m *Module) ShowLearningArchiveMenu(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.learningsvc.ListArchivedCollections(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list archived collections failed")
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyTrackArchiveLoadFailed), nil)
		return
	}
	if len(items) == 0 {
		m.sendOrEditLearning(ctx, edit, i18n.T(ctx.Language, i18n.KeyLearningArchiveEmpty), nil)
		return
	}
	menu := learning.LearningArchiveInlineMenu(ctx.Language, items)
	m.sendOrEditLearning(ctx, edit, learning.LearningArchiveTitle(ctx.Language, len(items)), &menu)
}

func (m *Module) RestoreArchivedCollection(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.RestoreCollection(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("restore collection failed")
	}
	m.ShowLearningArchiveMenu(ctx, true)
}

func (m *Module) DeleteArchivedCollectionForever(ctx *tgctx.MsgContext, collectionID int64) {
	if err := m.learningsvc.DeleteCollectionForever(ctx.Ctx, ctx.DBUserID, collectionID); err != nil {
		log.Error().Err(err).Msg("delete collection forever failed")
	}
	m.ShowLearningArchiveMenu(ctx, true)
}

func (m *Module) sendOrEditLearning(ctx *tgctx.MsgContext, edit bool, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	if edit && ctx.MessageID > 0 {
		var out tgbotapi.Chattable
		if menu != nil {
			e := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, *menu)
			e.ParseMode = "Markdown"
			out = e
		} else {
			e := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
			e.ParseMode = "Markdown"
			out = e
		}
		if _, err := m.bot.Send(out); err != nil {
			// edit can fail (stale/deleted message); previously swallowed silently, leaving a frozen
			// screen — fall back to a fresh message instead.
			log.Error().Err(err).Msg("edit learning screen failed, sending fresh message instead")
			m.sendOrEditLearning(ctx, false, text, menu)
		}
		return
	}
	out := tgbotapi.NewMessage(ctx.ChatID, text)
	out.ParseMode = "Markdown"
	if menu != nil {
		out.ReplyMarkup = *menu
	}
	if _, err := m.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("send learning screen failed")
	}
}

func (m *Module) ShowReviewIntervalPicker(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningReviewIntervalPrompt))
	msg.ReplyMarkup = learning.LearningPushIntervalReplyMenu(ctx.Language, learning.BuiltInPushIntervals)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ActivateReviews(ctx *tgctx.MsgContext, intervalMin int) {
	if err := m.learningsvc.Activate(ctx.Ctx, ctx.DBUserID, intervalMin); err != nil {
		log.Error().Err(err).Msg("activate learning reviews failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningReviewActivateFailed)))
		return
	}
	// clears the old keyboard on this same message to avoid a window where a stray tap lands as text
	// on the next screen.
	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyLearningReviewActivated, intervalMin))
	confirm.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(confirm)
	m.ShowLearningMenu(ctx)
}

func (m *Module) StopReviews(ctx *tgctx.MsgContext) {
	if err := m.learningsvc.Stop(ctx.Ctx, ctx.DBUserID); err != nil {
		log.Error().Err(err).Msg("stop learning reviews failed")
	}
	m.ShowLearningMenu(ctx)
}

// scheduler call (no MsgContext) — same language lookup as SendPromptMessage.
func (m *Module) SendLearningPromptMessage(ctx context.Context, chatID int64, userID int64) error {
	due, err := m.learningsvc.PickDueWord(ctx, userID)
	if err != nil {
		return err
	}
	if due == nil {
		return nil
	}

	lang := i18n.Default
	if stats, err := m.profilesvc.GetProfileStats(ctx, chatID); err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("load language for learning prompt failed")
	} else if stats.Language != nil {
		lang = i18n.Normalize(*stats.Language)
	}

	msg := tgbotapi.NewMessage(chatID, learning.LearningReviewCardText(lang, due.CollectionName, due.Term))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = learning.LearningReviewRevealInlineMenu(lang, due.ID)
	_, err = m.bot.Send(msg)
	return err
}

func (m *Module) ShowReviewReveal(ctx *tgctx.MsgContext, wordID int64) {
	collectionName, term, translation, err := m.learningsvc.PeekWord(ctx.Ctx, ctx.DBUserID, wordID)
	if err != nil {
		log.Error().Err(err).Msg("peek word failed")
		return
	}

	// previews each grade's resulting delay upfront so buttons show "in 10m"/"in 1d" before the
	// user picks, Anki-style.
	again, hard, good, easy, err := m.learningsvc.PreviewGradeDelays(ctx.Ctx, ctx.DBUserID, wordID)
	if err != nil {
		log.Error().Err(err).Msg("preview grade delays failed")
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID,
		ctx.MessageID,
		learning.LearningReviewRevealedText(ctx.Language, collectionName, term, translation),
		learning.LearningReviewGradeInlineMenu(ctx.Language, wordID, again, hard, good, easy),
	)
	edit.ParseMode = "Markdown"
	if _, err := m.bot.Send(edit); err != nil {
		log.Error().Err(err).Msg("reveal review card failed")
	}
}

func (m *Module) RecordReviewGrade(ctx *tgctx.MsgContext, wordID int64, grade models.LearningGrade) {
	_, term, _, err := m.learningsvc.PeekWord(ctx.Ctx, ctx.DBUserID, wordID)
	if err != nil {
		log.Error().Err(err).Msg("peek word before grading failed")
		return
	}

	nextReviewAt, learned, err := m.learningsvc.GradeAnswer(ctx.Ctx, ctx.DBUserID, wordID, grade)
	if err != nil {
		log.Error().Err(err).Msg("grade answer failed")
		return
	}

	// must replace (not just edit text) to clear the stale grade buttons — otherwise a second tap
	// can silently re-grade the same word.
	menu := learning.LearningBackToMainInlineMenu(ctx.Language)
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		ctx.ChatID, ctx.MessageID,
		learning.LearningReviewGradedText(ctx.Language, term, grade, nextReviewAt, learned),
		menu,
	)
	edit.ParseMode = "Markdown"
	if _, err := m.bot.Send(edit); err != nil {
		log.Error().Err(err).Msg("record review grade failed")
	}
}

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

func (m *Module) ShowChallengesMenu(ctx *tgctx.MsgContext) {
	items, err := m.challengesvc.ListChallenges(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list challenges failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyChallengeListLoadFailed)))
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, challenge.ListTitle(ctx.Language, len(items)))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = challenge.ListInlineMenu(ctx.Language, items)
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send challenges menu failed")
	}
}

func (m *Module) PromptCreateChallenge(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, challenge.CreatePromptNameText(ctx.Language))
	msg.ReplyMarkup = challenge.WaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ProcessCreateChallengeName(ctx *tgctx.MsgContext) (name string, ok bool) {
	name = strings.TrimSpace(ctx.Text)
	if strings.Contains(name, "\n") || len(name) < 2 || len(name) > 60 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonNameSingleLineInvalid)))
		return "", false
	}
	hide := tgbotapi.NewMessage(ctx.ChatID, challenge.CreatePromptRangeText(ctx.Language))
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	return name, true
}

func (m *Module) ShowCreateChallengeCalendar(ctx *tgctx.MsgContext, month, from, to time.Time) {
	msg := tgbotapi.NewMessage(ctx.ChatID, challenge.CreateCalendarHeaderText(ctx.Language))
	msg.ReplyMarkup = challenge.CalendarInlineMenu(ctx.Language, month, from, to)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ShowCreateChallengeCalendarInPlace(ctx *tgctx.MsgContext, month, from, to time.Time) {
	if ctx.MessageID == 0 {
		return
	}
	edit := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, challenge.CreateCalendarHeaderText(ctx.Language), challenge.CalendarInlineMenu(ctx.Language, month, from, to))
	if _, err := m.bot.Send(edit); err != nil {
		log.Error().Err(err).Msg("edit create-challenge calendar failed")
	}
}

func (m *Module) CreateChallenge(ctx *tgctx.MsgContext, name string, from, to time.Time) bool {
	id, err := m.challengesvc.CreateChallenge(ctx.Ctx, ctx.DBUserID, name, from, to, ctx.Location)
	if err != nil {
		var msgKey string
		switch {
		case errors.Is(err, models.ErrChallengeExists):
			msgKey = i18n.KeyChallengeCreateExists
		case errors.Is(err, models.ErrChallengeInvalidRange):
			msgKey = i18n.KeyChallengeCreateInvalidRange
		case errors.Is(err, models.ErrChallengeInvalidName):
			msgKey = i18n.KeyCommonNameSingleLineInvalid
		default:
			log.Error().Err(err).Msg("create challenge failed")
			msgKey = i18n.KeyChallengeCreateFailed
		}
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, msgKey)))
		m.ShowChallengesMenu(ctx)
		return false
	}
	totalDays := int(to.Sub(from).Hours()/24) + 1
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, challenge.CreatedText(ctx.Language, name, totalDays)))
	_ = id
	m.ShowChallengesMenu(ctx)
	return true
}

func (m *Module) ShowChallengeGrid(ctx *tgctx.MsgContext, challengeID int64, edit bool) {
	item, err := m.challengesvc.GetChallenge(ctx.Ctx, ctx.DBUserID, challengeID)
	if err != nil {
		m.sendOrEditChallenge(ctx, edit, i18n.T(ctx.Language, i18n.KeyChallengeNotFound), nil)
		return
	}
	days, err := m.challengesvc.ListDays(ctx.Ctx, ctx.DBUserID, challengeID)
	if err != nil {
		log.Error().Err(err).Msg("list challenge days failed")
		m.sendOrEditChallenge(ctx, edit, i18n.T(ctx.Language, i18n.KeyChallengeLoadFailed), nil)
		return
	}
	today := apptime.NowIn(ctx.Location)
	todayMidnight := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	menu := challenge.GridInlineMenu(ctx.Language, challengeID, days, todayMidnight)
	m.sendOrEditChallenge(ctx, edit, challenge.GridTitle(ctx.Language, item), &menu)
}

// ShowChallengeDayConfirm renders the day-square tap screen: the existing
// Done/Skip buttons, now with a progress donut, trend strip, and streak
// above them (via challengesvc.GetDayDetail) instead of a bare status line.
func (m *Module) ShowChallengeDayConfirm(ctx *tgctx.MsgContext, challengeID int64, day time.Time) {
	item, err := m.challengesvc.GetChallenge(ctx.Ctx, ctx.DBUserID, challengeID)
	if err != nil {
		m.sendOrEditChallenge(ctx, true, i18n.T(ctx.Language, i18n.KeyChallengeNotFound), nil)
		return
	}
	today := apptime.NowIn(ctx.Location)
	todayMidnight := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	detail, err := m.challengesvc.GetDayDetail(ctx.Ctx, ctx.DBUserID, challengeID, day, todayMidnight)
	if err != nil {
		m.sendOrEditChallenge(ctx, true, i18n.T(ctx.Language, i18n.KeyChallengeDayNotFound), nil)
		return
	}
	menu := challenge.DayConfirmInlineMenu(ctx.Language, challengeID, day)
	m.sendOrEditChallenge(ctx, true, challenge.DayConfirmTitle(ctx.Language, item, detail), &menu)
}

func (m *Module) MarkChallengeDay(ctx *tgctx.MsgContext, challengeID int64, day time.Time, done bool) {
	if err := m.challengesvc.MarkDay(ctx.Ctx, ctx.DBUserID, challengeID, day, done); err != nil {
		log.Error().Err(err).Msg("mark challenge day failed")
	}
	m.ShowChallengeGrid(ctx, challengeID, true)
}

func (m *Module) ArchiveChallenge(ctx *tgctx.MsgContext, challengeID int64) {
	if err := m.challengesvc.ArchiveChallenge(ctx.Ctx, ctx.DBUserID, challengeID); err != nil {
		log.Error().Err(err).Msg("archive challenge failed")
	}
	m.ShowChallengesMenu(ctx)
}

func (m *Module) ShowChallengeArchive(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.challengesvc.ListArchivedChallenges(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list archived challenges failed")
		m.sendOrEditChallenge(ctx, edit, i18n.T(ctx.Language, i18n.KeyTrackArchiveLoadFailed), nil)
		return
	}
	if len(items) == 0 {
		m.sendOrEditChallenge(ctx, edit, i18n.T(ctx.Language, i18n.KeyChallengeArchiveEmpty), nil)
		return
	}
	menu := challenge.ArchiveInlineMenu(ctx.Language, items)
	m.sendOrEditChallenge(ctx, edit, challenge.ArchiveTitle(ctx.Language, len(items)), &menu)
}

func (m *Module) RestoreChallenge(ctx *tgctx.MsgContext, challengeID int64) {
	if err := m.challengesvc.RestoreChallenge(ctx.Ctx, ctx.DBUserID, challengeID, ctx.Location); err != nil {
		log.Error().Err(err).Msg("restore challenge failed")
	}
	m.ShowChallengeArchive(ctx, true)
}

func (m *Module) DeleteChallengeForever(ctx *tgctx.MsgContext, challengeID int64) {
	if err := m.challengesvc.DeleteChallengeForever(ctx.Ctx, ctx.DBUserID, challengeID); err != nil {
		log.Error().Err(err).Msg("delete challenge forever failed")
	}
	m.ShowChallengeArchive(ctx, true)
}

func (m *Module) sendOrEditChallenge(ctx *tgctx.MsgContext, edit bool, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	if edit && ctx.MessageID > 0 {
		var out tgbotapi.Chattable
		if menu != nil {
			e := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, *menu)
			e.ParseMode = "Markdown"
			out = e
		} else {
			e := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
			e.ParseMode = "Markdown"
			out = e
		}
		if _, err := m.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("edit challenge screen failed, sending fresh message instead")
			m.sendOrEditChallenge(ctx, false, text, menu)
		}
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	if menu != nil {
		msg.ReplyMarkup = *menu
	}
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send challenge screen failed")
	}
}

// skips sending (no error) if the day's already marked via the grid, but still reschedules tomorrow's push.
func (m *Module) SendChallengePush(ctx context.Context, chatID, dbUserID, challengeID int64, challengeName string, startDate, endDate time.Time) error {
	loc := apptime.Location
	if stats, err := m.profilesvc.GetProfileStats(ctx, chatID); err == nil && stats.TimeZone != nil {
		loc = apptime.Resolve(*stats.TimeZone)
	}
	now := apptime.NowIn(loc)
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	defer func() {
		if err := m.challengesvc.AdvancePush(ctx, challengeID, now, endDate, loc); err != nil {
			log.Error().Err(err).Int64("challenge_id", challengeID).Msg("advance challenge push schedule failed")
		}
	}()

	status, err := m.challengesvc.GetDayStatus(ctx, dbUserID, challengeID, todayMidnight)
	if err != nil {
		return err
	}
	if status != models.ChallengeDayPending {
		return nil
	}

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	dayNum := int(todayMidnight.Sub(startDate).Hours()/24) + 1

	lang := i18n.Default
	if stats, err := m.profilesvc.GetProfileStats(ctx, chatID); err == nil && stats.Language != nil {
		lang = i18n.Normalize(*stats.Language)
	}
	msg := tgbotapi.NewMessage(chatID, challenge.PushText(lang, challengeName, dayNum, totalDays))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = challenge.PushInlineMenu(lang, challengeID)
	_, err = m.bot.Send(msg)
	return err
}

func (m *Module) RecordChallengePushAnswer(ctx *tgctx.MsgContext, challengeID int64, done bool) {
	today := apptime.NowIn(ctx.Location)
	todayMidnight := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if err := m.challengesvc.MarkDay(ctx.Ctx, ctx.DBUserID, challengeID, todayMidnight, done); err != nil {
		log.Error().Err(err).Msg("record challenge push answer failed")
		return
	}
	resultKey := i18n.KeyChallengePushMarkedSkipped
	if done {
		resultKey = i18n.KeyChallengePushMarkedDone
	}
	if ctx.MessageID > 0 {
		edit := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, i18n.T(ctx.Language, resultKey))
		_, _ = m.bot.Send(edit)
	}
}

// ShowRoadmapMenu loads roadmap stats and renders the Roadmap root screen.
func (m *Module) ShowRoadmapMenu(ctx *tgctx.MsgContext) {
	stats, err := m.roadmapsvc.GetRoadmapStats(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("GetRoadmapStats failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLoadFailed)))
		return
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, roadmap.RoadmapMenuText(ctx.Language, stats))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = roadmap.RoadmapEntryInlineMenu(ctx.Language, stats.TimerActive, m.hasRoadmapOrphans(ctx))

	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send roadmap menu failed")
	}
}

// The "without a goal" entry only earns a slot on the root screen when
// something is actually unattached — otherwise it is a dead end.
func (m *Module) hasRoadmapOrphans(ctx *tgctx.MsgContext) bool {
	orphans, err := m.roadmapsvc.ListOrphanRoadmaps(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list orphan roadmaps failed")
		return false
	}
	return len(orphans) > 0
}

// sendOrEditRoadmap sends a fresh message or edits the current inline
// message — shared by every Roadmap sub-screen (copy of sendOrEditLearning).
func (m *Module) sendOrEditRoadmap(ctx *tgctx.MsgContext, edit bool, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	if edit && ctx.MessageID > 0 {
		var out tgbotapi.Chattable
		if menu != nil {
			e := tgbotapi.NewEditMessageTextAndMarkup(ctx.ChatID, ctx.MessageID, text, *menu)
			e.ParseMode = "Markdown"
			out = e
		} else {
			e := tgbotapi.NewEditMessageText(ctx.ChatID, ctx.MessageID, text)
			e.ParseMode = "Markdown"
			out = e
		}
		if _, err := m.bot.Send(out); err != nil {
			log.Error().Err(err).Msg("edit roadmap screen failed, sending fresh message instead")
			m.sendOrEditRoadmap(ctx, false, text, menu)
		}
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, text)
	msg.ParseMode = "Markdown"
	if menu != nil {
		msg.ReplyMarkup = *menu
	}
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send roadmap screen failed, retrying as plain text")
		m.sendRoadmapPlain(ctx.ChatID, text, menu)
	}
}

// sendRoadmapPlain re-sends a screen with no ParseMode. Roadmap screens embed
// user-authored text (card lines, criteria, technology names) in a Markdown
// message, and that text is routinely pasted technical prose — "snake_case",
// "sync.Mutex", "SELECT *". An unbalanced _ or * makes Telegram reject the
// whole message, which would otherwise mean the screen (or a reminder digest)
// silently never arrives. Dropping the formatting is a far better failure
// mode than dropping the message, and it beats escaping the user's own text
// into something they didn't type.
func (m *Module) sendRoadmapPlain(chatID int64, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if menu != nil {
		msg.ReplyMarkup = *menu
	}
	if _, err := m.bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("send roadmap screen as plain text failed too")
	}
}

// ShowRoadmapGoals lists the user's goals.
func (m *Module) ShowRoadmapGoals(ctx *tgctx.MsgContext, edit bool) {
	goals, err := m.roadmapsvc.ListGoals(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list goals failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapGoalsLoadFailed), nil)
		return
	}
	if len(goals) == 0 {
		menu := roadmap.RoadmapGoalsInlineMenu(ctx.Language, goals)
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapGoalsEmpty), &menu)
		return
	}
	menu := roadmap.RoadmapGoalsInlineMenu(ctx.Language, goals)
	m.sendOrEditRoadmap(ctx, edit, roadmap.RoadmapGoalsTitle(ctx.Language, len(goals)), &menu)
}

// ShowRoadmapGoalDetail renders one goal with the technologies under it.
func (m *Module) ShowRoadmapGoalDetail(ctx *tgctx.MsgContext, goalID int64, edit bool) {
	goal, err := m.roadmapsvc.Goal(ctx.Ctx, ctx.DBUserID, goalID)
	if err != nil {
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapGoalNotFound), nil)
		return
	}
	items, err := m.roadmapsvc.ListRoadmaps(ctx.Ctx, ctx.DBUserID, goalID)
	if err != nil {
		log.Error().Err(err).Msg("list roadmaps in goal failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapLoadFailed), nil)
		return
	}
	menu := roadmap.RoadmapGoalDetailInlineMenu(ctx.Language, goalID, items)
	m.sendOrEditRoadmap(ctx, edit, roadmap.RoadmapGoalDetailText(ctx.Language, goal), &menu)
}

// PromptCreateGoal asks for a goal name, checking the cap up front so the
// user learns they're full before typing rather than after. The service
// re-checks on insert regardless.
func (m *Module) PromptCreateGoal(ctx *tgctx.MsgContext) (ok bool) {
	goals, err := m.roadmapsvc.ListGoals(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list goals failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalsLoadFailed)))
		return false
	}
	if len(goals) >= models.MaxRoadmapGoalsPerUser {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalLimitReached, models.MaxRoadmapGoalsPerUser)))
		return false
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalCreatePrompt))
	msg.ReplyMarkup = roadmap.RoadmapWaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
	return true
}

// ProcessCreateGoalName creates a goal from typed text. done is true once the
// waiting state should be cleared (on both success and unrecoverable input,
// matching ProcessCreateCollectionName's contract).
func (m *Module) ProcessCreateGoalName(ctx *tgctx.MsgContext) (goalID int64, done bool) {
	name := strings.TrimSpace(ctx.Text)

	id, err := m.roadmapsvc.CreateGoal(ctx.Ctx, ctx.DBUserID, name)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRoadmapInvalidName):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonNameSingleLineInvalid)))
			return 0, false
		case errors.Is(err, models.ErrRoadmapGoalExists):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalExists)))
			return 0, false
		case errors.Is(err, models.ErrRoadmapGoalLimitReached):
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalLimitReached, models.MaxRoadmapGoalsPerUser))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = m.bot.Send(hide)
			return 0, true
		}
		log.Error().Err(err).Msg("create goal failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalCreateFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		return 0, true
	}

	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalCreated, name))
	confirm.ParseMode = "Markdown"
	confirm.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(confirm)
	m.ShowRoadmapGoalDetail(ctx, id, false)
	return id, true
}

func (m *Module) PromptRenameGoal(ctx *tgctx.MsgContext, goalID int64) {
	goal, err := m.roadmapsvc.Goal(ctx.Ctx, ctx.DBUserID, goalID)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalNotFound)))
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, roadmap.RoadmapGoalRenamePromptText(ctx.Language, goal.Name))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = roadmap.RoadmapWaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ProcessRenameGoal(ctx *tgctx.MsgContext, goalID int64) (done bool) {
	name := strings.TrimSpace(ctx.Text)

	if err := m.roadmapsvc.RenameGoal(ctx.Ctx, ctx.DBUserID, goalID, name); err != nil {
		switch {
		case errors.Is(err, models.ErrRoadmapInvalidName):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonNameSingleLineInvalid)))
			return false
		case errors.Is(err, models.ErrRoadmapGoalExists):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalExists)))
			return false
		}
		log.Error().Err(err).Msg("rename goal failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalRenameFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		m.ShowRoadmapGoalDetail(ctx, goalID, false)
		return true
	}

	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalRenamed, name))
	hide.ParseMode = "Markdown"
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	m.ShowRoadmapGoalDetail(ctx, goalID, false)
	return true
}

// HandleGoalArchive archives a goal and returns to the goal list.
func (m *Module) HandleGoalArchive(ctx *tgctx.MsgContext, goalID int64) {
	if err := m.roadmapsvc.ArchiveGoal(ctx.Ctx, ctx.DBUserID, goalID); err != nil {
		log.Error().Err(err).Msg("archive goal failed")
	}
	m.ShowRoadmapGoals(ctx, true)
}

// PromptCreateRoadmap asks for a technology name inside a goal, checking that
// goal's cap up front.
func (m *Module) PromptCreateRoadmap(ctx *tgctx.MsgContext, goalID int64) (ok bool) {
	items, err := m.roadmapsvc.ListRoadmaps(ctx.Ctx, ctx.DBUserID, goalID)
	if err != nil {
		log.Error().Err(err).Msg("list roadmaps in goal failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLoadFailed)))
		return false
	}
	if len(items) >= models.MaxRoadmapsPerGoal {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLimitReached, models.MaxRoadmapsPerGoal)))
		return false
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCreatePrompt))
	msg.ReplyMarkup = roadmap.RoadmapWaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
	return true
}

// ProcessCreateRoadmapName creates a technology inside goalID and chains
// straight into the criteria prompt.
func (m *Module) ProcessCreateRoadmapName(ctx *tgctx.MsgContext, goalID int64) (roadmapID int64, done bool) {
	name := strings.TrimSpace(ctx.Text)

	id, err := m.roadmapsvc.CreateRoadmap(ctx.Ctx, ctx.DBUserID, goalID, name)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRoadmapInvalidName):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonNameSingleLineInvalid)))
			return 0, false
		case errors.Is(err, models.ErrRoadmapExists):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCreateExists)))
			return 0, false
		case errors.Is(err, models.ErrRoadmapLimitReached):
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLimitReached, models.MaxRoadmapsPerGoal))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = m.bot.Send(hide)
			return 0, true
		case errors.Is(err, models.ErrRoadmapGoalNotFound):
			hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalNotFound))
			hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = m.bot.Send(hide)
			return 0, true
		}
		log.Error().Err(err).Msg("create roadmap failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCreateFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		return 0, true
	}

	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCreateConfirmed, name))
	confirm.ParseMode = "Markdown"
	_, _ = m.bot.Send(confirm)
	m.PromptRoadmapCriteria(ctx, id)
	return id, true
}

// ShowRoadmapDetail renders one technology's card checklist and actions.
func (m *Module) ShowRoadmapDetail(ctx *tgctx.MsgContext, roadmapID int64, edit bool) {
	item, err := m.roadmapsvc.Roadmap(ctx.Ctx, ctx.DBUserID, roadmapID)
	if err != nil {
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapNotFound), nil)
		return
	}
	cards, err := m.roadmapsvc.ListCards(ctx.Ctx, ctx.DBUserID, roadmapID)
	if err != nil {
		log.Error().Err(err).Msg("list roadmap cards failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapCardsLoadFailed), nil)
		return
	}

	menu := roadmap.RoadmapDetailInlineMenu(ctx.Language, item, cards, m.roadmapaisvc.Enabled())
	m.sendOrEditRoadmap(ctx, edit, roadmap.RoadmapDetailText(ctx.Language, item, len(cards)), &menu)
}

// ShowRoadmapOrphans lists technologies attached to no goal — v1 leftovers,
// or ones whose goal was deleted.
func (m *Module) ShowRoadmapOrphans(ctx *tgctx.MsgContext, edit bool) {
	items, err := m.roadmapsvc.ListOrphanRoadmaps(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list orphan roadmaps failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapLoadFailed), nil)
		return
	}
	if len(items) == 0 {
		menu := roadmap.RoadmapBackToMainInlineMenu(ctx.Language)
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapOrphansEmpty), &menu)
		return
	}
	menu := roadmap.RoadmapOrphansInlineMenu(ctx.Language, items)
	m.sendOrEditRoadmap(ctx, edit, roadmap.RoadmapOrphansTitle(ctx.Language, len(items)), &menu)
}

// ShowRoadmapAssignGoal offers the goals an unattached technology can move
// into.
func (m *Module) ShowRoadmapAssignGoal(ctx *tgctx.MsgContext, roadmapID int64) {
	item, err := m.roadmapsvc.Roadmap(ctx.Ctx, ctx.DBUserID, roadmapID)
	if err != nil {
		m.sendOrEditRoadmap(ctx, true, i18n.T(ctx.Language, i18n.KeyRoadmapNotFound), nil)
		return
	}
	goals, err := m.roadmapsvc.ListGoals(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list goals failed")
		m.sendOrEditRoadmap(ctx, true, i18n.T(ctx.Language, i18n.KeyRoadmapGoalsLoadFailed), nil)
		return
	}
	if len(goals) == 0 {
		menu := roadmap.RoadmapBackToMainInlineMenu(ctx.Language)
		m.sendOrEditRoadmap(ctx, true, i18n.T(ctx.Language, i18n.KeyRoadmapAssignNoGoals), &menu)
		return
	}
	menu := roadmap.RoadmapAssignGoalInlineMenu(ctx.Language, roadmapID, goals)
	m.sendOrEditRoadmap(ctx, true, roadmap.RoadmapAssignPromptText(ctx.Language, item.Name), &menu)
}

// HandleRoadmapAssign moves a technology into a goal and lands the user on
// that goal, which is where the technology now lives.
func (m *Module) HandleRoadmapAssign(ctx *tgctx.MsgContext, roadmapID, goalID int64) {
	if err := m.roadmapsvc.AssignRoadmapToGoal(ctx.Ctx, ctx.DBUserID, roadmapID, goalID); err != nil {
		if errors.Is(err, models.ErrRoadmapLimitReached) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLimitReached, models.MaxRoadmapsPerGoal)))
			return
		}
		log.Error().Err(err).Msg("assign roadmap to goal failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAssignFailed)))
		return
	}
	m.ShowRoadmapGoalDetail(ctx, goalID, true)
}

// PromptRoadmapCriteria asks what "I know this" means for the user. The
// criteria is optional, so the keyboard offers Skip alongside Cancel.
func (m *Module) PromptRoadmapCriteria(ctx *tgctx.MsgContext, roadmapID int64) {
	name, err := m.roadmapsvc.RoadmapName(ctx.Ctx, ctx.DBUserID, roadmapID)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapNotFound)))
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, roadmap.RoadmapCriteriaPromptText(ctx.Language, name))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = roadmap.RoadmapCriteriaReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ProcessRoadmapCriteria(ctx *tgctx.MsgContext, roadmapID int64) (done bool) {
	if err := m.roadmapsvc.SetMasteryCriteria(ctx.Ctx, ctx.DBUserID, roadmapID, ctx.Text); err != nil {
		if errors.Is(err, models.ErrRoadmapCriteriaTooLong) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCriteriaTooLong, models.MaxRoadmapCriteriaLen)))
			return false
		}
		log.Error().Err(err).Msg("set mastery criteria failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCriteriaFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		return true
	}

	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCriteriaSaved))
	confirm.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(confirm)
	return true
}

func (m *Module) NoticeRoadmapCriteriaSkipped(ctx *tgctx.MsgContext) {
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCriteriaSkipped))
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(msg)
}

func (m *Module) PromptRenameRoadmap(ctx *tgctx.MsgContext, roadmapID int64) {
	name, err := m.roadmapsvc.RoadmapName(ctx.Ctx, ctx.DBUserID, roadmapID)
	if err != nil {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapNotFound)))
		return
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, roadmap.RoadmapRenamePromptText(ctx.Language, name))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = roadmap.RoadmapWaitingReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

func (m *Module) ProcessRenameRoadmap(ctx *tgctx.MsgContext, roadmapID int64) (done bool) {
	name := strings.TrimSpace(ctx.Text)

	if err := m.roadmapsvc.RenameRoadmap(ctx.Ctx, ctx.DBUserID, roadmapID, name); err != nil {
		switch {
		case errors.Is(err, models.ErrRoadmapInvalidName):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyCommonNameSingleLineInvalid)))
			return false
		case errors.Is(err, models.ErrRoadmapExists):
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapCreateExists)))
			return false
		}
		log.Error().Err(err).Msg("rename roadmap failed")
		hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapRenameFailed))
		hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = m.bot.Send(hide)
		m.ShowRoadmapDetail(ctx, roadmapID, false)
		return true
	}

	hide := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapRenamed, name))
	hide.ParseMode = "Markdown"
	hide.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(hide)
	m.ShowRoadmapDetail(ctx, roadmapID, false)
	return true
}

// PromptAddRoadmapCards asks the user to paste checklist lines. first is true
// right after a technology is created (slightly different copy).
func (m *Module) PromptAddRoadmapCards(ctx *tgctx.MsgContext, roadmapID int64, first bool) {
	key := i18n.KeyRoadmapAddCardsPromptMore
	if first {
		key = i18n.KeyRoadmapAddCardsPromptFirst
	}
	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, key))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = roadmap.RoadmapAddCardsReplyMenu(ctx.Language)
	_, _ = m.bot.Send(msg)
}

// ProcessAddRoadmapCards turns pasted lines into cards. The caller keeps the
// waiting state active afterward — the user may paste more, and only exits
// via the "Done" button.
func (m *Module) ProcessAddRoadmapCards(ctx *tgctx.MsgContext, roadmapID int64) {
	added, skipped, err := m.roadmapsvc.AddCardsFromText(ctx.Ctx, ctx.DBUserID, roadmapID, ctx.Text)
	if err != nil {
		if errors.Is(err, models.ErrRoadmapNoCardsParsed) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAddCardsNoneParsed)))
			return
		}
		log.Error().Err(err).Msg("add roadmap cards failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapAddCardsFailed)))
		return
	}

	text := i18n.T(ctx.Language, i18n.KeyRoadmapAddCardsAdded, added)
	if skipped > 0 {
		text += i18n.T(ctx.Language, i18n.KeyRoadmapAddCardsSkipped, skipped)
	}
	_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, text))
}

// HandleRoadmapCardToggle ticks/unticks one card and re-renders its
// technology's checklist.
func (m *Module) HandleRoadmapCardToggle(ctx *tgctx.MsgContext, cardID int64) {
	roadmapID, err := m.roadmapsvc.ToggleCardDone(ctx.Ctx, ctx.DBUserID, cardID)
	if err != nil {
		log.Error().Err(err).Msg("toggle roadmap card failed")
		return
	}
	m.ShowRoadmapDetail(ctx, roadmapID, true)
}

// HandleRoadmapCardDifficulty cycles one card's difficulty, which also
// re-sorts the checklist (it is ordered easiest-first).
func (m *Module) HandleRoadmapCardDifficulty(ctx *tgctx.MsgContext, cardID int64) {
	roadmapID, err := m.roadmapsvc.CycleCardDifficulty(ctx.Ctx, ctx.DBUserID, cardID)
	if err != nil {
		log.Error().Err(err).Msg("cycle roadmap card difficulty failed")
		return
	}
	m.ShowRoadmapDetail(ctx, roadmapID, true)
}

// HandleRoadmapDigestToggle ticks a card straight off a reminder push and
// re-renders that same push with whatever is still pending — so the
// notification stays a live checklist instead of going stale.
func (m *Module) HandleRoadmapDigestToggle(ctx *tgctx.MsgContext, cardID int64) {
	if _, err := m.roadmapsvc.ToggleCardDone(ctx.Ctx, ctx.DBUserID, cardID); err != nil {
		log.Error().Err(err).Msg("toggle roadmap card from digest failed")
		return
	}

	// No AI advice on a re-render: this runs in the dispatcher's serial
	// update loop, and the note would also change under the user on every
	// tick. The advice is written once, on the scheduled push.
	text, menu, cards, err := m.buildRoadmapDigest(ctx.Ctx, ctx.DBUserID, ctx.Language)
	if err != nil {
		log.Error().Err(err).Msg("rebuild roadmap digest failed")
		return
	}
	if len(cards) == 0 {
		m.sendOrEditRoadmap(ctx, true, i18n.T(ctx.Language, i18n.KeyRoadmapDigestEmpty), nil)
		return
	}
	m.sendOrEditRoadmap(ctx, true, text, &menu)
}

func (m *Module) HandleRoadmapCardDelete(ctx *tgctx.MsgContext, cardID, roadmapID int64) {
	if err := m.roadmapsvc.DeleteCard(ctx.Ctx, ctx.DBUserID, cardID); err != nil {
		log.Error().Err(err).Msg("delete roadmap card failed")
	}
	m.ShowRoadmapDetail(ctx, roadmapID, true)
}

// HandleRoadmapToggle flips whether a technology feeds reminder digests.
func (m *Module) HandleRoadmapToggle(ctx *tgctx.MsgContext, roadmapID int64) {
	if err := m.roadmapsvc.ToggleRoadmapActive(ctx.Ctx, ctx.DBUserID, roadmapID); err != nil {
		log.Error().Err(err).Msg("toggle roadmap active failed")
	}
	m.ShowRoadmapDetail(ctx, roadmapID, true)
}

// HandleRoadmapArchive archives a technology and returns to wherever it
// lived: its goal, or the orphan list.
func (m *Module) HandleRoadmapArchive(ctx *tgctx.MsgContext, roadmapID int64) {
	item, err := m.roadmapsvc.Roadmap(ctx.Ctx, ctx.DBUserID, roadmapID)
	if err != nil {
		m.sendOrEditRoadmap(ctx, true, i18n.T(ctx.Language, i18n.KeyRoadmapNotFound), nil)
		return
	}
	goalID := item.GoalID
	if err := m.roadmapsvc.ArchiveRoadmap(ctx.Ctx, ctx.DBUserID, roadmapID); err != nil {
		log.Error().Err(err).Msg("archive roadmap failed")
	}
	if goalID != nil {
		m.ShowRoadmapGoalDetail(ctx, *goalID, true)
		return
	}
	m.ShowRoadmapOrphans(ctx, true)
}

// ShowRoadmapArchiveMenu lists archived goals and technologies together.
func (m *Module) ShowRoadmapArchiveMenu(ctx *tgctx.MsgContext, edit bool) {
	goals, err := m.roadmapsvc.ListArchivedGoals(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list archived goals failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyTrackArchiveLoadFailed), nil)
		return
	}
	items, err := m.roadmapsvc.ListArchivedRoadmaps(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("list archived roadmaps failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyTrackArchiveLoadFailed), nil)
		return
	}
	if len(goals) == 0 && len(items) == 0 {
		menu := roadmap.RoadmapBackToMainInlineMenu(ctx.Language)
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapArchiveEmpty), &menu)
		return
	}
	menu := roadmap.RoadmapArchiveInlineMenu(ctx.Language, goals, items)
	m.sendOrEditRoadmap(ctx, edit, roadmap.RoadmapArchiveText(ctx.Language, goals, items), &menu)
}

// Restoring re-occupies a slot, so a full list is reported rather than
// silently doing nothing.
func (m *Module) RestoreArchivedRoadmap(ctx *tgctx.MsgContext, roadmapID int64) {
	if err := m.roadmapsvc.RestoreRoadmap(ctx.Ctx, ctx.DBUserID, roadmapID); err != nil {
		if errors.Is(err, models.ErrRoadmapLimitReached) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLimitReached, models.MaxRoadmapsPerGoal)))
			return
		}
		log.Error().Err(err).Msg("restore roadmap failed")
	}
	m.ShowRoadmapArchiveMenu(ctx, true)
}

func (m *Module) DeleteArchivedRoadmapForever(ctx *tgctx.MsgContext, roadmapID int64) {
	if err := m.roadmapsvc.DeleteRoadmapForever(ctx.Ctx, ctx.DBUserID, roadmapID); err != nil {
		log.Error().Err(err).Msg("delete roadmap forever failed")
	}
	m.ShowRoadmapArchiveMenu(ctx, true)
}

func (m *Module) RestoreArchivedGoal(ctx *tgctx.MsgContext, goalID int64) {
	if err := m.roadmapsvc.RestoreGoal(ctx.Ctx, ctx.DBUserID, goalID); err != nil {
		if errors.Is(err, models.ErrRoadmapGoalLimitReached) {
			_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapGoalLimitReached, models.MaxRoadmapGoalsPerUser)))
			return
		}
		log.Error().Err(err).Msg("restore goal failed")
	}
	m.ShowRoadmapArchiveMenu(ctx, true)
}

// Deleting a goal leaves its technologies behind as unattached (ON DELETE SET
// NULL), so this never destroys a plan by itself.
func (m *Module) DeleteArchivedGoalForever(ctx *tgctx.MsgContext, goalID int64) {
	if err := m.roadmapsvc.DeleteGoalForever(ctx.Ctx, ctx.DBUserID, goalID); err != nil {
		log.Error().Err(err).Msg("delete goal forever failed")
	}
	m.ShowRoadmapArchiveMenu(ctx, true)
}

// ShowRoadmapStatsDetail renders the full progress breakdown.
func (m *Module) ShowRoadmapStatsDetail(ctx *tgctx.MsgContext, edit bool) {
	detail, err := m.roadmapsvc.GetStatsDetail(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("get roadmap stats detail failed")
		m.sendOrEditRoadmap(ctx, edit, i18n.T(ctx.Language, i18n.KeyRoadmapStatsLoadFailed), nil)
		return
	}
	menu := roadmap.RoadmapBackToMainInlineMenu(ctx.Language)
	m.sendOrEditRoadmap(ctx, edit, roadmap.RoadmapStatsDetailText(ctx.Language, detail), &menu)
}

// ShowRoadmapIntervalPicker shows the reminder-interval picker. ok is false
// when there is nothing to remind about yet — activating then would only send
// empty digests, so the user is told instead.
func (m *Module) ShowRoadmapIntervalPicker(ctx *tgctx.MsgContext) (ok bool) {
	cards, err := m.roadmapsvc.PickDigestCards(ctx.Ctx, ctx.DBUserID)
	if err != nil {
		log.Error().Err(err).Msg("pick roadmap digest cards failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapLoadFailed)))
		return false
	}
	if len(cards) == 0 {
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapPushNeedOne)))
		return false
	}

	msg := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapPushIntervalPrompt))
	msg.ReplyMarkup = roadmap.RoadmapPushIntervalReplyMenu(ctx.Language, roadmap.BuiltInPushIntervals)
	_, _ = m.bot.Send(msg)
	return true
}

func (m *Module) ActivateRoadmapReminders(ctx *tgctx.MsgContext, intervalMin int) {
	if err := m.roadmapsvc.Activate(ctx.Ctx, ctx.DBUserID, intervalMin); err != nil {
		log.Error().Err(err).Msg("activate roadmap reminders failed")
		_, _ = m.bot.Send(tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapPushActivateFailed)))
		return
	}
	// Clear the interval-picker keyboard on the confirmation message itself,
	// same reasoning as ActivateReviews.
	confirm := tgbotapi.NewMessage(ctx.ChatID, i18n.T(ctx.Language, i18n.KeyRoadmapPushActivated, intervalMin))
	confirm.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = m.bot.Send(confirm)
	m.ShowRoadmapMenu(ctx)
}

func (m *Module) StopRoadmapReminders(ctx *tgctx.MsgContext) {
	if err := m.roadmapsvc.Stop(ctx.Ctx, ctx.DBUserID); err != nil {
		log.Error().Err(err).Msg("stop roadmap reminders failed")
	}
	m.ShowRoadmapMenu(ctx)
}

// buildRoadmapDigest assembles a reminder digest: the pending-card text plus
// a keyboard with one tick button per card. hasCards is false when nothing is
// pending, in which case there is nothing worth sending.
func (m *Module) buildRoadmapDigest(ctx context.Context, userID int64, lang i18n.Lang) (string, tgbotapi.InlineKeyboardMarkup, []models.RoadmapDigestCard, error) {
	cards, err := m.roadmapsvc.PickDigestCards(ctx, userID)
	if err != nil {
		return "", tgbotapi.InlineKeyboardMarkup{}, nil, err
	}
	if len(cards) == 0 {
		return "", tgbotapi.InlineKeyboardMarkup{}, nil, nil
	}

	// Per-technology done/total counts for the digest's group headers — the
	// digest query itself only returns the pending cards.
	byRoadmap := make(map[int64]models.RoadmapItem)
	for _, c := range cards {
		if _, ok := byRoadmap[c.RoadmapID]; ok {
			continue
		}
		item, err := m.roadmapsvc.Roadmap(ctx, userID, c.RoadmapID)
		if err != nil {
			return "", tgbotapi.InlineKeyboardMarkup{}, nil, err
		}
		byRoadmap[c.RoadmapID] = item
	}

	return roadmap.RoadmapDigestText(lang, cards, byRoadmap), roadmap.RoadmapDigestInlineMenu(lang, cards), cards, nil
}

// SendRoadmapDigestMessage sends one reminder digest of the easiest pending
// cards across the user's active technologies. Called off the scheduler, not
// a live user request — mirrors SendLearningPromptMessage's language lookup.
//
// With nothing pending it returns nil without sending; the scheduler still
// advances the schedule afterward, so an all-done user isn't retried on every
// tick.
func (m *Module) SendRoadmapDigestMessage(ctx context.Context, chatID int64, userID int64) error {
	lang := i18n.Default
	if stats, err := m.profilesvc.GetProfileStats(ctx, chatID); err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("load language for roadmap digest failed")
	} else if stats.Language != nil {
		lang = i18n.Normalize(*stats.Language)
	}

	text, menu, cards, err := m.buildRoadmapDigest(ctx, userID, lang)
	if err != nil {
		return err
	}
	if len(cards) == 0 {
		return nil
	}

	// The advice is a bonus on top of the digest, so a failed or disabled
	// provider costs the note and nothing else — the reminder still goes out.
	// Safe to call inline: the scheduler runs on its own goroutine, unlike
	// the dispatcher.
	if m.roadmapaisvc.Enabled() {
		if advice, err := m.roadmapaisvc.DigestAdvice(ctx, userID, cards, string(lang)); err != nil {
			log.Warn().Err(err).Int64("user_id", userID).Msg("roadmap digest advice failed, sending the digest without it")
		} else {
			text += i18n.T(lang, i18n.KeyRoadmapAIDigestHintFmt, advice)
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = menu
	if _, err := m.bot.Send(msg); err != nil {
		// Same Markdown-vs-pasted-text hazard sendRoadmapPlain covers, and
		// here it matters more: a rejected digest is a reminder the user
		// simply never gets, with the schedule advancing regardless.
		log.Error().Err(err).Int64("user_id", userID).Msg("send roadmap digest failed, retrying as plain text")
		plain := tgbotapi.NewMessage(chatID, text)
		plain.ReplyMarkup = menu
		_, err = m.bot.Send(plain)
		return err
	}
	return nil
}
