package dispatcher

import (
	"sync"
	"time"
)

// userSession holds every piece of per-user, in-memory state the dispatcher
// juggles between messages: what screen they're on, what (if anything) it's
// waiting for them to type/share next, their resolved timezone, and the
// scratch state for building a report. This replaces what used to be ten
// separate parallel maps on Dispatcher, keyed by the same Telegram user id.
//
// Concurrency: sessionStore.get is safe to call from multiple goroutines,
// but the *userSession it returns is not independently locked — field reads
// and writes on it are only safe because Dispatcher.Run processes updates
// one at a time. If updates are ever handled concurrently, this struct needs
// its own mutex too.
type userSession struct {
	dbID int64

	screen       string
	screenLoaded bool // true once loaded from user_ui_state (or set explicitly)

	tz       string // IANA zone name, "" until the user shares a location
	tzLoaded bool   // true once loaded from the users table (or set explicitly)

	lang       string // raw users.language code, "" until loaded/set — see i18n.Normalize
	langLoaded bool   // true once loaded from the users table (or invalidated after a change)

	waitingActivityName       bool
	waitingLocation           bool
	waitingLanguage           bool
	waitingPeriodRange        bool
	waitingCustomTimerMinutes bool

	waitingLearningCollectionName   bool
	waitingLearningWords            bool
	waitingLearningRenameCollection bool
	// learningCollectionID is the collection currently being edited — set
	// when creating a collection or opening "Add words" on an existing one,
	// used while waitingLearningWords is true.
	learningCollectionID int64

	waitingAdminBroadcastText bool
	// pendingBroadcastText holds the typed broadcast message between the
	// confirm-step prompt and the admin tapping Send/Cancel on it.
	pendingBroadcastText string

	reportSelected map[int64]bool
	reportFrom     time.Time
	reportTo       time.Time
	reportCalMonth time.Time
	reportCalFrom  time.Time
	reportCalTo    time.Time

	lastSeen time.Time
}

// sessionStore is a concurrency-safe registry of userSessions, keyed by
// Telegram user id.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[int64]*userSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[int64]*userSession)}
}

// get returns the session for userID, creating an empty one on first touch.
func (s *sessionStore) get(userID int64) *userSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[userID]
	if !ok {
		sess = &userSession{}
		s.sessions[userID] = sess
	}
	sess.lastSeen = time.Now()
	return sess
}

// sweep evicts sessions untouched for longer than maxAge, so a long-running
// process doesn't grow this map forever. Safe to run periodically in the
// background (see Dispatcher.sweepLoop) — an evicted user simply gets a
// fresh session next time they message the bot, rehydrated from the
// database (screen, timezone) exactly like on a cold process start.
func (s *sessionStore) sweep(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if sess.lastSeen.Before(cutoff) {
			delete(s.sessions, id)
		}
	}
}
