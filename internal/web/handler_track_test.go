package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"tracker-bot/internal/models"
)

// newTrackServer wires a server around a tracker fake the test can inspect.
func newTrackServer(t *testing.T) (*Server, *fakeTrackSvc) {
	t.Helper()
	entrysvc, profilesvc := newFakes()
	tracksvc := newFakeTrackSvc()
	srv, err := NewServer(t.Context(), testWebConfig(), testBotToken, entrysvc, profilesvc, tracksvc)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return srv, tracksvc
}

type overviewBody struct {
	Today struct {
		TotalSeconds    int64 `json:"total_seconds"`
		Sessions        int   `json:"sessions"`
		ActivitiesCount int   `json:"activities_count"`
		TopActivities   []struct {
			ActivityID   int64   `json:"activity_id"`
			Name         string  `json:"name"`
			Emoji        string  `json:"emoji"`
			Seconds      int64   `json:"seconds"`
			SharePercent float64 `json:"share_percent"`
		} `json:"top_activities"`
	} `json:"today"`
	Current *struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		Emoji         string `json:"emoji"`
		TodaySeconds  int64  `json:"today_seconds"`
		StreakDays    int    `json:"streak_days"`
		TargetMinutes *int   `json:"target_minutes"`
	} `json:"current_activity"`
	Meta struct {
		Timezone    string `json:"timezone"`
		GeneratedAt string `json:"generated_at"`
	} `json:"meta"`
}

func getOverview(t *testing.T, srv *Server) overviewBody {
	t.Helper()
	rec := call(t, srv, "/api/v1/track/overview", authHeader(t, knownTgUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
	var body overviewBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %q)", err, rec.Body.String())
	}
	return body
}

func TestOverviewSeparatesTheDayFromTheCurrentActivity(t *testing.T) {
	srv, _ := newTrackServer(t)
	body := getOverview(t, srv)

	// The day's total across every activity.
	if body.Today.TotalSeconds != 7200 {
		t.Fatalf("today.total_seconds = %d, want 7200", body.Today.TotalSeconds)
	}
	if body.Today.Sessions != 5 {
		t.Fatalf("today.sessions = %d, want 5", body.Today.Sessions)
	}
	if body.Today.ActivitiesCount != 2 {
		t.Fatalf("today.activities_count = %d, want 2", body.Today.ActivitiesCount)
	}

	// One activity's own figures, nested so they cannot be read as the day's.
	// 42 minutes is the current activity's time today; the day total is two
	// hours. Seeing 2520 at the root would be the old MainStats bug.
	if body.Current == nil {
		t.Fatal("current_activity is null")
	}
	if body.Current.TodaySeconds != 2520 {
		t.Fatalf("current.today_seconds = %d, want 2520", body.Current.TodaySeconds)
	}
	if body.Current.StreakDays != 7 {
		t.Fatalf("current.streak_days = %d, want 7", body.Current.StreakDays)
	}
	if body.Current.Name != "Go" || body.Current.Emoji != "🐹" {
		t.Fatalf("current name/emoji = %q/%q, want them apart", body.Current.Name, body.Current.Emoji)
	}
	if body.Current.TargetMinutes == nil || *body.Current.TargetMinutes != 90 {
		t.Fatalf("current.target_minutes = %v, want 90", body.Current.TargetMinutes)
	}
}

func TestOverviewSharesAddUp(t *testing.T) {
	srv, _ := newTrackServer(t)
	body := getOverview(t, srv)

	// 90 and 30 minutes of a two-hour day.
	if got := body.Today.TopActivities[0].SharePercent; got != 75 {
		t.Fatalf("first share = %v, want 75", got)
	}
	if got := body.Today.TopActivities[1].SharePercent; got != 25 {
		t.Fatalf("second share = %v, want 25", got)
	}
}

// Every aggregate has to be bucketed by the user's own day boundaries, so the
// handler must thread the resolved zone down rather than let the service fall
// back to the server default.
func TestOverviewPassesTheUsersTimezoneToTheService(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	body := getOverview(t, srv)

	if tracksvc.gotLoc == nil {
		t.Fatal("the service was called with a nil location")
	}
	if tracksvc.gotLoc.String() != "America/New_York" {
		t.Fatalf("service location = %q, want America/New_York", tracksvc.gotLoc.String())
	}
	if body.Meta.Timezone != "America/New_York" {
		t.Fatalf("meta.timezone = %q", body.Meta.Timezone)
	}
}

