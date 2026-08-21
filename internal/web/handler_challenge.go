package web

import (
	"errors"
	"net/http"

	"tracker-bot/internal/models"
	"tracker-bot/internal/utils/webctx"
	"tracker-bot/internal/web/apidto"
	"tracker-bot/pkg/apptime"
)

// handleChallenges returns the active challenges with their full day grids.
//
// The grid is the whole point of this section — a challenge is a row of squares
// — and it is bounded by the schema itself: a challenge spans under a hundred
// days, so sending every day costs nothing and saves a request per challenge.
func (s *Server) handleChallenges(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	items, err := s.challengesvc.ListChallenges(r.Context(), user.DBUserID)
	if err != nil {
		s.fail(w, r, err, "list challenges failed")
		return
	}

	now := apptime.NowIn(user.Location)
	out := apidto.ChallengesResponse{
		Challenges: make([]apidto.Challenge, 0, len(items)),
		Meta:       apidto.NewMeta(user.Location, now),
	}

	for _, item := range items {
		days, err := s.challengesvc.ListDays(r.Context(), user.DBUserID, item.ID)
		if err != nil {
			s.fail(w, r, err, "list challenge days failed")
			return
		}
		// The streaks are computed by the service rather than here: it already
		// owns that arithmetic for the bot, and a second implementation would
		// eventually disagree with the first.
		detail, err := s.challengesvc.GetDayDetail(r.Context(), user.DBUserID, item.ID, now, now)
		if err != nil {
			s.fail(w, r, err, "challenge day detail failed")
			return
		}
		out.Challenges = append(out.Challenges,
			apidto.FromChallenge(item, days, detail.CurrentStreak, detail.BestStreak))
	}

	writeJSON(w, r, http.StatusOK, out)
}

// handleChallengeDay marks one day.
//
// A PUT carrying the state, like the roadmap card: the request can arrive twice
// and must land in the same place. The service maps done=false to "skipped",
// which is the bot's own vocabulary — there is deliberately no way back to
// "pending", because forgetting a day and un-deciding it are different things
// and only the first is a state the schema keeps.
func (s *Server) handleChallengeDay(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	challengeID, perr := pathID(r, "id")
	if perr != nil {
		s.failRequest(w, r, perr, "challenge id")
		return
	}

	rawDate := r.PathValue("date")
	day, err := apptime.ParseDay(rawDate, user.Location)
	if err != nil {
		s.failRequest(w, r, badParam("date must be YYYY-MM-DD, got %q", rawDate), "challenge day")
		return
	}

	var body struct {
		Done *bool `json:"done"`
	}
	if err := decodeBody(r, &body); err != nil {
		s.failRequest(w, r, err, "challenge day body")
		return
	}
	if body.Done == nil {
		s.failRequest(w, r, badParam(`body must be {"done": true} or {"done": false}`), "challenge day body")
		return
	}

	if err := s.challengesvc.MarkDay(r.Context(), user.DBUserID, challengeID, day, *body.Done); err != nil {
		if errors.Is(err, models.ErrChallengeDayNotFound) {
			// Also the answer for a day outside the challenge, or someone
			// else's challenge: which it is, is not worth confirming.
			writeErr(w, r, http.StatusNotFound, codeNotFound, "день не найден")
			return
		}
		s.fail(w, r, err, "mark challenge day failed")
		return
	}

	status := string(models.ChallengeDaySkipped)
	if *body.Done {
		status = string(models.ChallengeDayDone)
	}
	writeJSON(w, r, http.StatusOK, apidto.ChallengeDayStateResponse{
		ChallengeID: challengeID,
		Date:        apidto.Date(day.Format("2006-01-02")),
		Status:      status,
	})
}
