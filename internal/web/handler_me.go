package web

import (
	"net/http"

	"tracker-bot/internal/utils/webctx"
	"tracker-bot/internal/web/apidto"
	"tracker-bot/pkg/apptime"
)

// handleMe answers who the request resolved to. Its real job is diagnostic: it
// is the one endpoint that proves the whole chain — a signed launch from
// Telegram, verified here, mapped to a row in this database, with the timezone
// every other endpoint will bucket by.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())
	writeJSON(w, r, http.StatusOK, apidto.NewMeResponse(
		user.TgUserID, user.Location, user.Language, apptime.NowIn(user.Location),
	))
}
