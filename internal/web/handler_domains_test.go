package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

type learningBody struct {
	TotalWords      int     `json:"total_words"`
	DueWords        int     `json:"due_words"`
	LearnedWords    int     `json:"learned_words"`
	StreakDays      int     `json:"streak_days"`
	ReviewsTotal    int     `json:"reviews_total"`
	ReviewsCorrect  int     `json:"reviews_correct"`
	AccuracyPercent float64 `json:"accuracy_percent"`
	Collections     []struct {
		Name       string `json:"name"`
		TotalWords int    `json:"total_words"`
	} `json:"collections"`
	ReviewsByDay []struct {
		Date    string `json:"date"`
		Total   int    `json:"total"`
		Correct int    `json:"correct"`
	} `json:"reviews_by_day"`
	ReminderActive   bool     `json:"reminder_active"`
	ReminderInterval int      `json:"reminder_interval_minutes"`
	Meta             metaBody `json:"meta"`
}

func TestLearning(t *testing.T) {
	deps := testDeps()
	learningsvc := deps.Learning.(*fakeLearningSvc)
	srv, err := NewServer(t.Context(), testWebConfig(), deps)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	body := getJSON[learningBody](t, srv, "/api/v1/learning?days=7")

	if body.TotalWords != 120 || body.DueWords != 12 || body.LearnedWords != 40 || body.StreakDays != 5 {
		t.Fatalf("counters = %+v", body)
	}
	// 261 of 300, to one decimal.
	if body.AccuracyPercent != 87 {
		t.Fatalf("accuracy = %v, want 87", body.AccuracyPercent)
	}
	if len(body.Collections) != 2 {
		t.Fatalf("collections = %+v", body.Collections)
	}
	// The reminder is a number plus a flag, never the service's pre-formatted
	// "17 min" — that string is built for a Telegram message.
	if !body.ReminderActive || body.ReminderInterval != 60 {
		t.Fatalf("reminder = %v/%d", body.ReminderActive, body.ReminderInterval)
	}

	// A window of seven days is seven points, gaps included: a missed day is
	// the interesting part of a review habit and must not be closed up.
	if len(body.ReviewsByDay) != 7 {
		t.Fatalf("got %d days, want 7", len(body.ReviewsByDay))
	}
	var withReviews, total int
	for _, d := range body.ReviewsByDay {
		if d.Total > 0 {
			withReviews++
		}
		total += d.Total
	}
	if total != 3 {
		t.Fatalf("counted %d reviews, want 3", total)
	}
	if withReviews != 2 {
		t.Fatalf("%d days have reviews, want 2 — the third day is a gap", withReviews)
	}

	// The window is built from the user's own midnights, so its bounds are not
	// at 00:00 UTC.
	if learningsvc.gotFrom.Hour() == 0 && learningsvc.gotFrom.Location() == time.UTC {
		t.Fatalf("window starts at UTC midnight: %s", learningsvc.gotFrom)
	}
	if learningsvc.gotLoc == nil || learningsvc.gotLoc.String() != "America/New_York" {
		t.Fatalf("service location = %v", learningsvc.gotLoc)
	}
}

// The pre-formatted NextPushIn must not reach the wire under any name.
func TestLearningDoesNotLeakPresentationStrings(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	rec := call(t, srv, "/api/v1/learning", authHeader(t, knownTgUserID))
	if strings.Contains(rec.Body.String(), "17 min") {
		t.Fatalf("a Telegram-formatted string reached the API: %s", rec.Body.String())
	}
}

func TestLearningRejectsBadDays(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	for _, q := range []string{"?days=0", "?days=-3", "?days=abc", "?days=5000"} {
		rec := call(t, srv, "/api/v1/learning"+q, authHeader(t, knownTgUserID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400", q, rec.Code)
		}
	}
}

type challengesBody struct {
	Challenges []struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		StartDate     string `json:"start_date"`
		EndDate       string `json:"end_date"`
		TotalDays     int    `json:"total_days"`
		DoneDays      int    `json:"done_days"`
		SkippedDays   int    `json:"skipped_days"`
		PendingDays   int    `json:"pending_days"`
		CurrentStreak int    `json:"current_streak"`
		BestStreak    int    `json:"best_streak"`
		Days          []struct {
			Date   string `json:"date"`
			Status string `json:"status"`
		} `json:"days"`
	} `json:"challenges"`
}

