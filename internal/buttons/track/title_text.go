package track

import (
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

type TrackMainStats struct {
	CurrentActivityName string
	TodayTrackedTime    string
	TodayActivityCount  string
	CurrentStreakDays   string
}

type TrackActivityReportStats struct {
	ActivityStartDate    string
	ConsecutiveDaysCount string
	TodayAccumulatedTime string
	AverageDailyTime     string
	ReportDate           string
}

func TrackingMenuText(lang i18n.Lang, stats models.MainStats) string {
	target := 120 * time.Minute
	if stats.TargetMinutes != nil {
		target = time.Duration(*stats.TargetMinutes) * time.Minute
	}
	progress := progressBar(lang, stats.TodayTracked, target, 10)
	return fmt.Sprintf(
		"%s\n\n%s *%s*\n%s *%s*\n`%s`\n%s *%d*\n%s *%d*\n",
		i18n.T(lang, i18n.KeyTrackMainTitle),
		i18n.T(lang, i18n.KeyTrackMainCurrentActivity), safeText(currentActivityLabel(stats)),
		i18n.T(lang, i18n.KeyTrackMainTodayTime), formatDuration(stats.TodayTracked),
		progress,
		i18n.T(lang, i18n.KeyTrackMainStreak), stats.StreakDays,
		i18n.T(lang, i18n.KeyTrackMainTodayCount), stats.TodaySessions,
	)
}

// currentActivityLabel joins the emoji and name the service now returns
// separately. The join lives here because it is presentation: the dashboard API
// reads the same MainStats and needs the two apart.
func currentActivityLabel(stats models.MainStats) string {
	if strings.TrimSpace(stats.CurrentActivityEmoji) == "" {
		return stats.CurrentActivityName
	}
	return stats.CurrentActivityEmoji + " " + stats.CurrentActivityName
}

// TrackActivityTargetPromptText asks for a daily minutes target for one
// activity — the numeric-input prompt behind the 🎯 button in
// TrackActivitiesInlineMenu.
func TrackActivityTargetPromptText(lang i18n.Lang, activityName string) string {
	return i18n.T(lang, i18n.KeyTrackActivityTargetPromptFmt, activityName)
}

// TrackActivityTargetButtonLabel renders the per-row 🎯 button: the
// configured target if set, or an invitation to set one.
func TrackActivityTargetButtonLabel(lang i18n.Lang, item models.TrackActivityItem) string {
	if item.TargetMinutes != nil {
		return i18n.T(lang, i18n.KeyTrackActivityTargetButtonSetFmt, *item.TargetMinutes)
	}
	return i18n.T(lang, i18n.KeyTrackActivityTargetButtonUnset)
}

// TrackTimerStatusText renders the "which activities are currently
// activated, and when's the next reminder" screen. now is the reference
// instant nextPingAt's remaining time is computed against.
func TrackTimerStatusText(lang i18n.Lang, enabled bool, intervalMin int, nextPingAt, now time.Time, selected []models.TrackActivityItem) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyTrackTimerStatusTitle))
	b.WriteString("\n\n")
	if enabled {
		remaining := nextPingAt.Sub(now)
		if remaining < 0 {
			remaining = 0
		}
		b.WriteString(i18n.T(lang, i18n.KeyTrackTimerStatusActiveFmt, intervalMin, formatDuration(remaining)))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyTrackTimerStatusInactive))
	}
	b.WriteString("\n\n")

	if len(selected) == 0 {
		b.WriteString(i18n.T(lang, i18n.KeyTrackTimerStatusNoActivities))
		return b.String()
	}
	b.WriteString(i18n.T(lang, i18n.KeyTrackTimerStatusActivitiesHeader))
	for _, item := range selected {
		b.WriteString("\n")
		b.WriteString(activityLine(item))
	}
	return b.String()
}

func activityLine(item models.TrackActivityItem) string {
	if item.Emoji != "" {
		return "• " + item.Emoji + " " + item.Name
	}
	return "• " + item.Name
}

func formatDuration(d time.Duration) string {
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

// Mon(0)..Sun(6) — must match the calendar picker's header row order.
var weekdayShortKeys = [...]string{
	i18n.KeyTrackCalendarMon, i18n.KeyTrackCalendarTue, i18n.KeyTrackCalendarWed,
	i18n.KeyTrackCalendarThu, i18n.KeyTrackCalendarFri, i18n.KeyTrackCalendarSat, i18n.KeyTrackCalendarSun,
}

// gridStart must be a Monday; today is the last day considered "past" (inclusive).
func TrackHeatmapText(lang i18n.Lang, gridStart, today time.Time, trackedDays map[string]bool, weeks int) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapTitle))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapHint))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapLegend))
	b.WriteString("\n\n")

	trackedCount, totalPast := 0, 0
	for w := 0; w < weeks; w++ {
		for d := 0; d < 7; d++ {
			day := gridStart.AddDate(0, 0, w*7+d)
			if day.After(today) {
				continue
			}
			totalPast++
			if trackedDays[day.Format("2006-01-02")] {
				trackedCount++
			}
		}
	}
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapDaysTracked, trackedCount, totalPast))
	return b.String()
}

func TrackHeatmapDayTitle(lang i18n.Lang, day time.Time) string {
	weekday := i18n.T(lang, weekdayShortKeys[(int(day.Weekday())+6)%7])
	return fmt.Sprintf("📅 %s, %d %s %d", weekday, day.Day(), monthName(lang, day.Month()), day.Year())
}

func safeText(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func progressBar(lang i18n.Lang, value, target time.Duration, width int) string {
	if width <= 0 {
		width = 10
	}
	if target <= 0 {
		target = time.Minute
	}
	if value < 0 {
		value = 0
	}
	ratio := float64(value) / float64(target)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	percent := int(ratio * 100)
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return i18n.T(lang, i18n.KeyTrackMainProgress, bar, percent, formatDuration(target))
}
