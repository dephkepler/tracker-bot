package tgctx

import (
	"context"
	"time"
	"tracker-bot/internal/i18n"
)

type MsgContext struct {
	Ctx context.Context

	ChatID   int64
	UserID   int64
	DBUserID int64
	// not kept in sync with the separate users.username column (set in ensureUser).
	Username string
	// always a valid i18n.Lang (i18n.Normalize guarantees it) — safe to pass to i18n.T with no check.
	Language i18n.Lang
	// true only on the call that created this user's row (their first /start), not on later returns.
	IsNewUser bool
	// use this instead of apptime.Now()/apptime.Location directly whenever a specific user is in scope.
	Location *time.Location

	Text      string
	MessageID int
}
