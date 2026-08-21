package web

import (
	"context"
	"errors"
	"sync"
	"time"

	"tracker-bot/internal/models"
	"tracker-bot/internal/service"
	"tracker-bot/internal/utils/webctx"
	"tracker-bot/pkg/apptime"
)

// Identity turns a Telegram user id into the request-scoped user: users.id plus
// the timezone every aggregate buckets by.
//
// Cached because it costs two queries and the answer changes rarely, mirroring
// how the dispatcher caches the same three fields on its in-memory session
// (internal/dispatcher/dispatcher.go). Unlike that one, this cache expires:
// a timezone changed in the bot has to reach the dashboard without a restart.
type Identity struct {
	entrysvc   service.EntryService
	profilesvc service.ProfileService

	hitTTL  time.Duration
	missTTL time.Duration
	now     func() time.Time

	mu    sync.RWMutex
	cache map[int64]cachedUser
}

type cachedUser struct {
	user    webctx.User
	err     error
	expires time.Time
}

const (
	// Short enough that changing your timezone in the bot shows up in the
	// dashboard while you are still looking at it.
	identityHitTTL = 5 * time.Minute
	// Negative results are cached too, just briefly: without this a client
	// looping on a 404 would run two queries per request forever.
	identityMissTTL = 30 * time.Second
)

func NewIdentity(entrysvc service.EntryService, profilesvc service.ProfileService) *Identity {
	return &Identity{
		entrysvc:   entrysvc,
		profilesvc: profilesvc,
		hitTTL:     identityHitTTL,
		missTTL:    identityMissTTL,
		now:        time.Now,
		cache:      make(map[int64]cachedUser),
	}
}

// Resolve returns the user behind a Telegram id.
//
// models.ErrUserNotFound means the person has a genuine Telegram credential but
// no row here — they opened the Mini App link without ever pressing /start.
// That is deliberately not a reason to create one: a read-only dashboard has no
// business writing users.
func (i *Identity) Resolve(ctx context.Context, tgUserID int64) (webctx.User, error) {
	if tgUserID <= 0 {
		return webctx.User{}, models.ErrUserNotFound
	}

	if entry, ok := i.lookup(tgUserID); ok {
		return entry.user, entry.err
	}

	user, err := i.load(ctx, tgUserID)
	// A cancelled request or a dead database is not a fact about this user, so
	// it must not be cached — otherwise one blip pins an error for 30 seconds.
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return webctx.User{}, err
	}
	i.store(tgUserID, user, err)
	return user, err
}

func (i *Identity) lookup(tgUserID int64) (cachedUser, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	entry, ok := i.cache[tgUserID]
	if !ok || i.now().After(entry.expires) {
		return cachedUser{}, false
	}
	return entry, true
}

func (i *Identity) store(tgUserID int64, user webctx.User, err error) {
	ttl := i.hitTTL
	if err != nil {
		ttl = i.missTTL
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cache[tgUserID] = cachedUser{user: user, err: err, expires: i.now().Add(ttl)}
}

func (i *Identity) load(ctx context.Context, tgUserID int64) (webctx.User, error) {
	// GetProfileStats keys on the Telegram id; GetDBIDByTgUserID is what
	// produces the users.id everything else wants.
	stats, err := i.profilesvc.GetProfileStats(ctx, tgUserID)
	if err != nil {
		return webctx.User{}, err
	}
	dbID, err := i.entrysvc.GetDBIDByTgUserID(ctx, tgUserID)
	if err != nil {
		return webctx.User{}, err
	}

	tzName := ""
	if stats.TimeZone != nil {
		tzName = *stats.TimeZone
	}
	language := ""
	if stats.Language != nil {
		language = *stats.Language
	}

	// apptime.Resolve never errors and never returns nil — an unset or
	// unparseable zone falls back to the app default, same as the bot.
	loc := apptime.Resolve(tzName)
	return webctx.User{
		TgUserID: tgUserID,
		DBUserID: dbID,
		Location: loc,
		TZName:   loc.String(),
		Language: language,
	}, nil
}
