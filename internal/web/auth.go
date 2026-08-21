package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"tracker-bot/internal/models"
	"tracker-bot/internal/utils/webctx"
	"tracker-bot/internal/web/tgauth"
)

// authScheme is the scheme Telegram's own SDKs converged on for handing raw
// init data to a backend. It is not a bearer token — nothing here was issued to
// us, it is a signed snapshot of the launch.
const authScheme = "tma"

// withAuth verifies the launch and puts the resolved user in the request
// context.
//
// The credential travels in a header, never a query parameter: it stays valid
// for the whole init-data window, and a URL would copy it into Caddy's access
// log, the browser history and any Referer the page sends.
func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tgUserID, ok := s.authenticate(w, r)
		if !ok {
			return
		}

		user, err := s.identity.Resolve(r.Context(), tgUserID)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				// A real Telegram user with no row here. Not 403: the
				// credential is fine, the account simply does not exist yet.
				writeErr(w, r, http.StatusNotFound, codeUserNotFound,
					"open the bot and press Start first")
				return
			}
			log.Error().
				Str("request_id", requestIDOf(r.Context())).
				Int64("tg_user_id", tgUserID).
				Err(err).
				Msg("web identity lookup failed")
			writeErr(w, r, http.StatusInternalServerError, codeInternal, "internal error")
			return
		}

		next.ServeHTTP(w, r.WithContext(webctx.With(r.Context(), user)))
	})
}

// authenticate resolves the Telegram id the request speaks for, answering the
// client itself on failure.
func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) (int64, bool) {
	// The dev bypass exists so the dashboard can be opened in an ordinary
	// browser, where there is no Telegram to sign anything. config.Web.Validate
	// refuses to start with this set alongside an https origin, so it cannot be
	// reached in production.
	if s.cfg.DevTgUserID != 0 {
		return s.cfg.DevTgUserID, true
	}

	raw, ok := initDataFromHeader(r.Header.Get("Authorization"))
	if !ok {
		writeErr(w, r, http.StatusUnauthorized, codeMissingCredentials,
			"open this from the bot")
		return 0, false
	}

	data, err := s.verifier.Verify(raw)
	if err != nil {
		status, code := http.StatusUnauthorized, codeUnauthorized
		switch {
		case errors.Is(err, tgauth.ErrStale), errors.Is(err, tgauth.ErrFuture):
			code = codeInitDataExpired
		}
		// Logged, not returned: which check failed is useful to us and useful
		// to someone probing.
		log.Warn().
			Str("request_id", requestIDOf(r.Context())).
			Err(err).
			Msg("web init data rejected")
		writeErr(w, r, status, code, "this launch is no longer valid, reopen the app")
		return 0, false
	}

	return data.User.ID, true
}

// initDataFromHeader pulls the init data out of an "Authorization: tma <data>"
// header. The scheme is compared case-insensitively, as RFC 9110 requires.
func initDataFromHeader(header string) (string, bool) {
	scheme, rest, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found || !strings.EqualFold(scheme, authScheme) {
		return "", false
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", false
	}
	return rest, true
}
