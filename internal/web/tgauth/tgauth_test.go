package tgauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A throwaway string in the shape of a real bot token. Never a live one: this
// file is committed to a public repository.
const testToken = "1234567890:AAFakeTokenForTestsOnly_NotARealBotToken"

var testNow = time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)

func newTestVerifier(t *testing.T, maxAge time.Duration) *Verifier {
	t.Helper()
	v, err := NewVerifier(testToken, maxAge)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}
	v.now = func() time.Time { return testNow }
	return v
}

// signFields builds init data the way Telegram does.
//
// Deliberately written from the specification rather than by calling
// expectedHash: sharing that function would make this helper agree with the
// implementation even if both were wrong, which is the one failure these tests
// exist to catch. It is still the same algorithm expressed twice by the same
// author, so it is not independent evidence — see TestGoldenVector.
func signFields(t *testing.T, token string, fields map[string]string) string {
	t.Helper()

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+fields[k])
	}

	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	secretMac.Write([]byte(token))

	mac := hmac.New(sha256.New, secretMac.Sum(nil))
	mac.Write([]byte(strings.Join(pairs, "\n")))

	out := url.Values{}
	for k, val := range fields {
		out.Set(k, val)
	}
	out.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return out.Encode()
}

func validFields() map[string]string {
	return map[string]string{
		"user":          `{"id":424242,"first_name":"Davyd","username":"dephkepler","language_code":"ru"}`,
		"auth_date":     strconv.FormatInt(testNow.Add(-time.Minute).Unix(), 10),
		"chat_type":     "sender",
		"chat_instance": "-1234567890123456789",
	}
}

func TestVerifyAcceptsAGenuineLaunch(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	raw := signFields(t, testToken, validFields())

	data, err := v.Verify(raw)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if data.User.ID != 424242 {
		t.Fatalf("user id = %d, want 424242", data.User.ID)
	}
	if data.User.Username != "dephkepler" {
		t.Fatalf("username = %q", data.User.Username)
	}
	if data.User.LanguageCode != "ru" {
		t.Fatalf("language = %q", data.User.LanguageCode)
	}
}

// The check string is built from decoded values, so a name containing a space
// or an emoji must verify. Signing the percent-encoded form instead passes
// every ASCII test and then 401s exactly those users in production.
func TestVerifyAcceptsSpacesAndEmojiInUserFields(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	fields["user"] = `{"id":7,"first_name":"Анна Мария 🌿","last_name":"Ко ва ль"}`

	if _, err := v.Verify(signFields(t, testToken, fields)); err != nil {
		t.Fatalf("Verify with spaces and emoji: %v", err)
	}
}

// Current clients send an Ed25519 "signature" field for third-party
// validation. It is still one of the "received fields", so it belongs in the
// data-check-string for the bot-token HMAC — only "hash" is left out. Omitting
// it rejected every real launch in production while this suite stayed green.
func TestVerifyIncludesTheSignatureFieldInTheCheckString(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	fields["signature"] = "GkQr1AvOBmXqDdaHnUYh5Ck7Q0nGZ0Kz2fWnGkTFxSg"

	if _, err := v.Verify(signFields(t, testToken, fields)); err != nil {
		t.Fatalf("Verify with a signed signature field: %v", err)
	}
}

// The corollary: because the field is part of the string, adding or altering it
// after signing invalidates the hash, exactly like any other tampering.
func TestVerifyRejectsSignatureAddedAfterSigning(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	values, err := url.ParseQuery(signFields(t, testToken, validFields()))
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	values.Set("signature", "appended-after-the-fact")

	if _, err := v.Verify(values.Encode()); !errors.Is(err, ErrBadHash) {
		t.Fatalf("err = %v, want ErrBadHash", err)
	}
}

func TestVerifyRejectsTamperedUser(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	raw := signFields(t, testToken, validFields())

	values, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	// Swapping in someone else's id after signing is the attack that matters:
	// it is how one user would read another's data.
	values.Set("user", `{"id":999999,"first_name":"Mallory"}`)

	if _, err := v.Verify(values.Encode()); !errors.Is(err, ErrBadHash) {
		t.Fatalf("err = %v, want ErrBadHash", err)
	}
}

func TestVerifyRejectsFlippedHashCharacter(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	values, err := url.ParseQuery(signFields(t, testToken, validFields()))
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}

	h := []byte(values.Get("hash"))
	if h[0] == 'a' {
		h[0] = 'b'
	} else {
		h[0] = 'a'
	}
	values.Set("hash", string(h))

	if _, err := v.Verify(values.Encode()); !errors.Is(err, ErrBadHash) {
		t.Fatalf("err = %v, want ErrBadHash", err)
	}
}

func TestVerifyRejectsAnotherBotsToken(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	// Correctly signed — just not by us.
	raw := signFields(t, "9999999:BBSomeOtherBotsTokenEntirelyDifferent", validFields())

	if _, err := v.Verify(raw); !errors.Is(err, ErrBadHash) {
		t.Fatalf("err = %v, want ErrBadHash", err)
	}
}

