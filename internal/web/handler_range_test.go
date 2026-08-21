package web

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func getJSON[T any](t *testing.T, srv *Server, path string) T {
	t.Helper()
	rec := call(t, srv, path, authHeader(t, knownTgUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: status = %d, want 200 (body %q)", path, rec.Code, rec.Body.String())
	}
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s: decode: %v (body %q)", path, err, rec.Body.String())
	}
	return out
}

type metaBody struct {
	Timezone    string  `json:"timezone"`
	From        string  `json:"from"`
	To          string  `json:"to"`
	Granularity string  `json:"granularity"`
	ActivityIDs []int64 `json:"activity_ids"`
}

type breakdownBody struct {
	TotalSeconds  int64 `json:"total_seconds"`
	TotalSessions int   `json:"total_sessions"`
	Activities    []struct {
		Name         string  `json:"name"`
		Seconds      int64   `json:"seconds"`
		SharePercent float64 `json:"share_percent"`
	} `json:"activities"`
	Monthly []struct {
		Month   string `json:"month"`
		Seconds int64  `json:"seconds"`
	} `json:"monthly"`
	Meta metaBody `json:"meta"`
}

type seriesBody struct {
	By      string `json:"by"`
	Buckets []struct {
		Start   string `json:"start"`
		Seconds int64  `json:"seconds"`
		Parts   []struct {
			Name    string `json:"name"`
			Seconds int64  `json:"seconds"`
		} `json:"parts"`
	} `json:"buckets"`
	Meta metaBody `json:"meta"`
}

func TestBreakdown(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	body := getJSON[breakdownBody](t, srv, "/api/v1/track/breakdown?period=month")

	if body.TotalSeconds != 4*3600 {
		t.Fatalf("total_seconds = %d, want 14400", body.TotalSeconds)
	}
	if body.Activities[0].SharePercent != 75 {
		t.Fatalf("first share = %v, want 75", body.Activities[0].SharePercent)
	}
	// A month bucket must stay a bare calendar date. Emitting it as an instant
	// invites the client to shift it into the previous month.
	if body.Monthly[0].Month != "2026-07-01" {
		t.Fatalf("monthly[0].month = %q, want 2026-07-01", body.Monthly[0].Month)
	}
	// An omitted filter has to arrive at the service as an explicit list: every
	// range query in the repository treats an empty slice as "nothing".
	if len(tracksvc.gotActivityIDs) != 2 {
		t.Fatalf("service got %d activity ids, want the 2 active ones expanded", len(tracksvc.gotActivityIDs))
	}
	if len(body.Meta.ActivityIDs) != 2 {
		t.Fatalf("meta.activity_ids = %v, want the expansion echoed back", body.Meta.ActivityIDs)
	}
}

// The window has to be built from the user's own midnights. The fake's user is
// in New York, so a day starts at 04:00 or 05:00 UTC — never at 00:00 UTC.
func TestRangeBoundsUseTheUsersMidnight(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	getJSON[breakdownBody](t, srv, "/api/v1/track/breakdown?from=2026-08-20&to=2026-08-20")

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load zone: %v", err)
	}
	wantFrom := time.Date(2026, time.August, 20, 0, 0, 0, 0, loc)
	wantTo := time.Date(2026, time.August, 21, 0, 0, 0, 0, loc)

	if !tracksvc.gotFrom.Equal(wantFrom) {
		t.Fatalf("from = %s, want %s", tracksvc.gotFrom, wantFrom)
	}
	// Half-open: the end is the next local midnight, so the whole 20th is in
	// and nothing of the 21st is.
	if !tracksvc.gotTo.Equal(wantTo) {
		t.Fatalf("to = %s, want %s", tracksvc.gotTo, wantTo)
	}
}

// Across a DST change a local day is not 24 hours long, so the exclusive end
// must be built by calendar arithmetic rather than by adding a fixed duration.
func TestRangeSpansADSTChange(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	// New York leaves DST on 2026-11-01, so that local day lasts 25 hours.
	getJSON[breakdownBody](t, srv, "/api/v1/track/breakdown?from=2026-11-01&to=2026-11-01")

	got := tracksvc.gotTo.Sub(tracksvc.gotFrom)
	if got != 25*time.Hour {
		t.Fatalf("the DST day spans %v, want 25h — the end was not built from the calendar", got)
	}
}

