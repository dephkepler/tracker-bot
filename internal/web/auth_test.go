package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"tracker-bot/internal/models"
	"tracker-bot/internal/utils/webctx"
)

// A throwaway string shaped like a bot token; never a live one, this repository
// is public.
const testBotToken = "1234567890:AAFakeTokenForTestsOnly_NotARealBotToken"

// signInitData produces init data the way Telegram does. Duplicated from
// tgauth's own tests on purpose — a test helper cannot be imported across
// packages, and reaching into the implementation instead would mean these tests
// pass even if the signing were wrong.
func signInitData(t *testing.T, token string, tgUserID int64, authDate time.Time) string {
	t.Helper()

	fields := map[string]string{
		"user":      fmt.Sprintf(`{"id":%d,"first_name":"Davyd","language_code":"ru"}`, tgUserID),
		"auth_date": strconv.FormatInt(authDate.Unix(), 10),
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+fields[k])
	}

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(pairs, "\n")))

	out := url.Values{}
	for k, v := range fields {
		out.Set(k, v)
	}
	out.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return out.Encode()
}

func authHeader(t *testing.T, tgUserID int64) string {
	t.Helper()
	return "tma " + signInitData(t, testBotToken, tgUserID, time.Now().Add(-time.Minute))
}

// call runs one request through the full server chain and returns the recorder.
func call(t *testing.T, srv *Server, path, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	return rec
}

func decodeErr(t *testing.T, rec *httptest.ResponseRecorder) errorDetail {
	t.Helper()
	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v (body %q)", err, rec.Body.String())
	}
	return body.Error
}

func TestMeRequiresCredentials(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	rec := call(t, srv, "/api/v1/me", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if code := decodeErr(t, rec).Code; code != codeMissingCredentials {
		t.Fatalf("code = %q, want %q", code, codeMissingCredentials)
	}
}

func TestMeRejectsWrongScheme(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	// A bearer token is the wrong shape here: nothing was ever issued to the
	// client, so anything but "tma" is a client that misunderstands the API.
	for _, header := range []string{
		"Bearer " + signInitData(t, testBotToken, knownTgUserID, time.Now()),
		"tma",
		"tma ",
		signInitData(t, testBotToken, knownTgUserID, time.Now()), // no scheme
	} {
		rec := call(t, srv, "/api/v1/me", header)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("header %q: status = %d, want 401", header, rec.Code)
		}
		if code := decodeErr(t, rec).Code; code != codeMissingCredentials {
			t.Fatalf("header %q: code = %q, want %q", header, code, codeMissingCredentials)
		}
	}
}

// RFC 9110 says the auth scheme is case-insensitive, and clients do vary.
func TestSchemeIsCaseInsensitive(t *testing.T) {
	srv := newTestServer(t, testWebConfig())
	raw := signInitData(t, testBotToken, knownTgUserID, time.Now().Add(-time.Minute))

	for _, scheme := range []string{"tma", "TMA", "Tma"} {
		rec := call(t, srv, "/api/v1/me", scheme+" "+raw)
		if rec.Code != http.StatusOK {
			t.Fatalf("scheme %q: status = %d, want 200", scheme, rec.Code)
		}
	}
}

func TestMeRejectsInitDataSignedByAnotherBot(t *testing.T) {
	srv := newTestServer(t, testWebConfig())
	raw := signInitData(t, "9999999:BBAnotherBotEntirely", knownTgUserID, time.Now())

	rec := call(t, srv, "/api/v1/me", "tma "+raw)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if code := decodeErr(t, rec).Code; code != codeUnauthorized {
		t.Fatalf("code = %q, want %q", code, codeUnauthorized)
	}
}

// A stale launch is reported separately from a bad signature: reopening the app
// fixes the former and never fixes the latter, and the client shows a different
// message for each.
func TestMeReportsAnExpiredLaunchDistinctly(t *testing.T) {
	cfg := testWebConfig()
	cfg.InitDataMaxAge = time.Hour
	srv := newTestServer(t, cfg)

	raw := signInitData(t, testBotToken, knownTgUserID, time.Now().Add(-2*time.Hour))
	rec := call(t, srv, "/api/v1/me", "tma "+raw)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if code := decodeErr(t, rec).Code; code != codeInitDataExpired {
		t.Fatalf("code = %q, want %q", code, codeInitDataExpired)
	}
}

// A genuine Telegram user who never pressed /start. Answering 404 rather than
// creating the row is the point: a read-only dashboard must not write users.
func TestMeAnswers404ForAnUnknownUser(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	rec := call(t, srv, "/api/v1/me", authHeader(t, 999999))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if code := decodeErr(t, rec).Code; code != codeUserNotFound {
		t.Fatalf("code = %q, want %q", code, codeUserNotFound)
	}
}

