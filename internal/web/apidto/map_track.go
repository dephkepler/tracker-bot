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

func FromPeriodReport(report models.ReportPeriodStats, meta Meta) BreakdownResponse {
	monthly := make([]MonthTotal, 0, len(report.Monthly))
	for _, m := range report.Monthly {
		monthly = append(monthly, MonthTotal{
			// Naive: formatted, never converted.
			Month:   DateFromNaive(m.Month),
			Seconds: Seconds(m.Duration),
		})
	}
	return BreakdownResponse{
		TotalSeconds:  Seconds(report.TotalTracked),
		TotalSessions: report.TotalSessions,
		Activities:    FromActivityStats(report.Activities, report.TotalTracked),
		Monthly:       monthly,
		Meta:          meta,
	}
}

// FromBuckets maps the total-per-bucket series. granularity decides how the
// bucket start is rendered: an hour bucket needs the time of day, a day or
// month bucket must stay a bare calendar date so no client can shift it.
func FromBuckets(starts []time.Time, durations []time.Duration, granularity string, meta Meta) SeriesResponse {
	n := min(len(starts), len(durations))
	buckets := make([]SeriesBucket, 0, n)
	for i := range n {
		buckets = append(buckets, SeriesBucket{
			Start:   formatBucketStart(starts[i], granularity),
			Seconds: Seconds(durations[i]),
		})
	}
	return SeriesResponse{By: "total", Buckets: buckets, Meta: meta}
}

// FromHourlyByActivity groups the per-activity rows into hour buckets.
//
// The repository already returns them ordered by bucket and then by duration
// descending, so this only has to notice when the bucket changes — no sorting,
// and the ranking inside each bucket is preserved.
func FromHourlyByActivity(rows []models.HourActivityDuration, meta Meta) SeriesResponse {
	buckets := make([]SeriesBucket, 0, 24)
	for _, row := range rows {
		start := formatBucketStart(row.BucketStart, "hour")
		if len(buckets) == 0 || buckets[len(buckets)-1].Start != start {
			buckets = append(buckets, SeriesBucket{Start: start})
		}
		b := &buckets[len(buckets)-1]
		b.Seconds += Seconds(row.Duration)
		b.Parts = append(b.Parts, SeriesPart{
			Name:    row.Name,
			Emoji:   row.Emoji,
			Seconds: Seconds(row.Duration),
		})
	}
	return SeriesResponse{By: "activity", Buckets: buckets, Meta: meta}
}

func FromDailyBuckets(starts []time.Time, durations []time.Duration, meta Meta) HeatmapResponse {
	n := min(len(starts), len(durations))
	days := make([]HeatDay, 0, n)
	var maxSeconds int64
	for i := range n {
		seconds := Seconds(durations[i])
		if seconds <= 0 {
			// The query only returns days that have sessions, but a zero-length
			// day would be noise in a heatmap either way.
			continue
		}
		days = append(days, HeatDay{Date: DateFromNaive(starts[i]), Seconds: seconds})
		maxSeconds = max(maxSeconds, seconds)
	}
	return HeatmapResponse{Days: days, MaxSeconds: maxSeconds, Meta: meta}
}

func formatBucketStart(t time.Time, granularity string) string {
	if granularity == "hour" {
		return string(HourFromNaive(t))
	}
	return string(DateFromNaive(t))
}
