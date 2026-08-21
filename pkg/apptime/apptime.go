// Package apptime resolves display timezone only; timestamps remain stored/read as UTC (TIMESTAMPTZ) in Postgres.
package apptime

import (
	"time"

	// ensures time.LoadLocation works without OS tzdata installed (e.g. bare alpine).
	_ "time/tzdata"

	"github.com/rs/zerolog/log"
)

const defaultZoneName = "Europe/Warsaw"

var Location = mustLoad(defaultZoneName)

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Error().Err(err).Str("zone", name).Msg("apptime: failed to load location, falling back to UTC")
		return time.UTC
	}
	return loc
}

// Resolve falls back to Location if tzName is empty or an invalid IANA zone.
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

// prefer NowIn with a specific user's resolved location wherever one is known.
func Now() time.Time {
	return time.Now().In(Location)
}

func NowIn(loc *time.Location) time.Time {
	if loc == nil {
		loc = Location
	}
	return time.Now().In(loc)
}

func ParseDay(s string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = Location
	}
	return time.ParseInLocation("2006-01-02", s, loc)
}
