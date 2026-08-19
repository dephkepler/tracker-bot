package tgctx

import (
	"context"
	"time"
	"tracker-bot/internal/i18n"
)

// MsgContext is Telegram update context for message-based handlers.
type MsgContext struct {
	Ctx context.Context

	ChatID   int64
	UserID   int64
	DBUserID int64
	// Username is this user's current Telegram @handle (no leading "@"),
	// as reported on the update that produced this context — empty if they
	// don't have one set. Used for the admin check (see handlers.Module.
	// IsAdmin); not persisted separately from the DB's own users.username
	// column, which is set independently in ensureUser.
	Username string
	// Language is this user's resolved interface language (from
	// users.language, loaded/cached the same way as Location — see
	// internal/dispatcher/dispatcher.go ensureUser). Always a supported
	// i18n.Lang (never empty/invalid — i18n.Normalize guarantees that), so
	// callers can pass it straight to i18n.T without a nil/empty check.
	Language i18n.Lang
	// IsNewUser is true only in the handleMessage call that first created this
	// user's row (i.e. their very first /start) — used to show a one-time
	// welcome instead of the plain "back home" message on every return.
	IsNewUser bool
	// Location is this user's resolved timezone (detected via shared
	// location, see internal/service/profile.go ChangeTimeZone, or
	// apptime.Location as fallback before they've shared one). Handlers
	// should use this instead of apptime.Now()/apptime.Location directly
	// whenever a specific user is in scope.
	Location *time.Location

	Text      string
	MessageID int
}
