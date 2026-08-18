// Package geotz resolves an IANA timezone name from geographic coordinates,
// fully offline (data is embedded in the binary via tzf-dist — no network
// call, no filesystem lookup, works fine in a minimal container).
package geotz

import (
	"fmt"
	"time"

	"github.com/ringsaturn/tzf"
)

var finder tzf.F

func init() {
	f, err := tzf.NewDefaultFinder()
	if err != nil {
		// Embedded data failing to parse means the binary itself is broken —
		// there is no reasonable fallback, fail fast at startup.
		panic(fmt.Errorf("geotz: init finder: %w", err))
	}
	finder = f
}

// Lookup returns the IANA timezone name for the given coordinates (e.g. a
// Telegram-shared location), validated to be loadable via time.LoadLocation.
func Lookup(lat, lng float64) (string, error) {
	// tzf takes longitude before latitude.
	name := finder.GetTimezoneName(lng, lat)
	if name == "" {
		return "", fmt.Errorf("geotz: no timezone found for lat=%v lng=%v", lat, lng)
	}
	if _, err := time.LoadLocation(name); err != nil {
		return "", fmt.Errorf("geotz: resolved zone %q does not load: %w", name, err)
	}
	return name, nil
}
