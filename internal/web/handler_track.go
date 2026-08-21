package web

import (
	"net/http"

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
