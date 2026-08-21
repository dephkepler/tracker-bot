// Package webctx carries the authenticated dashboard user through a request,
// the same role tgctx plays for a Telegram update.
package webctx

import (
	"context"
	"time"
)

// User is who the request is for, resolved once by the auth middleware.
type User struct {
	// TgUserID is the Telegram id, taken from the verified init data.
	TgUserID int64
	// DBUserID is users.id. Every track/learning/roadmap/challenge service
	// takes this one; only profilesvc takes TgUserID. Mixing them up returns
	// another user's data or none at all, which is why both are named for
	// which id they are.
	DBUserID int64
	// Location is the user's own timezone, already resolved through
	// apptime.Resolve, so it is never nil. Every aggregate has to bucket by
	// this user's day boundaries rather than the server's.
	Location *time.Location
	TZName   string
	Language string
}

type ctxKey struct{}

func With(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

func From(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxKey{}).(User)
	return u, ok
}

// MustFrom is for handlers, which only ever run behind the auth middleware. It
// panics rather than returning a zero user, because a zero DBUserID would
// quietly query user 0 instead of failing; the panic is caught by the recover
// middleware and answered as a 500.
func MustFrom(ctx context.Context) User {
	u, ok := From(ctx)
	if !ok {
		panic("webctx: no user in context — handler is not behind the auth middleware")
	}
	return u
}
