package llm

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// Retry policy for the hand-rolled transport. The Claude client does not use
// this — anthropic-sdk-go retries 408/409/429/5xx itself.
const (
	maxAttempts = 5
	// maxRetryAfter caps how long a provider's Retry-After can stall us. A
	// rate-limited provider answering "come back in 10 minutes" must not
	// hold a Telegram handler open that long.
	maxRetryAfter = 30 * time.Second
)

var backoffs = [maxAttempts]time.Duration{
	1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 8 * time.Second,
}

// httpStatusError is implemented by transport errors that carry a status.
type httpStatusError interface {
	error
	HTTPStatus() int
}

type retryAfterError interface {
	RetryAfter() (time.Duration, bool)
}

// parseRetryAfter reads a Retry-After header in either accepted form:
// delay-seconds or an HTTP-date.
func parseRetryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.ParseFloat(v, 64); err == nil {
		if secs > 0 {
			return time.Duration(secs * float64(time.Second))
		}
		return 0
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// shouldRetry retries only what a retry can fix: transport failures (status
// 0), rate limits, and 5xx. A 400 or 401 is a bug or a bad key and repeating
// it just burns the user's wait.
func shouldRetry(err error) bool {
	var he httpStatusError
	if err == nil || !errors.As(err, &he) {
		return false
	}
	status := he.HTTPStatus()
	if status == 0 || status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500 && status < 600
}

func statusOf(err error) int {
	var he httpStatusError
	if errors.As(err, &he) {
		return he.HTTPStatus()
	}
	return 0
}

func retryAfterOf(err error) (time.Duration, bool) {
	var ra retryAfterError
	if errors.As(err, &ra) {
		return ra.RetryAfter()
	}
	return 0, false
}

// withJitter spreads retries so several users hitting a rate limit at once
// do not march back in lockstep.
func withJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	return d/2 + time.Duration(rand.Int64N(int64(d)))
}

func doWithRetry(ctx context.Context, provider string, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			delay := withJitter(backoffs[attempt-2])
			if ra, ok := retryAfterOf(lastErr); ok && ra > delay {
				delay = min(ra, maxRetryAfter)
			}
			log.Warn().
				Str("provider", provider).
				Int("attempt", attempt).
				Dur("delay", delay).
				Int("status", statusOf(lastErr)).
				Err(lastErr).
				Msg("llm call failed, backing off")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if !shouldRetry(lastErr) {
			return lastErr
		}
	}
	return lastErr
}
