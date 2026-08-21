package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tracker-bot/internal/utils/webctx"
	"tracker-bot/pkg/apptime"
)

// Granularities the repository accepts. Kept here as well because the API takes
// "auto" too, which the repository knows nothing about.
const (
	granHour  = "hour"
	granDay   = "day"
	granMonth = "month"
)

// Range caps, by granularity. These bound the response size rather than the
// query cost: a year at hourly granularity is 8760 buckets, which is not a
// chart anyone can read and not a payload worth sending to a phone.
var rangeCapDays = map[string]int{
	granHour:  31,
	granDay:   366,
	granMonth: 1826,
}

// paramError is a client mistake: it becomes a 400 with a stable code the
// frontend can branch on, never a 500.
type paramError struct {
	code    string
	message string
}

func (e *paramError) Error() string { return e.message }

func badParam(format string, args ...any) *paramError {
	return &paramError{code: codeInvalidParameter, message: fmt.Sprintf(format, args...)}
}

// rangeParams is one parsed time window plus what to aggregate over it.
type rangeParams struct {
	// FromDay and ToDay are the inclusive local day bounds, echoed back in the
	// response so a client that asked for "week" learns which week it got.
	FromDay, ToDay string
	// From and To are the half-open instant range [From, To) the services take.
	From, To time.Time
	// Granularity is resolved: never "auto" by the time it leaves here.
	Granularity string
	// ActivityIDs is always explicit. An empty slice would mean "nothing" to
	// every repository query, not "everything".
	ActivityIDs []int64
}

// parseRange reads the period, activity filter and granularity off a request.
//
// bucketed says whether this endpoint aggregates into buckets; when it does not
// (a plain breakdown) the granularity is left empty and its cap does not apply.
// A *paramError result is the client's mistake and becomes a 400; any other
// error came from a service and becomes a 500.
func (s *Server) parseRange(ctx context.Context, r *http.Request, user webctx.User, bucketed bool) (rangeParams, error) {
	q := r.URL.Query()
	loc := user.Location
	now := apptime.NowIn(loc)

	fromDay, toDay, perr := periodDays(q, now)
	if perr != nil {
		return rangeParams{}, perr
	}

	from, to, err := localRange(fromDay, toDay, loc)
	if err != nil {
		return rangeParams{}, badParam("%v", err)
	}

	out := rangeParams{FromDay: fromDay, ToDay: toDay, From: from, To: to}

	if bucketed {
		gran, perr := resolveGranularity(q.Get("granularity"), fromDay, toDay, loc)
		if perr != nil {
			return rangeParams{}, perr
		}
		out.Granularity = gran

		days := int(to.Sub(from).Hours()/24 + 0.5)
		if cap, ok := rangeCapDays[gran]; ok && days > cap {
			return rangeParams{}, &paramError{
				code:    codeRangeTooLarge,
				message: fmt.Sprintf("%d days is too long for %s granularity, the limit is %d", days, gran, cap),
			}
		}
	}

	ids, err := s.resolveActivityIDs(ctx, user, q)
	if err != nil {
		return rangeParams{}, err
	}
	out.ActivityIDs = ids

	return out, nil
}

// periodDays turns either a named period or an explicit from/to pair into
// inclusive local day strings.
func periodDays(q map[string][]string, now time.Time) (string, string, *paramError) {
	period := strings.ToLower(strings.TrimSpace(first(q, "period")))
	fromRaw, toRaw := strings.TrimSpace(first(q, "from")), strings.TrimSpace(first(q, "to"))

	// An explicit range wins, and is also what "custom" means.
	if period == "" && fromRaw == "" && toRaw == "" {
		period = "today"
	}
	if fromRaw != "" || toRaw != "" || period == "custom" {
		if fromRaw == "" || toRaw == "" {
			return "", "", badParam("from and to are both required for a custom period")
		}
		return fromRaw, toRaw, nil
	}

	day := func(t time.Time) string { return t.Format("2006-01-02") }
	switch period {
	case "today":
		return day(now), day(now), nil
	case "week":
		// Trailing seven days including today, not the calendar week: the
		// dashboard answers "how have I been doing lately", and a calendar week
		// answers that badly on a Monday.
		return day(now.AddDate(0, 0, -6)), day(now), nil
	case "month":
		return day(now.AddDate(0, 0, -29)), day(now), nil
	case "year":
		return day(now.AddDate(0, 0, -364)), day(now), nil
	default:
		return "", "", badParam("unknown period %q, expected today, week, month, year or custom", period)
	}
}

