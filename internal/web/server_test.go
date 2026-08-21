package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"tracker-bot/internal/config"
)

func testWebConfig() config.Web {
	return config.Web{
		Enabled:        true,
		Addr:           "127.0.0.1:0",
		InitDataMaxAge: 24 * time.Hour,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
		IdleTimeout:    time.Second,
		ShutdownGrace:  time.Second,
		MaxInflight:    8,
	}
}

func newTestServer(t *testing.T, cfg config.Web) *Server {
	t.Helper()
	entrysvc, profilesvc := newFakes()
	srv, err := NewServer(context.Background(), cfg, testBotToken, entrysvc, profilesvc, newFakeTrackSvc())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return srv
}

func TestNewServerRejectsInvalidConfig(t *testing.T) {
	cfg := testWebConfig()
	cfg.MaxInflight = 0
	entrysvc, profilesvc := newFakes()

	if _, err := NewServer(context.Background(), cfg, testBotToken, entrysvc, profilesvc, newFakeTrackSvc()); err == nil {
		t.Fatal("NewServer accepted an invalid config")
	}
}

func TestNewServerRejectsNilContext(t *testing.T) {
	entrysvc, profilesvc := newFakes()
	//nolint:staticcheck // passing nil is the thing under test
	if _, err := NewServer(nil, testWebConfig(), testBotToken, entrysvc, profilesvc, newFakeTrackSvc()); err == nil {
		t.Fatal("NewServer accepted a nil context")
	}
}

// Without a bot token there is nothing to verify launches against, so starting
// would mean serving an endpoint that can only ever 401.
func TestNewServerRequiresABotTokenUnlessBypassed(t *testing.T) {
	entrysvc, profilesvc := newFakes()
	if _, err := NewServer(context.Background(), testWebConfig(), "", entrysvc, profilesvc, newFakeTrackSvc()); err == nil {
		t.Fatal("NewServer accepted an empty bot token")
	}

	// Except in dev-bypass mode, where nothing is verified at all.
	cfg := testWebConfig()
	cfg.DevTgUserID = knownTgUserID
	if _, err := NewServer(context.Background(), cfg, "", entrysvc, profilesvc, newFakeTrackSvc()); err != nil {
		t.Fatalf("dev bypass should not need a token: %v", err)
	}
}

func TestNewServerRequiresServices(t *testing.T) {
	if _, err := NewServer(context.Background(), testWebConfig(), testBotToken, nil, nil, nil); err == nil {
		t.Fatal("NewServer accepted nil services")
	}
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %q)", err, rec.Body.String())
	}
	if body["status"] != "ok" {
		t.Fatalf("body = %v", body)
	}
	// Every response is one user's own data behind a header credential, so
	// nothing may be cached by anything in between.
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("no X-Request-Id header")
	}
}

// An unrouted path must answer JSON, not net/http's text 404 — otherwise the
// frontend's error path has two shapes to parse.
func TestUnknownPathAnswersJSON(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %q)", err, rec.Body.String())
	}
	if body.Error.Code != codeNotFound {
		t.Fatalf("code = %q, want %q", body.Error.Code, codeNotFound)
	}
	if body.Error.RequestID == "" {
		t.Fatal("error body carries no request id")
	}
}

func TestRequestIDsAreUnique(t *testing.T) {
	srv := newTestServer(t, testWebConfig())
	h := srv.routes()

	seen := make(map[string]bool, 50)
	for range 50 {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		id := rec.Header().Get("X-Request-Id")
		if seen[id] {
			t.Fatalf("request id %q was reused", id)
		}
		seen[id] = true
	}
}

// A panicking handler must not reach the runtime: this server shares the bot's
// process, so an escaping panic would take the bot offline.
func TestPanicIsContainedAndAnswers500(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /boom", func(http.ResponseWriter, *http.Request) {
		panic("exploded")
	})
	var h http.Handler = mux
	h = withConcurrencyLimit(srv.cfg.MaxInflight, h)
	h = withRecover(h)
	h = withLogging(h)
	h = withRequestID(h)

	rec := httptest.NewRecorder()
	// The test itself fails rather than the process dying if this escapes.
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %q)", err, rec.Body.String())
	}
	if body.Error.Code != codeInternal {
		t.Fatalf("code = %q, want %q", body.Error.Code, codeInternal)
	}
	// The panic value must not leak to the client.
	if body.Error.Message == "exploded" {
		t.Fatal("the panic value reached the response body")
	}
}

func TestConcurrencyLimitRejectsOverflow(t *testing.T) {
	const limit = 2

	release := make(chan struct{})
	entered := make(chan struct{}, limit)

	blocked := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	})
	h := withRequestID(withConcurrencyLimit(limit, blocked))

	var wg sync.WaitGroup
	codes := make([]int, limit)
	for i := range limit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))
			codes[i] = rec.Code
		}()
	}

	// Both slots are taken and parked before the next request arrives.
	for range limit {
		<-entered
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("over-limit status = %d, want 503", rec.Code)
	}
	// Rejected immediately rather than queued — a queued request still holds
	// the connection it was denied.
	if got := rec.Header().Get("Retry-After"); got == "" {
		t.Fatal("503 carries no Retry-After")
	}

	close(release)
	wg.Wait()
	for i, code := range codes {
		if code != http.StatusOK {
			t.Fatalf("in-limit request %d got %d, want 200", i, code)
		}
	}

	// A freed slot must be reusable, or the limiter leaks capacity.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))
	if rec.Code == http.StatusServiceUnavailable {
		t.Fatal("slots were not released")
	}
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := testWebConfig()
	entrysvc, profilesvc := newFakes()

	srv, err := NewServer(ctx, cfg, testBotToken, entrysvc, profilesvc, newFakeTrackSvc())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	done := make(chan struct{})
	go func() {
		srv.Run()
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}
}
