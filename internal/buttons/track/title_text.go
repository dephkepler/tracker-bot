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

// TrackHeatmapText renders a GitHub-style calendar heatmap: 🟩 tracked, ⬛
// missed, ⬜ upcoming (only appears within the current, still-incomplete
// week). gridStart must be a Monday; today is the last day considered
// "past" (inclusive).
func TrackHeatmapText(lang i18n.Lang, gridStart, today time.Time, trackedDays map[string]bool, weeks int) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapTitle))
	b.WriteString("\n\n")

	header := make([]string, 0, 7)
	for _, k := range weekdayShortKeys {
		header = append(header, i18n.T(lang, k))
	}
	b.WriteString(strings.Join(header, " "))
	b.WriteString("\n")

	trackedCount, totalPast := 0, 0
	for w := 0; w < weeks; w++ {
		cells := make([]string, 0, 7)
		for d := 0; d < 7; d++ {
			day := gridStart.AddDate(0, 0, w*7+d)
			switch {
			case day.After(today):
				cells = append(cells, "⬜")
			case trackedDays[day.Format("2006-01-02")]:
				cells = append(cells, "🟩")
				trackedCount++
				totalPast++
			default:
				cells = append(cells, "⬛")
				totalPast++
			}
		}
		b.WriteString(strings.Join(cells, " "))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapLegend))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyTrackHeatmapDaysTracked, trackedCount, totalPast))
	return b.String()
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
