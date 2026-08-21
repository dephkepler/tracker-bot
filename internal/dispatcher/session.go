package dispatcher

import (
	"sync"
	"time"
)

// userSession fields are unsynchronized — safe only because Dispatcher.Run
// processes updates for all users serially. Handling updates concurrently
// would need a mutex on this struct too.
type userSession struct {
	dbID int64

	screen       string
	screenLoaded bool

	tz       string
	tzLoaded bool

	// lang is the raw users.language code — run it through i18n.Normalize
	// before use, it isn't normalized here.
	lang       string
	langLoaded bool

	waitingActivityName       bool
	waitingLocation           bool
	waitingLanguage           bool
	waitingPeriodRange        bool
	waitingCustomTimerMinutes bool
	waitingTrackTargetMinutes bool
	// pendingTrackTargetActivityID is only meaningful while
	// waitingTrackTargetMinutes is true.
	pendingTrackTargetActivityID int64

	waitingLearningCollectionName   bool
	waitingLearningWords            bool
	waitingLearningRenameCollection bool
	// learningCollectionID is only meaningful while waitingLearningWords is
	// true (set on collection create or "Add words").
	learningCollectionID int64

	waitingRoadmapGoalName   bool
	waitingRoadmapGoalRename bool
	waitingRoadmapName       bool
	waitingRoadmapCriteria   bool
	waitingRoadmapRename     bool
	waitingRoadmapCards      bool
	// waitingRoadmapAICards is the same paste flow as waitingRoadmapCards
	// but tagged by the model instead of by inline tags. Two flags rather
	// than one plus a mode, so neither path can silently become the other.
	waitingRoadmapAICards bool
	// roadmapID / roadmapGoalID are whichever technology and goal the user is
	// currently working inside — set when one is created or opened, and read
	// while any waitingRoadmap* flag above is true (the typed text carries no
	// id of its own).
	roadmapID     int64
	roadmapGoalID int64

	waitingAdminBroadcastText bool
	// pendingBroadcastText holds the typed broadcast between the prompt and
	// the admin's Send/Cancel tap on the confirm step.
	pendingBroadcastText string

	waitingChallengeName bool
	// pendingChallengeName holds the typed name between the name prompt and
	// the date-range calendar's confirm step.
	pendingChallengeName string
	// challengeID is the open challenge (grid view) — taps only carry a
	// day, so this is what resolves which challenge they belong to.
	challengeID       int64
	challengeCalMonth time.Time
	challengeCalFrom  time.Time
	challengeCalTo    time.Time

	reportSelected map[int64]bool
	reportFrom     time.Time
	reportTo       time.Time
	reportCalMonth time.Time
	reportCalFrom  time.Time
	reportCalTo    time.Time

	lastSeen time.Time
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[int64]*userSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[int64]*userSession)}
}

// get is not a pure lookup: it creates the session on first touch and
// bumps lastSeen on every call, existing session or not.
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

// sweep evicts sessions idle longer than maxAge; safe to run periodically
// since an evicted user just rehydrates from the database on their next
// message, same as a cold process start.
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
