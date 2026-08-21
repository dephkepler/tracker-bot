package apidto

// Meta says how the server interpreted the request. Timezone matters on every
// response: it is what makes the day boundaries behind these numbers
// meaningful, and the client did not choose it — the user's profile did.
type Meta struct {
	Timezone    string  `json:"timezone"`
	GeneratedAt Instant `json:"generated_at"`
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
