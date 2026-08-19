package track

import (
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

// Main screen
type TrackMainStats struct {
	CurrentActivityName string
	TodayTrackedTime    string
	TodayActivityCount  string
	CurrentStreakDays   string
}

// Activity report
type TrackActivityReportStats struct {
	ActivityStartDate    string
	ConsecutiveDaysCount string
	TodayAccumulatedTime string
	AverageDailyTime     string
	ReportDate           string
}

func TrackingMenuText(lang i18n.Lang, stats models.MainStats) string {
	target := 120 * time.Minute
	progress := progressBar(lang, stats.TodayTracked, target, 10)
	return fmt.Sprintf(
		"%s\n\n%s *%s*\n%s *%s*\n`%s`\n%s *%d*\n%s *%d*\n",
		i18n.T(lang, i18n.KeyTrackMainTitle),
		i18n.T(lang, i18n.KeyTrackMainCurrentActivity), safeText(stats.CurrentActivityName),
		i18n.T(lang, i18n.KeyTrackMainTodayTime), formatDuration(stats.TodayTracked),
		progress,
		i18n.T(lang, i18n.KeyTrackMainStreak), stats.StreakDays,
		i18n.T(lang, i18n.KeyTrackMainTodayCount), stats.TodaySessions,
	)
}

// formatDuration formats duration into human-readable string like "4h 30m".
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

// weekdayShortKeys indexes Mon(0)..Sun(6) to the matching i18n key — same
// order as the calendar picker's header row.
var weekdayShortKeys = [...]string{
	i18n.KeyTrackCalendarMon, i18n.KeyTrackCalendarTue, i18n.KeyTrackCalendarWed,
	i18n.KeyTrackCalendarThu, i18n.KeyTrackCalendarFri, i18n.KeyTrackCalendarSat, i18n.KeyTrackCalendarSun,
}

// TrackHeatmapText renders the caption above the tappable heatmap grid (see
// TrackHeatmapInlineMenu): title, tap hint, legend, and a tracked-days
// tally. gridStart must be a Monday; today is the last day considered
// "past" (inclusive).
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

// TrackHeatmapDayTitle renders a localized header for one day's drill-down,
// e.g. "📅 Wed, 19 Aug 2026".
func TrackHeatmapDayTitle(lang i18n.Lang, day time.Time) string {
	weekday := i18n.T(lang, weekdayShortKeys[(int(day.Weekday())+6)%7])
	return fmt.Sprintf("📅 %s, %d %s %d", weekday, day.Day(), monthName(lang, day.Month()), day.Year())
}

// safeText returns fallback when string is empty.
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
