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

// Date is a calendar day: no instant, no zone, "2006-01-02".
//
// Never emit a calendar day as an RFC3339 instant. A client doing
// new Date("2026-08-21") gets UTC midnight and renders the 20th in any
// negative-offset zone — the exact bug class this type exists to prevent.
type Date string

// LocalTime is a naive wall clock in the requesting user's zone,
// "2006-01-02T15:04". It carries no offset by design and is always read
// together with Meta.Timezone.
type LocalTime string

// DateFromNaive and HourFromNaive format a value that came out of a
// `date_trunc(... AT TIME ZONE $n)` column.
//
// Such a value is a `timestamp without time zone`: pgx hands it back with UTC
// attached but a *local* wall clock inside. So it is formatted, never
// converted — calling .In(loc) on one shifts it by the offset and lands on the
// wrong day. That mistake is exactly what made the streak read zero for every
// user west of UTC before it was fixed in the service layer.
func DateFromNaive(t time.Time) Date      { return Date(t.Format("2006-01-02")) }
func HourFromNaive(t time.Time) LocalTime { return LocalTime(t.Format("2006-01-02T15:04")) }

// Seconds converts a duration for the wire. Whole seconds, and every such field
// is named with a _seconds suffix, so nobody has to guess the unit: milliseconds
// would be noise at this resolution and minutes would lose a partial session.
func Seconds(d time.Duration) int64 {
	return int64(d / time.Second)
}
