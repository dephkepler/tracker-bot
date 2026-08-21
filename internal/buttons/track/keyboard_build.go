package track

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TrackEntryInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackButtonSelectActivity), TrackCBActivitySelect),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackButtonCreateActivity), TrackCBActivityCreate),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackButtonViewReports), TrackCBReportSummary),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackButtonViewArchive), TrackCBArchiveOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonHome), "go_home"),
		),
	)
}

func TrackActivityManageReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonActivityActivate)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonActivityDelete)),
		),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonViewArchive))),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonHome)),
		),
	)
}

func TrackArchiveReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonSelectActivity)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonViewArchive)),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonHome)),
		),
	)
}

func TrackReportsReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonToday)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonCalendar)),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonHeatmap)),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonHome)),
		),
	)
}

// interval buttons are tapped to activate/delete (see FormatTimerButton/ParseTimerButtonMinutes), not via inline keyboard.
func TrackTimerReplyMenu(lang i18n.Lang, builtIn, custom []int) tgbotapi.ReplyKeyboardMarkup {
	rows := make([][]tgbotapi.KeyboardButton, 0, 4)
	rows = append(rows, timerIntervalRows(lang, builtIn, TrackTimerActivatePrefix)...)
	rows = append(rows, timerIntervalRows(lang, custom, TrackTimerActivatePrefix)...)

	if len(custom) > 0 {
		rows = append(rows, buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonTimerCreate)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonTimerDelete)),
		))
	} else {
		rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyTrackButtonTimerCreate))))
	}
	rows = append(rows, buttonbuilder.RR(
		buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack)),
		buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonHome)),
	))
	return buttonbuilder.RK(rows...)
}

// tapping a custom interval button deletes it (see ParseTimerButtonMinutes with TrackTimerDeletePrefix).
func TrackTimerDeleteReplyMenu(lang i18n.Lang, custom []int) tgbotapi.ReplyKeyboardMarkup {
	rows := timerIntervalRows(lang, custom, TrackTimerDeletePrefix)
	rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack))))
	return buttonbuilder.RK(rows...)
}

func timerIntervalRows(lang i18n.Lang, intervals []int, prefix string) [][]tgbotapi.KeyboardButton {
	rows := make([][]tgbotapi.KeyboardButton, 0, (len(intervals)+1)/2)
	for i := 0; i < len(intervals); i += 2 {
		if i+1 < len(intervals) {
			rows = append(rows, buttonbuilder.RR(
				buttonbuilder.RB(FormatTimerButton(lang, prefix, intervals[i])),
				buttonbuilder.RB(FormatTimerButton(lang, prefix, intervals[i+1])),
			))
		} else {
			rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(FormatTimerButton(lang, prefix, intervals[i]))))
		}
	}
	return rows
}

func FormatTimerButton(lang i18n.Lang, prefix string, minutes int) string {
	return fmt.Sprintf("%s%d %s", prefix, minutes, i18n.T(lang, i18n.KeyTrackMinutesUnit))
}

// falls back to Default's unit (like i18n.Key); rejects any text lacking the "<N> <unit>" suffix, even if the prefix matches.
func ParseTimerButtonMinutes(lang i18n.Lang, text, prefix string) (int, bool) {
	if n, ok := parseTimerButtonMinutesUnit(text, prefix, i18n.T(lang, i18n.KeyTrackMinutesUnit)); ok {
		return n, true
	}
	if lang != i18n.Default {
		if n, ok := parseTimerButtonMinutesUnit(text, prefix, i18n.T(i18n.Default, i18n.KeyTrackMinutesUnit)); ok {
			return n, true
		}
	}
	return 0, false
}

func parseTimerButtonMinutesUnit(text, prefix, unit string) (int, bool) {
	if !strings.HasPrefix(text, prefix) {
		return 0, false
	}
	suffix := " " + unit
	if !strings.HasSuffix(text, suffix) {
		return 0, false
	}
	numStr := strings.TrimSuffix(strings.TrimPrefix(text, prefix), suffix)
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func TrackActivitiesInlineMenu(lang i18n.Lang, items []models.TrackActivityItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}

		check := "⚪"
		if item.Selected {
			check = "🟢"
		}

		title := check + " " + item.Name
		if item.Emoji != "" {
			title = check + " " + item.Emoji + " " + item.Name
		}

		callbackData := fmt.Sprintf("act_toggle_:%d", item.ID)
		targetCB := fmt.Sprintf("%s%d", TrackCBActivityTarget, item.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, callbackData),
			tgbotapi.NewInlineKeyboardButtonData(TrackActivityTargetButtonLabel(lang, item), targetCB),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelArchiveSelected), TrackCBArchiveSelected),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelBack), "back_to_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// dueAt is embedded in callback data so a late answer credits the interval that was actually due, not "now minus interval".
