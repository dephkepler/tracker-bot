package apidto

import (
	"math"
	"time"

	"tracker-bot/internal/models"
)

// Mapping is pure: no context, no HTTP, no database. That is what lets the
// rounding and the share arithmetic be tested directly.

// NewMeta stamps the response with the zone its numbers were bucketed by.
func NewMeta(loc *time.Location, now time.Time) Meta {
	return Meta{
		Timezone:    loc.String(),
		GeneratedAt: FromInstant(now.In(loc)),
	}
}

func FromActivityItems(items []models.TrackActivityItem, archived bool) []Activity {
	out := make([]Activity, 0, len(items))
	for _, item := range items {
		out = append(out, Activity{
			ID:            item.ID,
			Name:          item.Name,
			Emoji:         item.Emoji,
			Selected:      item.Selected,
			Archived:      archived,
			TargetMinutes: item.TargetMinutes,
		})
	}
	return out
}

// FromActivityStats converts a period breakdown, filling in each activity's
// share of total. total is passed in rather than summed here: the caller
// already has the authoritative figure from the query, and re-deriving it
// would disagree whenever the breakdown is a top-N slice rather than the whole
// period.
func FromActivityStats(stats []models.ActivityDurationStat, total time.Duration) []ActivityTotal {
	out := make([]ActivityTotal, 0, len(stats))
	for _, s := range stats {
		out = append(out, ActivityTotal{
			ActivityID:   s.ActivityID,
			Name:         s.Name,
			Emoji:        s.Emoji,
			Seconds:      Seconds(s.Duration),
			Sessions:     s.Sessions,
			SharePercent: sharePercent(s.Duration, total),
		})
	}
	return out
}

// sharePercent is rounded to one decimal. A zero total means an empty period,
// where every share is zero rather than a division by zero.
func sharePercent(part, total time.Duration) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(float64(part)/float64(total)*1000) / 10
}

// FromMainStats maps the current-activity block. It returns nil when nothing
// has ever been tracked: GetMainStats reports that as a zero-valued struct, and
// passing it through as an activity with id 0 would have the client render a
// nameless row with a zero streak.
func FromMainStats(stats models.MainStats) *CurrentActivity {
	if stats.CurrentActivityID == 0 {
		return nil
	}
	return &CurrentActivity{
		ID:            stats.CurrentActivityID,
		Name:          stats.CurrentActivityName,
		Emoji:         stats.CurrentActivityEmoji,
		TodaySeconds:  Seconds(stats.TodayTracked),
		StreakDays:    stats.StreakDays,
		TargetMinutes: stats.TargetMinutes,
	}
}

// FromTodayReport maps the day totals. activitiesCount comes from MainStats,
// which is the only place that number is computed.
func FromTodayReport(report models.ReportTodayStats, activitiesCount int) OverviewToday {
	return OverviewToday{
		TotalSeconds:    Seconds(report.TotalTracked),
		Sessions:        report.TotalSessions,
		ActivitiesCount: activitiesCount,
		TopActivities:   FromActivityStats(report.TopActivities, report.TotalTracked),
	}
}
