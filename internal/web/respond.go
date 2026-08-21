package web

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/rs/zerolog/log"
)

// Error codes the client branches on. Stable strings — the frontend keys off
// these, not off the human-readable message.
const (
	codeInternal = "internal"
	codeBusy     = "busy"
	codeNotFound = "not_found"

	// codeMissingCredentials means no Authorization header at all — the page
	// was loaded outside Telegram, or before the WebApp script was ready.
	codeMissingCredentials = "missing_credentials"
	// codeUnauthorized means the init data did not verify. Deliberately says
	// nothing about which check failed; the request id ties it to the log line
	// that does.
	codeUnauthorized = "unauthorized"
	// codeInitDataExpired is split out from codeUnauthorized because the client
	// can act on it: reopening the Mini App mints fresh init data, whereas a bad
	// signature is not something a retry fixes.
	codeInitDataExpired = "initdata_expired"
	// codeUserNotFound means the Telegram credential is genuine but nobody has
	// pressed /start — the dashboard tells them to open the bot first.
	codeUserNotFound = "user_not_found"

	codeInvalidParameter = "invalid_parameter"
	codeRangeTooLarge    = "range_too_large"
	// codeUnsupportedCombination is a request that is individually valid in
	// every parameter but asks for something the aggregation cannot do yet.
	codeUnsupportedCombination = "unsupported_combination"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Nothing here is cacheable by a shared cache: every response is one user's
	// own data, keyed by a credential in a header.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status line is already out, so this can only be logged.
		log.Error().
			Str("request_id", requestIDOf(r.Context())).
			Err(err).
			Msg("web response encode failed")
	}
}

// writeErr keeps the failing check out of the response body — the message is
// for a person, the request id is what ties it to the log line that has the
// detail.
func writeErr(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, r, status, errorBody{Error: errorDetail{
		Code:      code,
		Message:   message,
		RequestID: requestIDOf(r.Context()),
	}})
}

func stack() []byte {
	buf := make([]byte, 8<<10)
	return buf[:runtime.Stack(buf, false)]
}
