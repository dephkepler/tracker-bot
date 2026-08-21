package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"tracker-bot/internal/utils/webctx"
	"tracker-bot/internal/web/apidto"
	"tracker-bot/pkg/apptime"
)

// handleTrackOverview answers the dashboard's first screen: the whole day, plus
// the activity the user last touched.
//
// Two service calls rather than one because the two halves come from different
// queries and mean different things — GetTodayReport is the day across every
// activity, GetMainStats is one activity's own figures. Merging them into a
// flat response is what made models.MainStats misleading in the first place.
func (s *Server) handleTrackOverview(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	report, err := s.tracksvc.GetTodayReport(r.Context(), user.DBUserID, user.Location)
	if err != nil {
		s.fail(w, r, err, "today report failed")
		return
	}
	main, err := s.tracksvc.GetMainStats(r.Context(), user.DBUserID, user.Location)
	if err != nil {
		s.fail(w, r, err, "main stats failed")
		return
	}

	writeJSON(w, r, http.StatusOK, apidto.OverviewResponse{
		Today:   apidto.FromTodayReport(report, main.TodaySessions),
		Current: apidto.FromMainStats(main),
		Meta:    apidto.NewMeta(user.Location, apptime.NowIn(user.Location)),
	})
}

// handleTrackActivities lists the user's activities. Archived ones are excluded
// by default and only appear on request: they are history, and every aggregate
// in this API filters them out, so mixing them into the default list would
// invite a client to display totals that can never add up.
func (s *Server) handleTrackActivities(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	items, err := s.tracksvc.ListActivities(r.Context(), user.DBUserID)
	if err != nil {
		s.fail(w, r, err, "list activities failed")
		return
	}
	activities := apidto.FromActivityItems(items, false)

	if boolParam(r, "include_archived") {
		archived, err := s.tracksvc.ListArchivedActivities(r.Context(), user.DBUserID)
		if err != nil {
			s.fail(w, r, err, "list archived activities failed")
			return
		}
		activities = append(activities, apidto.FromActivityItems(archived, true)...)
	}

	writeJSON(w, r, http.StatusOK, apidto.ActivitiesResponse{
		Activities: activities,
		Meta:       apidto.NewMeta(user.Location, apptime.NowIn(user.Location)),
	})
}

// fail logs the real error and tells the client nothing about it. Every handler
// funnels through here so no endpoint can start leaking a database message.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error, msg string) {
	log.Error().
		Str("request_id", requestIDOf(r.Context())).
		Err(err).
		Msg("web: " + msg)
	writeErr(w, r, http.StatusInternalServerError, codeInternal, "internal error")
}

// failRequest routes an error to the right status: a client mistake becomes a
// 400 carrying its own code, anything else is ours and becomes a 500 with the
// detail kept in the log.
func (s *Server) failRequest(w http.ResponseWriter, r *http.Request, err error, msg string) {
	var perr *paramError
	if errors.As(err, &perr) {
		writeErr(w, r, http.StatusBadRequest, perr.code, perr.message)
		return
	}
	s.fail(w, r, err, msg)
}

// handleTrackBreakdown answers "where did the time go over this period".
func (s *Server) handleTrackBreakdown(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	// bucketed=false: a breakdown groups by activity, not by time, so no
	// granularity applies and its range cap would be arbitrary.
	p, err := s.parseRange(r.Context(), r, user, false)
	if err != nil {
		s.failRequest(w, r, err, "breakdown params")
		return
	}

	report, err := s.tracksvc.GetPeriodReport(r.Context(), user.DBUserID, p.From, p.To, p.ActivityIDs, user.Location)
	if err != nil {
		s.failRequest(w, r, err, "period report failed")
		return
	}

	writeJSON(w, r, http.StatusOK, apidto.FromPeriodReport(report, s.meta(user, p)))
}

// handleTrackSeries answers "how did it change over time".
func (s *Server) handleTrackSeries(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	p, err := s.parseRange(r.Context(), r, user, true)
	if err != nil {
		s.failRequest(w, r, err, "series params")
		return
	}

	switch by := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("by"))); by {
	case "", "total":
		starts, durations, err := s.tracksvc.GetPeriodBuckets(
			r.Context(), user.DBUserID, p.From, p.To, p.ActivityIDs, p.Granularity, user.Location)
		if err != nil {
			s.failRequest(w, r, err, "period buckets failed")
			return
		}
		writeJSON(w, r, http.StatusOK, apidto.FromBuckets(starts, durations, p.Granularity, s.meta(user, p)))

	case "activity":
		// Only hourly for now. A per-activity daily series needs
		// GetHourlyBucketsByActivity generalised to a granularity parameter,
		// which is real new SQL rather than a mapping change.
		if p.Granularity != granHour {
			writeErr(w, r, http.StatusBadRequest, codeUnsupportedCombination,
				"by=activity is only available at hour granularity")
			return
		}
		rows, err := s.tracksvc.GetHourlyBucketsByActivity(
			r.Context(), user.DBUserID, p.From, p.To, p.ActivityIDs, user.Location)
		if err != nil {
			s.failRequest(w, r, err, "hourly buckets failed")
			return
		}
		writeJSON(w, r, http.StatusOK, apidto.FromHourlyByActivity(rows, s.meta(user, p)))

	default:
		writeErr(w, r, http.StatusBadRequest, codeInvalidParameter,
			"unknown by="+by+", expected total or activity")
	}
}