func TestMeReturnsTheResolvedUser(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	rec := call(t, srv, "/api/v1/me", authHeader(t, knownTgUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}

	var body struct {
		TgUserID         int64  `json:"tg_user_id"`
		Timezone         string `json:"timezone"`
		UTCOffsetMinutes int    `json:"utc_offset_minutes"`
		Language         string `json:"language"`
		Now              string `json:"now"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %q)", err, rec.Body.String())
	}

	if body.TgUserID != knownTgUserID {
		t.Fatalf("tg_user_id = %d, want %d", body.TgUserID, knownTgUserID)
	}
	// The fake's user is in New York. Seeing Europe/Warsaw here would mean the
	// user's own zone was dropped somewhere and the app default used instead —
	// which would silently mis-bucket every aggregate that follows.
	if body.Timezone != "America/New_York" {
		t.Fatalf("timezone = %q, want America/New_York", body.Timezone)
	}
	if body.UTCOffsetMinutes >= 0 {
		t.Fatalf("utc_offset_minutes = %d, want a negative offset for New York", body.UTCOffsetMinutes)
	}
	if body.Language != "ru" {
		t.Fatalf("language = %q, want ru", body.Language)
	}
	if body.Now == "" {
		t.Fatal("now is empty")
	}
	// The database id is not the client's business.
	if strings.Contains(rec.Body.String(), "db_user_id") {
		t.Fatal("the response leaks the database id")
	}
}

func TestHandlerSeesTheResolvedUser(t *testing.T) {
	srv := newTestServer(t, testWebConfig())

	var got webctx.User
	h := srv.withAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = webctx.MustFrom(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/whatever", nil)
	req.Header.Set("Authorization", authHeader(t, knownTgUserID))
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got.TgUserID != knownTgUserID {
		t.Fatalf("TgUserID = %d, want %d", got.TgUserID, knownTgUserID)
	}
	// users.id, not the Telegram id — the fake maps 424242 to 7.
	if got.DBUserID != 7 {
		t.Fatalf("DBUserID = %d, want 7", got.DBUserID)
	}
	if got.Location == nil {
		t.Fatal("Location is nil; every aggregate would fall back to the server zone")
	}
	if got.Location.String() != "America/New_York" {
		t.Fatalf("Location = %q", got.Location.String())
	}
}

// The dev bypass is the only way to open the dashboard in an ordinary browser,
// where nothing can sign init data.
func TestDevBypassServesWithoutAnyCredential(t *testing.T) {
	cfg := testWebConfig()
	cfg.DevTgUserID = knownTgUserID
	cfg.PublicOrigin = "http://localhost:3000"
	srv := newTestServer(t, cfg)

	rec := call(t, srv, "/api/v1/me", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
}

// config.Web.Validate is what keeps the bypass out of production; asserted here
// too because this is the file where someone would go looking for it.
func TestDevBypassCannotBeConfiguredWithAPublicOrigin(t *testing.T) {
	cfg := testWebConfig()
	cfg.DevTgUserID = knownTgUserID
	cfg.PublicOrigin = "https://example.com"

	if err := cfg.Validate(); err == nil {
		t.Fatal("a dev bypass with an https origin was accepted")
	}
}

// Identity is consulted once per user, not once per request: it costs two
// queries and the answer changes rarely.
func TestIdentityIsCached(t *testing.T) {
	cfg := testWebConfig()
	entrysvc, profilesvc := newFakes()
	deps := testDeps()
	deps.Entry, deps.Profile = entrysvc, profilesvc
	srv, err := NewServer(t.Context(), cfg, deps)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	for range 3 {
		if rec := call(t, srv, "/api/v1/me", authHeader(t, knownTgUserID)); rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	}

	if profilesvc.calls != 1 || entrysvc.calls != 1 {
		t.Fatalf("lookups: profile=%d entry=%d, want 1 and 1", profilesvc.calls, entrysvc.calls)
	}
}

// A database blip is not a fact about the user, so it must not be cached —
// otherwise one failed query pins an error for the whole negative TTL.
func TestTransientErrorsAreNotCached(t *testing.T) {
	entrysvc, profilesvc := newFakes()
	profilesvc.err = fmt.Errorf("connection refused")
	id := NewIdentity(entrysvc, profilesvc)

	if _, err := id.Resolve(t.Context(), knownTgUserID); err == nil {
		t.Fatal("expected the transient error")
	}
	profilesvc.err = nil

	user, err := id.Resolve(t.Context(), knownTgUserID)
	if err != nil {
		t.Fatalf("second Resolve after recovery: %v", err)
	}
	if user.DBUserID != 7 {
		t.Fatalf("DBUserID = %d, want 7", user.DBUserID)
	}
}

func TestUnknownUserIsCachedBriefly(t *testing.T) {
	entrysvc, profilesvc := newFakes()
	id := NewIdentity(entrysvc, profilesvc)

	for range 3 {
		if _, err := id.Resolve(t.Context(), 555); err == nil {
			t.Fatal("expected ErrUserNotFound")
		}
	}
	// One lookup, not three: a client looping on a 404 must not loop on the
	// database too.
	if profilesvc.calls != 1 {
		t.Fatalf("profile lookups = %d, want 1", profilesvc.calls)
	}
}

func TestIdentityRejectsNonPositiveIDs(t *testing.T) {
	entrysvc, profilesvc := newFakes()
	id := NewIdentity(entrysvc, profilesvc)

	for _, tgUserID := range []int64{0, -1} {
		if _, err := id.Resolve(t.Context(), tgUserID); err != models.ErrUserNotFound {
			t.Fatalf("Resolve(%d) = %v, want ErrUserNotFound", tgUserID, err)
		}
	}
	if profilesvc.calls != 0 {
		t.Fatal("a non-positive id should never reach the services")
	}
}
