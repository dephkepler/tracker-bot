package challenge

import (
	"fmt"
	"time"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func WaitingReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancelX))),
	)
}

func ListInlineMenu(lang i18n.Lang, items []models.ChallengeItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+3)
	for _, item := range items {
		pct := 0
		if item.TotalDays > 0 {
			pct = item.DoneDays * 100 / item.TotalDays
		}
		label := i18n.T(lang, i18n.KeyChallengeItemLabelFmt, item.Name, pct, item.DoneDays, item.TotalDays)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, fmt.Sprintf("%s%d", CBOpen, item.ID))))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyChallengeButtonCreate), CBCreate)))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyChallengeButtonArchive), CBArchiveOpen)))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonHome), "go_home")))
	return buttonbuilder.IK(rows...)
}

// rows are 7 per range-day, not calendar-week-aligned
func GridInlineMenu(lang i18n.Lang, challengeID int64, days []models.ChallengeDay, today time.Time) tgbotapi.InlineKeyboardMarkup {
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
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyChallengeButtonArchiveThis), fmt.Sprintf("%s%d", CBArchiveThis, challengeID))))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), CBBackList)))
	return buttonbuilder.IK(rows...)
}

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

func DayConfirmInlineMenu(lang i18n.Lang, challengeID int64, day time.Time) tgbotapi.InlineKeyboardMarkup {
	dateStr := day.Format("2006-01-02")
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonDone), fmt.Sprintf("%s%d:%s", CBDayDone, challengeID, dateStr)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyChallengeButtonSkip), fmt.Sprintf("%s%d:%s", CBDaySkip, challengeID, dateStr)),
		),
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), fmt.Sprintf("%s%d", CBOpen, challengeID))),
	)
}

func ArchiveInlineMenu(lang i18n.Lang, items []models.ChallengeItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)*2+1)
	for _, item := range items {
		label := i18n.T(lang, i18n.KeyChallengeArchiveItemFmt, item.Name, item.DoneDays, item.TotalDays)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, "noop")))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelRestore), fmt.Sprintf("%s%d", CBArchiveRestore, item.ID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelDeleteForever), fmt.Sprintf("%s%d", CBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), CBBackList)))
	return buttonbuilder.IK(rows...)
}

func PushInlineMenu(lang i18n.Lang, challengeID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonDone), fmt.Sprintf("%s%d", CBPushDone, challengeID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyChallengeButtonSkip), fmt.Sprintf("%s%d", CBPushSkip, challengeID)),
		),
	)
}

// Mon(0)..Sun(6) and Jan(0)..Dec(11) — reuse Track's calendar translations
// rather than minting duplicate month/weekday keys (see keys.go's Challenge
// block comment). Each domain owns its own array since Track's is unexported.
var weekdayShortKeys = [...]string{
	i18n.KeyTrackCalendarMon, i18n.KeyTrackCalendarTue, i18n.KeyTrackCalendarWed,
	i18n.KeyTrackCalendarThu, i18n.KeyTrackCalendarFri, i18n.KeyTrackCalendarSat, i18n.KeyTrackCalendarSun,
}

var monthNameKeys = [...]string{
	i18n.KeyTrackCalendarMonth01, i18n.KeyTrackCalendarMonth02, i18n.KeyTrackCalendarMonth03,
	i18n.KeyTrackCalendarMonth04, i18n.KeyTrackCalendarMonth05, i18n.KeyTrackCalendarMonth06,
	i18n.KeyTrackCalendarMonth07, i18n.KeyTrackCalendarMonth08, i18n.KeyTrackCalendarMonth09,
	i18n.KeyTrackCalendarMonth10, i18n.KeyTrackCalendarMonth11, i18n.KeyTrackCalendarMonth12,
}

func monthName(lang i18n.Lang, m time.Month) string {
	return i18n.T(lang, monthNameKeys[m-1])
}

// deliberately a separate copy of track's calendar picker, to avoid reaching into Track's callback space
func CalendarInlineMenu(lang i18n.Lang, month time.Time, from, to time.Time) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 10)
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	startPad := (int(first.Weekday()) + 6) % 7

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("«Y", CBCalPrevYear),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %d", monthName(lang, first.Month()), first.Year()), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("Y»", CBCalNextYear),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀", CBCalPrev),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarMonth), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("▶", CBCalNext),
	))
	weekHeader := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	for _, k := range weekdayShortKeys {
		weekHeader = append(weekHeader, tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, k), "noop"))
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

	confirmLabel := i18n.T(lang, i18n.KeyTrackLabelSelectEndDate)
	confirmCB := "noop"
	if !from.IsZero() && !to.IsZero() {
		confirmLabel = i18n.T(lang, i18n.KeyTrackLabelConfirmRange)
		confirmCB = CBCalDone
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(confirmLabel, confirmCB),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyCommonCancelX), CBCalCancel),
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
