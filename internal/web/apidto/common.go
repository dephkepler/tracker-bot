// Package apidto holds the wire types of the dashboard API and the mapping
// from the service layer's models into them.
//
// Nothing from internal/models is ever handed to json.Marshal directly, for two
// reasons that have both already caused bugs elsewhere:
//
//   - time.Duration marshals as int64 nanoseconds. A chart fed that is six
//     orders of magnitude too tall, and nothing errors.
//   - Some service DTOs carry pre-formatted presentation strings
//     (models.LearningStats.NextPushIn and models.RoadmapStats.NextPushIn are
//     fmt.Sprintf("%d min", …)). Those are for a Telegram message, not an API.
package apidto

import "time"

// Instant is a real point in time, formatted RFC3339 with the user's own
// offset so a client reading it sees the wall clock the user would.
type Instant string

// FromInstant formats a genuine instant. Use it only for actual timestamps —
// calendar days go out as Date, which has no offset to misread.
func FromInstant(t time.Time) Instant {
	return Instant(t.Format(time.RFC3339))
}

// Seconds converts a duration for the wire. Whole seconds, and every such field
// is named with a _seconds suffix, so nobody has to guess the unit: milliseconds
// would be noise at this resolution and minutes would lose a partial session.
func Seconds(d time.Duration) int64 {
	return int64(d / time.Second)
}
