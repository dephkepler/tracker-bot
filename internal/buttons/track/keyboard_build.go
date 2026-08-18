package track

import (
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

func TrackEntryInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(TrackButtonSelectActivity, TrackCBActivitySelect),
			buttonbuilder.IB(TrackButtonCreateActivity, TrackCBActivityCreate),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackButtonViewReports, TrackCBReportSummary),
			buttonbuilder.IB(TrackButtonViewArchive, TrackCBArchiveOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackButtonBackHome, "go_home"),
		),
	)
}

// Reply button menus

func TrackActivityListReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonToday), buttonbuilder.RB(TrackButtonBack)),
	)
}

func TrackActivityReportReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonReportPeriod), buttonbuilder.RB(TrackButtonReportWeek)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonReportExport), buttonbuilder.RB(TrackButtonToday)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonReportDelete), buttonbuilder.RB(TrackButtonBack)),
	)
}

func TrackActivityManageReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonActivityActivate), buttonbuilder.RB(TrackButtonActivityArchive)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonActivityDelete), buttonbuilder.RB(TrackButtonViewArchive)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonBack), buttonbuilder.RB(TrackButtonBackHome)),
	)
}

func TrackArchiveReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonSelectActivity), buttonbuilder.RB(TrackButtonViewArchive)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonBack), buttonbuilder.RB(TrackButtonBackHome)),
	)
}

func TrackReportsReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonToday), buttonbuilder.RB(TrackButtonPeriod)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonBack), buttonbuilder.RB(TrackButtonBackHome)),
	)
}

func TrackTimerReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonTimer15), buttonbuilder.RB(TrackButtonTimer30)),
		buttonbuilder.RR(buttonbuilder.RB(TrackButtonBack), buttonbuilder.RB(TrackButtonBackHome)),
	)
}

func TrackActivitiesInlineMenu(items []models.TrackActivityItem) tgbotapi.InlineKeyboardMarkup {
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
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, callbackData),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelArchiveSelected, TrackCBArchiveSelected),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelBack, "back_to_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackPromptInlineMenu(items []models.TrackActivityItem, intervalMin int) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		title := item.Name
		if item.Emoji != "" {
			title = item.Emoji + " " + item.Name
		}
		callbackData := fmt.Sprintf("%s%d:%d", TrackCBPromptActivity, item.ID, intervalMin)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, callbackData),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelStopTimer, TrackCBPromptStopTimer),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackArchiveInlineMenu(items []models.TrackActivityItem) tgbotapi.InlineKeyboardMarkup {
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
			tgbotapi.NewInlineKeyboardButtonData(TrackLabelRestore, fmt.Sprintf("%s%d", TrackCBArchiveRestore, item.ID)),
			tgbotapi.NewInlineKeyboardButtonData(TrackLabelDeleteForever, fmt.Sprintf("%s%d", TrackCBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelActiveActivities, TrackCBArchiveToActive),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelBack, "back_to_main"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackCreateSuccessInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelOpenActivities, TrackCBOpenActivities),
			buttonbuilder.IB(TrackLabelCreateAnother, TrackCBCreateAnother),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelBack, "back_to_main"),
		),
	)
}

func TrackArchiveSuccessInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelOpenArchive, TrackCBOpenArchive),
			buttonbuilder.IB(TrackLabelOpenActivities, TrackCBOpenActivities),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelBack, "back_to_main"),
		),
	)
}

func TrackReportsHubInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(TrackButtonToday, TrackCBReportsToday),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackButtonPeriod, TrackCBReportsPeriodOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelBack, "back_to_main"),
		),
	)
}

func TrackReportTodayInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelSelectActivities, TrackCBReportsTodayBySelected),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(TrackLabelBackToReports, TrackCBReportsBackHub),
		),
	)
}

func TrackTodaySelectActivitiesInlineMenu(items []models.TrackActivityItem, selected map[int64]bool) tgbotapi.InlineKeyboardMarkup {
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
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelBuildChart, TrackCBReportsTodaySelBuild),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelBack, TrackCBReportsToday),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackReportPeriodInlineMenu(items []models.TrackActivityItem, selected map[int64]bool, rangeLabel string) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+5)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelSelectedActivities, "noop"),
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
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelRangePrefix+rangeLabel, TrackCBReportsPeriodSetRange),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelTextReport, TrackCBReportsPeriodText),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelChartReport, TrackCBReportsPeriodChart),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelBackToReports, TrackCBReportsBackHub),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TrackReportPeriodCalendarInlineMenu(month time.Time, from, to time.Time) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 14)
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	startPad := (int(first.Weekday()) + 6) % 7

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("«Y", TrackCBReportsCalPrevYear),
		tgbotapi.NewInlineKeyboardButtonData(first.Format("January 2006"), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("Y»", TrackCBReportsCalNextYear),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀", TrackCBReportsCalPrev),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelMonth, "noop"),
		tgbotapi.NewInlineKeyboardButtonData("▶", TrackCBReportsCalNext),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelMon, "noop"),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelTue, "noop"),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelWed, "noop"),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelThu, "noop"),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelFri, "noop"),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelSat, "noop"),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelSun, "noop"),
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
	confirmLabel := TrackLabelSelectEndDate
	confirmCB := "noop"
	if !from.IsZero() && !to.IsZero() {
		confirmLabel = TrackLabelConfirmRange
		confirmCB = TrackCBReportsCalDone
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(confirmLabel, confirmCB),
		tgbotapi.NewInlineKeyboardButtonData(TrackLabelCancel, TrackCBReportsCalCancel),
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