func TrackPromptInlineMenu(lang i18n.Lang, items []models.TrackActivityItem, intervalMin int, dueAt time.Time) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		title := item.Name
		if item.Emoji != "" {
			title = item.Emoji + " " + item.Name
		}
		callbackData := fmt.Sprintf("%s%d:%d:%d", TrackCBPromptActivity, item.ID, intervalMin, dueAt.UTC().Unix())
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, callbackData),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelStopTimer), TrackCBPromptStopTimer),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackArchiveInlineMenu(lang i18n.Lang, items []models.TrackActivityItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)*2+1)
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		title := item.Name
		if item.Emoji != "" {
			title = item.Emoji + " " + item.Name
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(TrackLabelArchiveItemPrefix+title, "noop"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelRestore), fmt.Sprintf("%s%d", TrackCBArchiveRestore, item.ID)),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelDeleteForever), fmt.Sprintf("%s%d", TrackCBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelActiveActivities), TrackCBArchiveToActive),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelBack), "back_to_main"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackCreateSuccessInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelOpenActivities), TrackCBOpenActivities),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelCreateAnother), TrackCBCreateAnother),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelBack), "back_to_main"),
		),
	)
}

func TrackArchiveSuccessInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelOpenArchive), TrackCBOpenArchive),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelOpenActivities), TrackCBOpenActivities),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelBack), "back_to_main"),
		),
	)
}

func TrackReportsHubInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackButtonToday), TrackCBReportsToday),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackButtonCalendar), TrackCBReportsPeriodOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelBack), "back_to_main"),
		),
	)
}

func TrackReportTodayInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelSelectActivities), TrackCBReportsTodayBySelected),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelBackToReports), TrackCBReportsBackHub),
		),
	)
}

func TrackTodaySelectActivitiesInlineMenu(lang i18n.Lang, items []models.TrackActivityItem, selected map[int64]bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+2)
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		check := "☐"
		if selected[item.ID] {
			check = "☑"
		}
		title := check + " " + item.Name
		if item.Emoji != "" {
			title = check + " " + item.Emoji + " " + item.Name
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, fmt.Sprintf("%s%d", TrackCBReportsTodaySelToggle, item.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelBuildChart), TrackCBReportsTodaySelBuild),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelBack), TrackCBReportsToday),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackReportPeriodInlineMenu(lang i18n.Lang, items []models.TrackActivityItem, selected map[int64]bool, rangeLabel string) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+5)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelSelectedActivities), "noop"),
	))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		check := "☐"
		if selected[item.ID] {
			check = "☑"
		}
		title := check + " " + item.Name
		if item.Emoji != "" {
			title = check + " " + item.Emoji + " " + item.Name
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, fmt.Sprintf("%s%d", TrackCBReportsPeriodToggle, item.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelRangePrefix, rangeLabel), TrackCBReportsPeriodSetRange),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelTextReport), TrackCBReportsPeriodText),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelChartReport), TrackCBReportsPeriodChart),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelBackToReports), TrackCBReportsBackHub),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// indexes time.Month (1-12) via m-1, since the array itself is 0-indexed.
var monthNameKeys = [...]string{
	i18n.KeyTrackCalendarMonth01, i18n.KeyTrackCalendarMonth02, i18n.KeyTrackCalendarMonth03,
	i18n.KeyTrackCalendarMonth04, i18n.KeyTrackCalendarMonth05, i18n.KeyTrackCalendarMonth06,
	i18n.KeyTrackCalendarMonth07, i18n.KeyTrackCalendarMonth08, i18n.KeyTrackCalendarMonth09,
	i18n.KeyTrackCalendarMonth10, i18n.KeyTrackCalendarMonth11, i18n.KeyTrackCalendarMonth12,
}

func monthName(lang i18n.Lang, m time.Month) string {
	return i18n.T(lang, monthNameKeys[m-1])
}

func TrackReportPeriodCalendarInlineMenu(lang i18n.Lang, month time.Time, from, to time.Time) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 14)
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	startPad := (int(first.Weekday()) + 6) % 7

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("«Y", TrackCBReportsCalPrevYear),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %d", monthName(lang, first.Month()), first.Year()), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("Y»", TrackCBReportsCalNextYear),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀", TrackCBReportsCalPrev),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarMonth), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("▶", TrackCBReportsCalNext),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarMon), "noop"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarTue), "noop"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarWed), "noop"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarThu), "noop"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarFri), "noop"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarSat), "noop"),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarSun), "noop"),
	))

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
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, TrackCBReportsCalPick+dt.Format("2006-01-02")))
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
		confirmCB = TrackCBReportsCalDone
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(confirmLabel, confirmCB),
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackCalendarCancel), TrackCBReportsCalCancel),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// gridStart must be a Monday; 🟩 tracked / ⬛ missed both open drill-down, ⬜ upcoming days are inert.
func TrackHeatmapInlineMenu(lang i18n.Lang, gridStart, today time.Time, trackedDays map[string]bool, weeks int) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, weeks+2)

	header := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	for _, k := range weekdayShortKeys {
		header = append(header, tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, k), "noop"))
	}
	rows = append(rows, header)

	for w := 0; w < weeks; w++ {
		row := make([]tgbotapi.InlineKeyboardButton, 0, 7)
		for d := 0; d < 7; d++ {
			day := gridStart.AddDate(0, 0, w*7+d)
			if day.After(today) {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("⬜", "noop"))
				continue
			}
			emoji := "⬛"
			if trackedDays[day.Format("2006-01-02")] {
				emoji = "🟩"
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(emoji, TrackCBHeatmapDay+day.Format("2006-01-02")))
		}
		rows = append(rows, row)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.KeyTrackLabelBack), "back_to_main"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackHeatmapDayDetailInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelBack), TrackCBHeatmapBack)),
	)
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
