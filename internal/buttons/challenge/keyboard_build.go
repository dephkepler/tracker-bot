package challenge

import (
	"fmt"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WaitingReplyMenu is shown while the bot is waiting for typed input (a new
// challenge's name) — lets the user bail out.
func WaitingReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyCancel)),
	)
}

// ListInlineMenu lists a user's active challenges — tap one to open its
// grid (see GridInlineMenu).
func ListInlineMenu(items []models.ChallengeItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+3)
	for _, item := range items {
		pct := 0
		if item.TotalDays > 0 {
			pct = item.DoneDays * 100 / item.TotalDays
		}
		label := fmt.Sprintf("🎯 %s — %d%% (%d/%d)", item.Name, pct, item.DoneDays, item.TotalDays)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, fmt.Sprintf("%s%d", CBOpen, item.ID))))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ButtonCreate, CBCreate)))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ButtonArchive, CBArchiveOpen)))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ButtonHome, "go_home")))
	return buttonbuilder.IK(rows...)
}

// GridInlineMenu renders one challenge's day-squares: ✅ done, ❌ skipped,
// 🔲 pending-and-markable (today or already past), ⬜ upcoming (inert).
// Seven days per row, in range order (not calendar-week-aligned).
func GridInlineMenu(challengeID int64, days []models.ChallengeDay, today time.Time) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(days)/7+3)
	var row []tgbotapi.InlineKeyboardButton
	for _, d := range days {
		emoji, clickable := dayCellState(d, today)
		cb := "noop"
		if clickable {
			cb = fmt.Sprintf("%s%d:%s", CBDayOpen, challengeID, d.Date.Format("2006-01-02"))
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(emoji, cb))
		if len(row) == 7 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ButtonArchive+" this", fmt.Sprintf("%s%d", CBArchiveThis, challengeID))))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ButtonBack, CBBackList)))
	return buttonbuilder.IK(rows...)
}

// dayCellState returns the emoji for one day and whether it should be
// clickable (today or earlier, within the challenge's range — future days
// are always inert).
func dayCellState(d models.ChallengeDay, today time.Time) (emoji string, clickable bool) {
	switch d.Status {
	case models.ChallengeDayDone:
		return "✅", !d.Date.After(today)
	case models.ChallengeDaySkipped:
		return "❌", !d.Date.After(today)
	default:
		if d.Date.After(today) {
			return "⬜", false
		}
		return "🔲", true
	}
}

// DayConfirmInlineMenu lets the user mark one specific day.
func DayConfirmInlineMenu(challengeID int64, day time.Time) tgbotapi.InlineKeyboardMarkup {
	dateStr := day.Format("2006-01-02")
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(ButtonMarkDone, fmt.Sprintf("%s%d:%s", CBDayDone, challengeID, dateStr)),
			buttonbuilder.IB(ButtonMarkSkip, fmt.Sprintf("%s%d:%s", CBDaySkip, challengeID, dateStr)),
		),
		buttonbuilder.IR(buttonbuilder.IB(ButtonBack, fmt.Sprintf("%s%d", CBOpen, challengeID))),
	)
}

// ArchiveInlineMenu lists archived challenges with restore/delete actions —
// mirrors learning.LearningArchiveInlineMenu.
func ArchiveInlineMenu(items []models.ChallengeItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)*2+1)
	for _, item := range items {
		label := fmt.Sprintf("📦 %s — %d/%d done", item.Name, item.DoneDays, item.TotalDays)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, "noop")))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(ButtonRestore, fmt.Sprintf("%s%d", CBArchiveRestore, item.ID)),
			buttonbuilder.IB(ButtonDeleteForever, fmt.Sprintf("%s%d", CBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ButtonBack, CBBackList)))
	return buttonbuilder.IK(rows...)
}

// PushInlineMenu is attached to the daily evening push message.
func PushInlineMenu(challengeID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(ButtonMarkDone, fmt.Sprintf("%s%d", CBPushDone, challengeID)),
			buttonbuilder.IB(ButtonMarkSkip, fmt.Sprintf("%s%d", CBPushSkip, challengeID)),
		),
	)
}

// --- calendar range-picker (creation flow) ---------------------------------

var monthNamesEN = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
var weekdayShortEN = [...]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// CalendarInlineMenu renders a month calendar for picking a challenge's
// start/end date — tap twice (start, then end); mirrors the shape of
// track.TrackReportPeriodCalendarInlineMenu but kept as an independent,
// smaller copy so this module doesn't reach into Track's callback space.
func CalendarInlineMenu(month time.Time, from, to time.Time) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 10)
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	startPad := (int(first.Weekday()) + 6) % 7

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("«Y", CBCalPrevYear),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %d", monthNamesEN[first.Month()-1], first.Year()), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("Y»", CBCalNextYear),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀", CBCalPrev),
		tgbotapi.NewInlineKeyboardButtonData("Month", "noop"),
		tgbotapi.NewInlineKeyboardButtonData("▶", CBCalNext),
	))
	weekHeader := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	for _, w := range weekdayShortEN {
		weekHeader = append(weekHeader, tgbotapi.NewInlineKeyboardButtonData(w, "noop"))
	}
	rows = append(rows, weekHeader)

	day := 1
	for week := 0; week < 6; week++ {
		row := make([]tgbotapi.InlineKeyboardButton, 0, 7)
		for wd := 0; wd < 7; wd++ {
			cell := week*7 + wd
			if cell < startPad || day > last.Day() {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(" ", "noop"))
				continue
			}
			dt := time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, time.UTC)
			label := fmt.Sprintf("%2d", day)
			switch {
			case sameDay(dt, from):
				label = "🟢" + label
			case sameDay(dt, to):
				label = "🔵" + label
			case inRange(dt, from, to):
				label = "🟩" + label
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CBCalPick+dt.Format("2006-01-02")))
			day++
		}
		rows = append(rows, row)
		if day > last.Day() {
			break
		}
	}

	confirmLabel := ButtonSelectEnd
	confirmCB := "noop"
	if !from.IsZero() && !to.IsZero() {
		confirmLabel = ButtonConfirm
		confirmCB = CBCalDone
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(confirmLabel, confirmCB),
		tgbotapi.NewInlineKeyboardButtonData(ButtonCancel, CBCalCancel),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func sameDay(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func inRange(day, from, to time.Time) bool {
	if from.IsZero() || to.IsZero() {
		return false
	}
	if day.Before(from) || day.After(to) {
		return false
	}
	return !sameDay(day, from) && !sameDay(day, to)
}
