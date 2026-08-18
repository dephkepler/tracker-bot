// Package apptime resolves the timezone used for calendar/day boundary
// decisions ("today", "this month", period report defaults) — per user when
// known (detected once via a shared Telegram location, see pkg/geotz and
// internal/service/profile.go ChangeTimeZone), falling back to a single
// default otherwise (new users who haven't shared a location yet, and
// background jobs with no specific user in scope, like the timer scheduler).
//
// All timestamps are still stored in and read from Postgres as UTC instants
// (TIMESTAMPTZ) — nothing here changes that. This only affects which calendar
// day/month/year a given instant is considered to fall into when the bot
// talks to a human.
package apptime

import (
	"time"

	// Ensures time.LoadLocation works even on minimal images without the OS
	// tzdata package installed (e.g. bare alpine without apk add tzdata).
	_ "time/tzdata"

	"github.com/rs/zerolog/log"
)

// defaultZoneName is used until a user shares their location, and for
// anything running without a specific user in scope.
const defaultZoneName = "Europe/Warsaw"

// Location is the fallback/default timezone.
var Location = mustLoad(defaultZoneName)

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Error().Err(err).Str("zone", name).Msg("apptime: failed to load location, falling back to UTC")
		return time.UTC
	}
	return loc
}

// Resolve returns the *time.Location for a stored IANA zone name (as saved
// by ChangeTimeZone), falling back to Location if empty or invalid.
func Resolve(tzName string) *time.Location {
	if tzName == "" {
		return Location
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		log.Error().Err(err).Str("zone", tzName).Msg("apptime: invalid stored zone, falling back to default")
		return Location
	}
	return loc
}

// Now returns the current time in the default Location. Prefer NowIn with a
// specific user's resolved location wherever one is known.
func Now() time.Time {
	return time.Now().In(Location)
}

// NowIn returns the current time in the given location.
func NowIn(loc *time.Location) time.Time {
	if loc == nil {
		loc = Location
	}
	return time.Now().In(loc)
}

// ParseDay parses a "2006-01-02" date as midnight in the given location.
func ParseDay(s string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = Location
	}
	return time.ParseInLocation("2006-01-02", s, loc)
}
