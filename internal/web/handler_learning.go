package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"tracker-bot/internal/utils/webctx"
	"tracker-bot/internal/web/apidto"
	"tracker-bot/pkg/apptime"
)

const (
	learningDefaultDays = 30
	learningMaxDays     = 366
)

// handleLearning answers the flashcards section: how many words there are, how
// the reviews have been going, and per-collection progress.
//
// Read-only, and staying that way. Grading a word means the SM-2 scheduler, the
// four-grade keyboard and the relearning steps — the bot does that well, and
// duplicating a scheduler across two front-ends is how the two start disagreeing
// about when a word is next due.
func (s *Server) handleLearning(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	days, perr := boundedDays(r.URL.Query().Get("days"), learningDefaultDays, learningMaxDays)
	if perr != nil {
		s.failRequest(w, r, perr, "learning params")
		return
	}

	detail, err := s.learningsvc.GetStatsDetail(r.Context(), user.DBUserID, user.Location)
	if err != nil {
		s.fail(w, r, err, "learning stats failed")
		return
	}

	// The window is built from the user's own midnights, like every other range
	// in this API.
	now := apptime.NowIn(user.Location)
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, user.Location).AddDate(0, 0, 1)
	from := to.AddDate(0, 0, -days)

	entries, err := s.learningsvc.ListReviewsOnDay(r.Context(), user.DBUserID, from, to)
	if err != nil {
		s.fail(w, r, err, "list reviews failed")
		return
	}

	meta := apidto.NewMeta(user.Location, now)
	meta.From = apidto.Date(from.Format("2006-01-02"))
	meta.To = apidto.Date(to.AddDate(0, 0, -1).Format("2006-01-02"))
	meta.Granularity = granDay

	writeJSON(w, r, http.StatusOK, apidto.FromLearningDetail(
		detail, apidto.CountReviewsByDay(entries, from, to, user.Location), meta,
	))
}

// boundedDays reads a positive day count with a ceiling.
func boundedDays(raw string, def, maxDays int) (int, *paramError) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 {
		return 0, badParam("days must be a positive number, got %q", raw)
	}
	if days > maxDays {
		return 0, badParam("days must be at most %d", maxDays)
	}
	return days, nil
}
