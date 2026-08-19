package track

import (
	"fmt"
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
