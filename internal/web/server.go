// Package web serves the read-only dashboard API.
//
// It runs inside the bot process on its own goroutine, the same shape as the
// four schedulers in internal/application: NewServer captures the app context,
// Run starts the listener, and Stop drains it. That means it must never call
// log.Fatal after startup and never let a panic escape a handler — see
// withRecover.
//
// The package deliberately imports only config, service, models and pkg/*.
// Nothing from handlers, dispatcher, buttons, i18n or repo may appear here: the
// seam is what makes moving this into its own binary a small change later.
package web

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"

	"tracker-bot/internal/config"
	"tracker-bot/internal/service"
	"tracker-bot/internal/web/tgauth"
)

type Server struct {
	cfg  config.Web
	http *http.Server
	// ctx is the application context, captured at construction like the
	// schedulers do; Run stops when it is cancelled.
	ctx context.Context

	// verifier is nil only in dev-bypass mode, where nothing is verified.
	verifier *tgauth.Verifier
	identity *Identity
	tracksvc service.TrackerService
}

// NewServer builds the listener but does not start it.
//
// botToken signs nothing here — it is the key init data is verified against, so
// the dashboard authenticates against the same bot the user is talking to.
func NewServer(
	ctx context.Context,
	cfg config.Web,
	botToken string,
	entrysvc service.EntryService,
	profilesvc service.ProfileService,
	tracksvc service.TrackerService,
) (*Server, error) {
	if ctx == nil {
		return nil, errors.New("web: nil context")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if entrysvc == nil || profilesvc == nil || tracksvc == nil {
		return nil, errors.New("web: entry, profile and tracker services are required")
	}

	srv := &Server{
		cfg:      cfg,
		ctx:      ctx,
		identity: NewIdentity(entrysvc, profilesvc),
		tracksvc: tracksvc,
	}

	if cfg.DevTgUserID != 0 {
		// Loud on every start: this is the setting that turns authentication
		// off, and it should never be a surprise in a log.
		log.Warn().
			Int64("tg_user_id", cfg.DevTgUserID).
			Msg("dashboard auth is BYPASSED — every request is served as this user")
	} else {
		verifier, err := tgauth.NewVerifier(botToken, cfg.InitDataMaxAge)
		if err != nil {
			return nil, err
		}
		srv.verifier = verifier
	}
	srv.http = &http.Server{
		Addr:    cfg.Addr,
		Handler: srv.routes(),
		// Read timeouts can be short: every request is a GET with no body.
		// Writes get longer, to cover the slowest aggregate query reaching a
		// phone on a bad connection. Idle outlives a Mini App session so
		// keep-alive survives a user reading for a minute.
		ReadHeaderTimeout: cfg.ReadTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	return srv, nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Unauthenticated and outside /api on purpose: this is what a deploy check
	// curls, and it must answer before any Telegram credential exists.
	mux.HandleFunc("GET /healthz", s.handleHealth)

	// Everything under /api needs a verified launch. Registered through one
	// api() wrapper so a new endpoint cannot accidentally be added unprotected.
	mux.Handle("GET /api/v1/me", s.api(s.handleMe))
	mux.Handle("GET /api/v1/track/overview", s.api(s.handleTrackOverview))
	mux.Handle("GET /api/v1/track/activities", s.api(s.handleTrackActivities))
	mux.Handle("GET /api/v1/track/breakdown", s.api(s.handleTrackBreakdown))
	mux.Handle("GET /api/v1/track/series", s.api(s.handleTrackSeries))
	mux.Handle("GET /api/v1/track/heatmap", s.api(s.handleTrackHeatmap))
	mux.Handle("GET /api/v1/track/day", s.api(s.handleTrackDay))

	// Anything unrouted answers JSON rather than net/http's text 404, so the
	// frontend's error path never has to parse two shapes.
	mux.HandleFunc("/", s.handleNotFound)

	// Outermost first: an id exists before anything logs, logging sees the real
	// status a recovered panic produced, and the concurrency limit sits inside
	// recover so a rejected request is still logged.
	var h http.Handler = mux
	h = withConcurrencyLimit(s.cfg.MaxInflight, h)
	h = withRecover(h)
	h = withLogging(h)
	h = withRequestID(h)
	return h
}

// api wraps a handler in authentication. Every /api route goes through it.
func (s *Server) api(h http.HandlerFunc) http.Handler {
	return s.withAuth(h)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeErr(w, r, http.StatusNotFound, codeNotFound, "no such endpoint")
}

// Run serves until the application context is cancelled, then drains. Errors
// are logged rather than returned: this runs as a background goroutine beside
// the bot, and a dead listener must not take the bot with it.
func (s *Server) Run() {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		log.Error().Err(err).Str("addr", s.cfg.Addr).Msg("web server failed to listen")
		return
	}
	log.Info().Str("addr", s.cfg.Addr).Int("max_inflight", s.cfg.MaxInflight).Msg("web server listening")

	go func() {
		<-s.ctx.Done()
		// A fresh context: the app context is already cancelled, so passing it
		// to Shutdown would abort the drain instead of bounding it.
		stopCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownGrace)
		defer cancel()
		if err := s.http.Shutdown(stopCtx); err != nil {
			log.Warn().Err(err).Msg("web server shutdown was not clean")
		}
	}()

	if err := s.http.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("web server stopped with an error")
		return
	}
	log.Info().Msg("web server stopped")
}