func TestSeriesGranularityAuto(t *testing.T) {
	srv, tracksvc := newTrackServer(t)

	cases := []struct {
		query string
		want  string
	}{
		// A single day is worth showing by hour.
		{"from=2026-08-21&to=2026-08-21", "hour"},
		// A stretch inside one year, by day.
		{"from=2026-08-01&to=2026-08-21", "day"},
		// Crossing a year boundary, by month.
		{"from=2025-12-01&to=2026-08-21", "month"},
	}
	for _, tc := range cases {
		body := getJSON[seriesBody](t, srv, "/api/v1/track/series?"+tc.query)
		if body.Meta.Granularity != tc.want {
			t.Fatalf("%s: granularity = %q, want %q", tc.query, body.Meta.Granularity, tc.want)
		}
		if tracksvc.gotGranularity != tc.want {
			t.Fatalf("%s: service got granularity %q, want %q", tc.query, tracksvc.gotGranularity, tc.want)
		}
	}
}

func TestSeriesRendersBucketStartsByGranularity(t *testing.T) {
	srv, _ := newTrackServer(t)

	daily := getJSON[seriesBody](t, srv, "/api/v1/track/series?from=2026-08-19&to=2026-08-21&granularity=day")
	// A day bucket is a bare date: no time, no offset, nothing to shift.
	if daily.Buckets[0].Start != "2026-08-19" {
		t.Fatalf("day bucket start = %q, want 2026-08-19", daily.Buckets[0].Start)
	}

	hourly := getJSON[seriesBody](t, srv, "/api/v1/track/series?from=2026-08-21&to=2026-08-21&granularity=hour&by=activity")
	// An hour bucket is a naive local wall clock, read together with
	// meta.timezone.
	if hourly.Buckets[0].Start != "2026-08-21T09:00" {
		t.Fatalf("hour bucket start = %q, want 2026-08-21T09:00", hourly.Buckets[0].Start)
	}
}

// The property that catches a boundary mistake: however the period is sliced,
// the slices have to add up to the same total the breakdown reports.
func TestDailyBucketsSumToTheBreakdownTotal(t *testing.T) {
	srv, _ := newTrackServer(t)

	breakdown := getJSON[breakdownBody](t, srv, "/api/v1/track/breakdown?from=2026-08-19&to=2026-08-21")
	series := getJSON[seriesBody](t, srv, "/api/v1/track/series?from=2026-08-19&to=2026-08-21&granularity=day")

	var sum int64
	for _, b := range series.Buckets {
		sum += b.Seconds
	}
	if sum != breakdown.TotalSeconds {
		t.Fatalf("buckets sum to %d but the breakdown total is %d", sum, breakdown.TotalSeconds)
	}
}

func TestSeriesByActivityGroupsIntoBuckets(t *testing.T) {
	srv, _ := newTrackServer(t)
	body := getJSON[seriesBody](t, srv, "/api/v1/track/series?from=2026-08-21&to=2026-08-21&by=activity")

	if body.By != "activity" {
		t.Fatalf("by = %q", body.By)
	}
	if len(body.Buckets) != 2 {
		t.Fatalf("got %d buckets, want 2", len(body.Buckets))
	}
	// First bucket: 45 + 15 minutes, and the repository's ranking inside the
	// bucket is preserved rather than re-sorted.
	if body.Buckets[0].Seconds != 3600 {
		t.Fatalf("first bucket = %d seconds, want 3600", body.Buckets[0].Seconds)
	}
	if len(body.Buckets[0].Parts) != 2 || body.Buckets[0].Parts[0].Name != "Go" {
		t.Fatalf("parts = %+v, want Go first", body.Buckets[0].Parts)
	}
}

// A per-activity daily series would need new SQL, so it is refused explicitly
// rather than silently answered with something else.
func TestSeriesByActivityRefusesNonHourly(t *testing.T) {
	srv, _ := newTrackServer(t)
	rec := call(t, srv, "/api/v1/track/series?from=2026-08-01&to=2026-08-21&by=activity", authHeader(t, knownTgUserID))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code := decodeErr(t, rec).Code; code != codeUnsupportedCombination {
		t.Fatalf("code = %q, want %q", code, codeUnsupportedCombination)
	}
}

func TestRangeCaps(t *testing.T) {
	srv, _ := newTrackServer(t)
	// 60 days of hourly buckets is 1440 points — not a chart, and not a payload
	// to send to a phone.
	rec := call(t, srv, "/api/v1/track/series?from=2026-06-01&to=2026-07-30&granularity=hour", authHeader(t, knownTgUserID))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code := decodeErr(t, rec).Code; code != codeRangeTooLarge {
		t.Fatalf("code = %q, want %q", code, codeRangeTooLarge)
	}
}