// handleTrackHeatmap answers "how unbroken has this been".
//
// Backed by the daily buckets rather than the distinct-tracked-days query: that
// one returns only whether a day had any session, which is a heatmap with no
// intensity. The per-day totals come from a query that already exists.
func (s *Server) handleTrackHeatmap(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	// weeks is the heatmap's own shorthand, since its window is a fixed grid
	// rather than the period the rest of the section is looking at.
	if !r.URL.Query().Has("from") && !r.URL.Query().Has("to") {
		weeks, perr := heatmapWeeks(r.URL.Query().Get("weeks"))
		if perr != nil {
			s.failRequest(w, r, perr, "heatmap params")
			return
		}
		now := apptime.NowIn(user.Location)
		q := r.URL.Query()
		q.Set("from", now.AddDate(0, 0, -(weeks*7-1)).Format("2006-01-02"))
		q.Set("to", now.Format("2006-01-02"))
		r.URL.RawQuery = q.Encode()
	}

	p, err := s.parseRange(r.Context(), r, user, true)
	if err != nil {
		s.failRequest(w, r, err, "heatmap params")
		return
	}
	// The grid is days by definition, whatever auto would have picked.
	p.Granularity = granDay

	starts, durations, err := s.tracksvc.GetPeriodBuckets(
		r.Context(), user.DBUserID, p.From, p.To, p.ActivityIDs, granDay, user.Location)
	if err != nil {
		s.failRequest(w, r, err, "daily buckets failed")
		return
	}

	writeJSON(w, r, http.StatusOK, apidto.FromDailyBuckets(starts, durations, s.meta(user, p)))
}

const (
	heatmapDefaultWeeks = 8
	heatmapMaxWeeks     = 52
)

func heatmapWeeks(raw string) (int, *paramError) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return heatmapDefaultWeeks, nil
	}
	weeks, err := strconv.Atoi(raw)
	if err != nil || weeks <= 0 {
		return 0, badParam("weeks must be a positive number, got %q", raw)
	}
	if weeks > heatmapMaxWeeks {
		return 0, badParam("weeks must be at most %d", heatmapMaxWeeks)
	}
	return weeks, nil
}

// handleTrackDay answers "what did one day look like" — the drill-down from a
// heatmap cell.
func (s *Server) handleTrackDay(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		s.failRequest(w, r, badParam("date is required, as YYYY-MM-DD"), "day params")
		return
	}
	// One day, expressed as the range the rest of the code already understands.
	q := r.URL.Query()
	q.Set("from", date)
	q.Set("to", date)
	q.Del("period")
	r.URL.RawQuery = q.Encode()

	p, err := s.parseRange(r.Context(), r, user, true)
	if err != nil {
		s.failRequest(w, r, err, "day params")
		return
	}

	report, err := s.tracksvc.GetPeriodReport(r.Context(), user.DBUserID, p.From, p.To, p.ActivityIDs, user.Location)
	if err != nil {
		s.failRequest(w, r, err, "day report failed")
		return
	}
	rows, err := s.tracksvc.GetHourlyBucketsByActivity(
		r.Context(), user.DBUserID, p.From, p.To, p.ActivityIDs, user.Location)
	if err != nil {
		s.failRequest(w, r, err, "day hours failed")
		return
	}

	meta := s.meta(user, p)
	hours := apidto.FromHourlyByActivity(rows, meta)
	writeJSON(w, r, http.StatusOK, apidto.DayResponse{
		Date:          apidto.Date(date),
		TotalSeconds:  apidto.Seconds(report.TotalTracked),
		TotalSessions: report.TotalSessions,
		Activities:    apidto.FromActivityStats(report.Activities, report.TotalTracked),
		Hours:         hours.Buckets,
		Meta:          meta,
	})
}

// meta stamps a response with the window the server actually used.
func (s *Server) meta(user webctx.User, p rangeParams) apidto.Meta {
	m := apidto.NewMeta(user.Location, apptime.NowIn(user.Location))
	m.From = apidto.Date(p.FromDay)
	m.To = apidto.Date(p.ToDay)
	m.Granularity = p.Granularity
	m.ActivityIDs = p.ActivityIDs
	return m
}
