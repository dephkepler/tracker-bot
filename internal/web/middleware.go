package web

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type ctxKey int

const ctxKeyRequestID ctxKey = iota

// requestIDOf returns the id assigned by withRequestID, or "" outside it.
func requestIDOf(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyRequestID).(string)
	return id
}

// withRequestID stamps every request with an id that goes into the logs and
// into error bodies, so a user reporting "it says something went wrong" can be
// traced to one log line without guessing.
func withRequestID(next http.Handler) http.Handler {
	var counter atomic.Uint64
	start := time.Now().UnixNano()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Process-local and monotonic rather than a UUID: a dependency for
		// something only ever read next to its own log line is not worth it.
		id := strconv.FormatInt(start, 36) + "-" + strconv.FormatUint(counter.Add(1), 36)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

// statusRecorder captures the status code for the access log. http.ResponseWriter
// gives no way to read back what was written.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		started := time.Now()

		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}
		event.
			Str("request_id", requestIDOf(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Int("bytes", rec.bytes).
			// Named with its unit and logged as an integer. zerolog's Dur writes
			// a bare float in milliseconds by default, and "latency: 6.04" was
			// read once as six seconds — a performance problem that did not
			// exist.
			Int64("latency_ms", time.Since(started).Milliseconds()).
			Msg("web request")
	})
}

// withRecover keeps a panicking handler from taking the bot down with it. The
// HTTP server shares the bot's process, so this is not defence in depth — it is
// the only thing between a nil map in a handler and the bot going offline.
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Error().
					Str("request_id", requestIDOf(r.Context())).
					Str("path", r.URL.Path).
					Interface("panic", v).
					Bytes("stack", stack()).
					Msg("web handler panicked")
				writeErr(w, r, http.StatusInternalServerError, codeInternal, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withConcurrencyLimit bounds requests in flight. The dashboard and the bot
// share one pgxpool, so an unbounded frontend loop would starve the bot of
// database connections — a web bug taking out the bot is the failure mode this
// exists to prevent. Over the limit answers 503 immediately rather than queuing,
// because a queued request just holds the connection it was denied.
func withConcurrencyLimit(limit int, next http.Handler) http.Handler {
	slots := make(chan struct{}, limit)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			next.ServeHTTP(w, r)
		default:
			log.Warn().
				Str("request_id", requestIDOf(r.Context())).
				Int("limit", limit).
				Msg("web request rejected, concurrency limit reached")
			w.Header().Set("Retry-After", "1")
			writeErr(w, r, http.StatusServiceUnavailable, codeBusy, "too many requests in flight")
		}
	})
}