func TestBadParams(t *testing.T) {
	srv, _ := newTrackServer(t)

	cases := map[string]string{
		"unknown period":   "/api/v1/track/breakdown?period=fortnight",
		"from without to":  "/api/v1/track/breakdown?from=2026-08-01",
		"reversed range":   "/api/v1/track/breakdown?from=2026-08-21&to=2026-08-01",
		"unparseable date": "/api/v1/track/breakdown?from=yesterday&to=today",
		"bad granularity":  "/api/v1/track/series?period=week&granularity=fortnightly",
		"bad activity id":  "/api/v1/track/breakdown?period=week&activity_ids=5,abc",
		"comma-only ids":   "/api/v1/track/breakdown?period=week&activity_ids=,,,",
		"unknown by":       "/api/v1/track/series?period=today&by=sideways",
		"day without date": "/api/v1/track/day",
		"bad weeks":        "/api/v1/track/heatmap?weeks=0",
		"too many weeks":   "/api/v1/track/heatmap?weeks=99",
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			rec := call(t, srv, path, authHeader(t, knownTgUserID))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
			}
			// A 400 must never look like a server fault to the client.
			if code := decodeErr(t, rec).Code; code == codeInternal {
				t.Fatalf("code = %q, want a client-error code", code)
			}
		})
	}
}

// A present-but-empty filter is the same as no filter: a query builder with
// nothing selected emits exactly that, and failing the request over it would
// punish a harmless habit.
func TestEmptyActivityFilterMeansEveryActivity(t *testing.T) {
	srv, tracksvc := newTrackServer(t)
	getJSON[breakdownBody](t, srv, "/api/v1/track/breakdown?period=week&activity_ids=")

	if len(tracksvc.gotActivityIDs) != 2 {
		t.Fatalf("service got %d ids, want the 2 active ones", len(tracksvc.gotActivityIDs))
	}
}

func TestHeatmap(t *testing.T) {
	srv, tracksvc := newTrackServer(t)

	var body struct {
		Days []struct {
			Date    string `json:"date"`
			Seconds int64  `json:"seconds"`
		} `json:"days"`
		MaxSeconds int64    `json:"max_seconds"`
		Meta       metaBody `json:"meta"`
	}
	rec := call(t, srv, "/api/v1/track/heatmap?weeks=2", authHeader(t, knownTgUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (body %q)", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// The grid is days whatever auto would have chosen for a two-week window.
	if tracksvc.gotGranularity != "day" {
		t.Fatalf("service got granularity %q, want day", tracksvc.gotGranularity)
	}
	if body.Meta.Granularity != "day" {
		t.Fatalf("meta.granularity = %q, want day", body.Meta.Granularity)
	}
	if body.Days[0].Date != "2026-08-19" {
		t.Fatalf("first day = %q, want a bare date", body.Days[0].Date)
	}
	// max_seconds saves the client a second pass just to scale the ramp.
	if body.MaxSeconds != 90*60 {
		t.Fatalf("max_seconds = %d, want 5400", body.MaxSeconds)
	}
	// Two weeks ending today, inclusive.
	if body.Meta.From == "" || body.Meta.To == "" {
		t.Fatal("meta does not report the window it used")
	}
}

func TestDay(t *testing.T) {
	srv, _ := newTrackServer(t)

	var body struct {
		Date         string `json:"date"`
		TotalSeconds int64  `json:"total_seconds"`
		Activities   []struct {
			Name string `json:"name"`
		} `json:"activities"`
		Hours []struct {
			Start   string `json:"start"`
			Seconds int64  `json:"seconds"`
		} `json:"hours"`
		Meta metaBody `json:"meta"`
	}
	rec := call(t, srv, "/api/v1/track/day?date=2026-08-21", authHeader(t, knownTgUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (body %q)", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Date != "2026-08-21" {
		t.Fatalf("date = %q", body.Date)
	}
	if body.TotalSeconds == 0 || len(body.Activities) == 0 {
		t.Fatalf("day is empty: %+v", body)
	}
	if len(body.Hours) != 2 || body.Hours[0].Start != "2026-08-21T09:00" {
		t.Fatalf("hours = %+v", body.Hours)
	}
}