func TestChallenges(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)
	body := getJSON[challengesBody](t, srv, "/api/v1/challenges")

	if len(body.Challenges) != 1 {
		t.Fatalf("got %d challenges, want 1", len(body.Challenges))
	}
	ch := body.Challenges[0]

	// Bare dates: challenge_days.day_date is a zoneless DATE, so sending an
	// instant would invite a client to shift the square by a day.
	if ch.StartDate != "2026-08-19" || ch.EndDate != "2026-08-22" {
		t.Fatalf("dates = %q…%q", ch.StartDate, ch.EndDate)
	}
	if len(ch.Days) != 4 {
		t.Fatalf("got %d days, want the whole grid", len(ch.Days))
	}
	if ch.Days[0].Status != "done" || ch.Days[1].Status != "skipped" || ch.Days[3].Status != "pending" {
		t.Fatalf("statuses = %+v", ch.Days)
	}
	// Derived server-side so two clients cannot disagree.
	if ch.PendingDays != 1 {
		t.Fatalf("pending = %d, want 1", ch.PendingDays)
	}
	// Streaks come from the service, which already owns that arithmetic.
	if ch.CurrentStreak != 1 || ch.BestStreak != 2 {
		t.Fatalf("streaks = %d/%d", ch.CurrentStreak, ch.BestStreak)
	}
}

func TestChallengeDayMarking(t *testing.T) {
	deps := testDeps()
	challengesvc := deps.Challenge.(*fakeChallengeSvc)
	srv, err := NewServer(t.Context(), testWebConfig(), deps)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	// Twice with the same body: it carries the state wanted, so the second
	// request is not an undo.
	for range 2 {
		rec := send(t, srv, http.MethodPut, "/api/v1/challenges/5/days/2026-08-22", `{"done":true}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (body %q)", rec.Code, rec.Body.String())
		}
		var body struct {
			Status string `json:"status"`
			Date   string `json:"date"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Status != "done" || body.Date != "2026-08-22" {
			t.Fatalf("body = %+v", body)
		}
	}
	if len(challengesvc.marked) != 2 {
		t.Fatalf("%d writes, want 2", len(challengesvc.marked))
	}

	// done=false is the bot's "skipped", not a return to pending.
	rec := send(t, srv, http.MethodPut, "/api/v1/challenges/5/days/2026-08-20", `{"done":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "skipped" {
		t.Fatalf("status = %q, want skipped", body.Status)
	}
}

func TestChallengeDayRejectsBadRequests(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	cases := map[string]struct {
		path, body string
		want       int
	}{
		"bad date":      {"/api/v1/challenges/5/days/22-08-2026", `{"done":true}`, http.StatusBadRequest},
		"no body":       {"/api/v1/challenges/5/days/2026-08-22", "", http.StatusBadRequest},
		"missing done":  {"/api/v1/challenges/5/days/2026-08-22", `{}`, http.StatusBadRequest},
		"unknown field": {"/api/v1/challenges/5/days/2026-08-22", `{"status":"done"}`, http.StatusBadRequest},
		"bad id":        {"/api/v1/challenges/x/days/2026-08-22", `{"done":true}`, http.StatusBadRequest},
		"day outside":   {"/api/v1/challenges/5/days/2026-09-01", `{"done":true}`, http.StatusNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := send(t, srv, http.MethodPut, tc.path, tc.body)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestDomainEndpointsNeedAuth(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	for _, path := range []string{"/api/v1/learning", "/api/v1/challenges"} {
		if rec := call(t, srv, path, ""); rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s: status = %d, want 401", path, rec.Code)
		}
	}
}
