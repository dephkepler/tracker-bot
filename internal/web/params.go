package web

import (
	"net/http"
	"strings"
)

// boolParam reads a query flag. Present-but-empty ("?include_archived") counts
// as true, which is how people type flags by hand; anything unrecognised is
// false rather than an error, because a mistyped optional flag should not fail
// the whole request.
func boolParam(r *http.Request, name string) bool {
	if !r.URL.Query().Has(name) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get(name))) {
	case "", "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
