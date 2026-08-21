package apidto

// Meta says how the server interpreted the request. Timezone matters on every
// response: it is what makes the day boundaries behind these numbers
// meaningful, and the client did not choose it — the user's profile did.
type Meta struct {
	Timezone    string  `json:"timezone"`
	GeneratedAt Instant `json:"generated_at"`
	// The window the server actually used, inclusive on both ends. Echoed back
	// because the client may have asked for "week" and needs to know which
	// seven days that turned out to be.
	From Date `json:"from,omitempty"`
	To   Date `json:"to,omitempty"`
	// Granularity is resolved, never "auto".
	Granularity string `json:"granularity,omitempty"`
	// ActivityIDs is the set actually queried — the expansion of an omitted
	// filter, so the client can tell what "everything" meant.
	ActivityIDs []int64 `json:"activity_ids,omitempty"`
}

// Activity is one of the user's activities.
type Activity struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	// Selected is the bot's "what I'm tracking right now" flag.
	Selected bool `json:"selected"`
	Archived bool `json:"archived"`
	// TargetMinutes is this activity's own daily goal, null when unset. Per
	// activity since migration 0013 — there is deliberately no global target
	// anywhere in this API.
	TargetMinutes *int `json:"target_minutes"`
}

type ActivitiesResponse struct {
	Activities []Activity `json:"activities"`
	Meta       Meta       `json:"meta"`
}

// ActivityTotal is one activity's slice of a period.
type ActivityTotal struct {
	ActivityID int64  `json:"activity_id"`
	Name       string `json:"name"`
	Emoji      string `json:"emoji"`
	Seconds    int64  `json:"seconds"`
	Sessions   int    `json:"sessions"`
	// SharePercent is this activity's share of the total in *this* response,
	// 0..100 with one decimal. Computed server-side so two clients cannot
	// round it differently.
	SharePercent float64 `json:"share_percent"`
}

// CurrentActivity is the last activity the user tracked.
//
// Every number here is scoped to that one activity, which is why they are
// nested rather than sitting beside the day totals: TodaySeconds is not the
// day's total and StreakDays is not an overall streak. Flattening them into
// the response root is exactly the mistake models.MainStats used to invite.
type CurrentActivity struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Emoji         string `json:"emoji"`
	TodaySeconds  int64  `json:"today_seconds"`
	StreakDays    int    `json:"streak_days"`
	TargetMinutes *int   `json:"target_minutes"`
}

// OverviewToday is the whole day, across every activity.
type OverviewToday struct {
	TotalSeconds int64 `json:"total_seconds"`
	Sessions     int   `json:"sessions"`
	// ActivitiesCount is how many distinct activities were touched today.
	ActivitiesCount int             `json:"activities_count"`
	TopActivities   []ActivityTotal `json:"top_activities"`
}

type OverviewResponse struct {
	Today OverviewToday `json:"today"`
	// Current is null when nothing has ever been tracked — a fresh account,
	// which the client has to render as an empty state rather than as zeros
	// for an activity that does not exist.
	Current *CurrentActivity `json:"current_activity"`
	Meta    Meta             `json:"meta"`
}

// MonthTotal is one month of a period. Month is the first day of that month in
// the user's zone.
type MonthTotal struct {
	Month   Date  `json:"month"`
	Seconds int64 `json:"seconds"`
}

type BreakdownResponse struct {
	TotalSeconds  int64           `json:"total_seconds"`
	TotalSessions int             `json:"total_sessions"`
	Activities    []ActivityTotal `json:"activities"`
	// Monthly is only populated for periods long enough for the service to
	// bucket by month; it is a convenience, not the series endpoint.
	Monthly []MonthTotal `json:"monthly"`
	Meta    Meta         `json:"meta"`
}

// SeriesPart is one activity's slice of a bucket.
//
// No activity id: repo.GetHourlyBucketsByActivity selects only the name and
// emoji. Adding the id means touching that query and
// models.HourActivityDuration, which is a change worth making deliberately
// rather than in passing.
type SeriesPart struct {
	Name    string `json:"name"`
	Emoji   string `json:"emoji"`
	Seconds int64  `json:"seconds"`
}

// SeriesBucket is one point of the series. Start is a Date for day and month
// granularity and a LocalTime for hour, which is why it is a plain string.
type SeriesBucket struct {
	Start   string       `json:"start"`
	Seconds int64        `json:"seconds"`
	Parts   []SeriesPart `json:"parts,omitempty"`
}

type SeriesResponse struct {
	// By is "total" or "activity".
	By      string         `json:"by"`
	Buckets []SeriesBucket `json:"buckets"`
	Meta    Meta           `json:"meta"`
}

// HeatDay is one day of the calendar heatmap. Only days with tracked time are
// sent; the client fills the grid from Meta.From and Meta.To.
type HeatDay struct {
	Date    Date  `json:"date"`
	Seconds int64 `json:"seconds"`
}

type HeatmapResponse struct {
	Days []HeatDay `json:"days"`
	// MaxSeconds is the busiest day in the window, so the client can scale the
	// colour ramp without a second pass over the data.
	MaxSeconds int64 `json:"max_seconds"`
	Meta       Meta  `json:"meta"`
}

type DayResponse struct {
	Date          Date            `json:"date"`
	TotalSeconds  int64           `json:"total_seconds"`
	TotalSessions int             `json:"total_sessions"`
	Activities    []ActivityTotal `json:"activities"`
	// Hours is the same shape as a series bucketed by hour, with each bucket's
	// activities as parts.
	Hours []SeriesBucket `json:"hours"`
	Meta  Meta           `json:"meta"`
}