func TestVerifyRejectsMissingHash(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	values := url.Values{}
	for k, val := range validFields() {
		values.Set(k, val)
	}

	if _, err := v.Verify(values.Encode()); !errors.Is(err, ErrMissingHash) {
		t.Fatalf("err = %v, want ErrMissingHash", err)
	}
}

func TestVerifyRejectsStaleLaunch(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	fields["auth_date"] = strconv.FormatInt(testNow.Add(-25*time.Hour).Unix(), 10)

	if _, err := v.Verify(signFields(t, testToken, fields)); !errors.Is(err, ErrStale) {
		t.Fatalf("err = %v, want ErrStale", err)
	}
}

func TestVerifyAcceptsLaunchJustInsideTheWindow(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	fields["auth_date"] = strconv.FormatInt(testNow.Add(-23*time.Hour-59*time.Minute).Unix(), 10)

	if _, err := v.Verify(signFields(t, testToken, fields)); err != nil {
		t.Fatalf("Verify just inside the window: %v", err)
	}
}

func TestVerifyRejectsFutureAuthDate(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	fields["auth_date"] = strconv.FormatInt(testNow.Add(10*time.Minute).Unix(), 10)

	if _, err := v.Verify(signFields(t, testToken, fields)); !errors.Is(err, ErrFuture) {
		t.Fatalf("err = %v, want ErrFuture", err)
	}
}

// A few seconds of clock disagreement between Telegram and this box is normal
// and must not lock anybody out.
func TestVerifyToleratesSmallClockSkew(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	fields["auth_date"] = strconv.FormatInt(testNow.Add(20*time.Second).Unix(), 10)

	if _, err := v.Verify(signFields(t, testToken, fields)); err != nil {
		t.Fatalf("Verify with 20s of skew: %v", err)
	}
}

func TestVerifyRejectsMissingUser(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)
	fields := validFields()
	delete(fields, "user")

	if _, err := v.Verify(signFields(t, testToken, fields)); !errors.Is(err, ErrNoUser) {
		t.Fatalf("err = %v, want ErrNoUser", err)
	}
}

func TestVerifyRejectsMalformedInput(t *testing.T) {
	v := newTestVerifier(t, 24*time.Hour)

	cases := map[string]string{
		"empty":            "",
		"blank":            "   ",
		"not a query":      "%%%",
		"user is not json": "",
	}

	// The user-is-not-JSON case has to be signed to get past the hash check.
	fields := validFields()
	fields["user"] = "not json at all"
	cases["user is not json"] = signFields(t, testToken, fields)

	// A user object without an id is signed correctly but unusable.
	fields = validFields()
	fields["user"] = `{"first_name":"NoId"}`
	cases["user without id"] = signFields(t, testToken, fields)

	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := v.Verify(raw); err == nil {
				t.Fatal("Verify accepted malformed input")
			}
		})
	}
}

func TestNewVerifierRejectsBadArguments(t *testing.T) {
	if _, err := NewVerifier("", time.Hour); err == nil {
		t.Fatal("empty token was accepted")
	}
	if _, err := NewVerifier(testToken, 0); err == nil {
		t.Fatal("zero maxAge was accepted")
	}
}

// TestGoldenVector pins this package against a real Telegram client, and is not
// filled in yet.
//
// What is already established: the HMAC construction itself — the "WebAppData"
// key derivation, the sorted newline-joined check string over decoded values,
// and the hex comparison — has been cross-checked against an independent
// implementation. A signature produced by a separate Python script verified
// here on the first attempt, which is evidence signFields alone cannot give,
// since it and expectedHash are the same reading of the spec written twice.
//
// What only a real client could settle — and did, by rejecting every launch
// with "hash mismatch" — was the payload shape: this package used to exclude
// the "signature" field from the check string, which no test could catch
// because signFields made the same mistake. Filling the vector in below is what
// would catch the next mistake of that kind.
//
// To fill it in — this is step 9 of the plan, the first time the Mini App opens
// against a live client:
//
//  1. create a scratch bot with BotFather and register a Mini App on it;
//  2. open it once and log Telegram.WebApp.initData;
//  3. paste the string and that bot's token into the constants below;
//  4. revoke the scratch token with BotFather, so what is committed is dead;
//  5. delete the t.Skip.
//
// Until then, treat the HMAC as unverified against reality: every test above
// only proves this package is self-consistent.
func TestGoldenVector(t *testing.T) {
	const (
		goldenToken    = ""
		goldenInitData = ""
	)

	if goldenToken == "" || goldenInitData == "" {
		t.Skip("no golden vector captured yet — see the comment above, step 9 of the plan")
	}

	v, err := NewVerifier(goldenToken, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}
	// The capture is older than any real window, so freshness is pinned to it.
	values, err := url.ParseQuery(goldenInitData)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	secs, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		t.Fatalf("auth_date: %v", err)
	}
	v.now = func() time.Time { return time.Unix(secs, 0).Add(time.Minute) }

	if _, err := v.Verify(goldenInitData); err != nil {
		t.Fatalf("golden vector failed to verify: %v — the HMAC construction is wrong", err)
	}
}
