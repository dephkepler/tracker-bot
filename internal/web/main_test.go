package web

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

// The middleware logs an access line per request, which for a few hundred test
// requests buries the one line that matters — a failure. Silenced here rather
// than by injecting a logger everywhere: the logging itself is behaviour worth
// keeping, it is only the output that is unwanted in a test run.
func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}
