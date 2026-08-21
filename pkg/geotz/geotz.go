// resolves timezone from lat/lng entirely offline, via tzf-dist data embedded in the binary.
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
		// embedded data failing to parse means the binary is broken; no fallback — fail fast.
		panic(fmt.Errorf("geotz: init finder: %w", err))
	}
	finder = f
}

// also validates the result loads via time.LoadLocation, in case tzf returns a stale/unknown zone.
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