// localRange converts inclusive YYYY-MM-DD bounds interpreted in loc into a
// half-open instant range.
//
// Every range endpoint must get its bounds from here. repo.GetPeriodActivities
// filters on absolute instants with no timezone of its own, so a boundary built
// at UTC midnight instead of the user's would silently move a whole day of
// sessions in or out of the period.
func localRange(fromDay, toDay string, loc *time.Location) (time.Time, time.Time, error) {
	from, err := apptime.ParseDay(fromDay, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("from %q is not a YYYY-MM-DD date", fromDay)
	}
	to, err := apptime.ParseDay(toDay, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("to %q is not a YYYY-MM-DD date", toDay)
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("to %q is before from %q", toDay, fromDay)
	}
	// AddDate rather than adding 24h: across a DST change a day is not 24 hours,
	// and the exclusive end has to land on the next local midnight either way.
	return from, to.AddDate(0, 0, 1), nil
}

// resolveGranularity expands "auto" the same way the bot's own report does
// (see appendGranularityText in internal/handlers): a single day is worth
// showing by hour, a range crossing years by month, anything else by day.
func resolveGranularity(requested, fromDay, toDay string, loc *time.Location) (string, *paramError) {
	switch strings.ToLower(strings.TrimSpace(requested)) {
	case granHour, granDay, granMonth:
		return strings.ToLower(strings.TrimSpace(requested)), nil
	case "", "auto":
	default:
		return "", badParam("unknown granularity %q, expected auto, hour, day or month", requested)
	}

	from, err := apptime.ParseDay(fromDay, loc)
	if err != nil {
		return "", badParam("from %q is not a YYYY-MM-DD date", fromDay)
	}
	to, err := apptime.ParseDay(toDay, loc)
	if err != nil {
		return "", badParam("to %q is not a YYYY-MM-DD date", toDay)
	}

	switch {
	case fromDay == toDay:
		return granHour, nil
	case from.Year() != to.Year():
		return granMonth, nil
	default:
		return granDay, nil
	}
}

// resolveActivityIDs expands an absent filter into the user's active
// activities.
//
// This is load-bearing, not a convenience: GetPeriodActivities,
// GetPeriodMonthlyTotals, GetPeriodBuckets and GetHourlyBucketsByActivity all
// return nothing at all for an empty id slice. Passing the parameter straight
// through would make an unfiltered dashboard show zeros.
func (s *Server) resolveActivityIDs(ctx context.Context, user webctx.User, q map[string][]string) ([]int64, error) {
	// A present-but-empty value counts as absent. A query builder that has
	// nothing selected naturally emits "activity_ids=", and refusing that would
	// turn a harmless client habit into a failed request. A client that means
	// "no activities" should simply not ask for a breakdown.
	raw := strings.TrimSpace(first(q, "activity_ids"))
	if raw == "" {
		items, err := s.tracksvc.ListActivities(ctx, user.DBUserID)
		if err != nil {
			// A service failure, not a bad request: returned as-is so the
			// handler answers 500 rather than blaming the client.
			return nil, err
		}
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		return ids, nil
	}

	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, badParam("activity_ids contains %q, which is not an activity id", part)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		// Reachable for a value that is all separators, like ",,," — which is
		// a malformed list rather than an empty one.
		return nil, badParam("activity_ids has no ids in it")
	}
	return ids, nil
}

func first(q map[string][]string, key string) string {
	if v, ok := q[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

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