// A fresh account has never tracked anything: GetMainStats returns a zero
// struct, and the client needs an explicit null to render an empty state
// instead of an activity with no name.
func TestOverviewOnAFreshAccount(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	tracksvc.main = models.MainStats{}
	tracksvc.today = models.ReportTodayStats{}

	body := getOverview(t, srv)
	if body.Current != nil {
		t.Fatalf("current_activity = %+v, want null", body.Current)
	}
	if body.Today.TotalSeconds != 0 {
		t.Fatalf("today.total_seconds = %d, want 0", body.Today.TotalSeconds)
	}
	if body.Today.TopActivities == nil {
		t.Fatal("top_activities is null; it must marshal as an empty array")
	}
}

func TestOverviewReportsServiceFailureAsInternal(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	tracksvc.err = fmt.Errorf("select: connection reset by peer")

	rec := call(t, srv, "/api/v1/track/overview", authHeader(t, knownTgUserID))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	// The database's own words must not reach the client.
	if body := rec.Body.String(); strings.Contains(body, "connection reset") {
		t.Fatalf("the underlying error leaked: %s", body)
	}
	if code := decodeErr(t, rec).Code; code != codeInternal {
		t.Fatalf("code = %q, want %q", code, codeInternal)
	}
}

func TestOverviewNeedsAuth(t *testing.T) {
	srv, _ := newTrackServer(t)

	rec := call(t, srv, "/api/v1/track/overview", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

type activitiesBody struct {
	Activities []struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		Selected      bool   `json:"selected"`
		Archived      bool   `json:"archived"`
		TargetMinutes *int   `json:"target_minutes"`
	} `json:"activities"`
}

func getActivities(t *testing.T, srv *Server, query string) activitiesBody {
	t.Helper()
	rec := call(t, srv, "/api/v1/track/activities"+query, authHeader(t, knownTgUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
	var body activitiesBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

// Archived activities are excluded by default: every aggregate filters them
// out, so listing them alongside would invite totals that cannot add up.
func TestActivitiesExcludesArchivedByDefault(t *testing.T) {
	srv, _ := newTrackServer(t)
	body := getActivities(t, srv, "")

	if len(body.Activities) != 2 {
		t.Fatalf("got %d activities, want 2", len(body.Activities))
	}
	for _, a := range body.Activities {
		if a.Archived {
			t.Fatalf("archived activity %q in the default list", a.Name)
		}
	}
}

func TestActivitiesIncludesArchivedOnRequest(t *testing.T) {
	srv, _ := newTrackServer(t)

	// A bare flag counts as true, which is how people type one by hand.
	for _, query := range []string{"?include_archived=true", "?include_archived=1", "?include_archived"} {
		body := getActivities(t, srv, query)
		if len(body.Activities) != 3 {
			t.Fatalf("%s: got %d activities, want 3", query, len(body.Activities))
		}
		var archived int
		for _, a := range body.Activities {
			if a.Archived {
				archived++
			}
		}
		if archived != 1 {
			t.Fatalf("%s: archived count = %d, want 1", query, archived)
		}
	}
}

// A mistyped optional flag must not fail the request.
func TestActivitiesIgnoresAnUnparseableFlag(t *testing.T) {
	srv, _ := newTrackServer(t)
	body := getActivities(t, srv, "?include_archived=maybe")

	if len(body.Activities) != 2 {
		t.Fatalf("got %d activities, want 2", len(body.Activities))
	}
}

func TestActivitiesCarriesTargetsAndSelection(t *testing.T) {
	srv, _ := newTrackServer(t)
	body := getActivities(t, srv, "")

	var withTarget int
	for _, a := range body.Activities {
		if a.TargetMinutes != nil {
			withTarget++
			if *a.TargetMinutes != 90 {
				t.Fatalf("target = %d, want 90", *a.TargetMinutes)
			}
		}
	}
	if withTarget != 1 {
		t.Fatalf("%d activities carry a target, want 1", withTarget)
	}
	if !body.Activities[0].Selected {
		t.Fatal("the selected flag was dropped")
	}
}
