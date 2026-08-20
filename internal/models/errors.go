package models

import "errors"

var (
	// Activity domain errors.
	ErrActivityExists   = errors.New("activity already exists")
	ErrActivityNotFound = errors.New("activity not found")
	ErrForbidden        = errors.New("forbidden")

	// User domain errors.
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	// Custom timer domain errors.
	ErrCustomTimerInvalidInterval = errors.New("custom timer interval must be between 1 and 360 minutes")
	ErrCustomTimerLimitReached    = errors.New("custom timer limit reached")
	ErrCustomTimerNotFound        = errors.New("custom timer not found")

	// Learning domain errors.
	ErrLearningCollectionExists   = errors.New("collection already exists")
	ErrLearningCollectionNotFound = errors.New("collection not found")
	ErrLearningNoWordsParsed      = errors.New("no valid \"word - translation\" lines found")
	ErrLearningWordNotFound       = errors.New("word not found")
	ErrLearningInvalidInterval    = errors.New("learning push interval must be between 1 and 1440 minutes")
	ErrLearningInvalidName        = errors.New("collection name must be 2-60 characters, single line")

	// Challenge domain errors.
	ErrChallengeExists       = errors.New("challenge already exists")
	ErrChallengeNotFound     = errors.New("challenge not found")
	ErrChallengeInvalidName  = errors.New("challenge name must be 2-60 characters, single line")
	ErrChallengeInvalidRange = errors.New("challenge must be 1-100 days, end date on or after start date")
	ErrChallengeDayNotFound  = errors.New("challenge day not found")

	// Roadmap domain errors.
	ErrRoadmapExists          = errors.New("roadmap already exists")
	ErrRoadmapNotFound        = errors.New("roadmap not found")
	ErrRoadmapCardNotFound    = errors.New("roadmap card not found")
	ErrRoadmapNoCardsParsed   = errors.New("no non-empty card lines found")
	ErrRoadmapInvalidInterval = errors.New("roadmap push interval must be between 1 and 1440 minutes")
	ErrRoadmapInvalidName     = errors.New("roadmap name must be 2-60 characters, single line")
	ErrRoadmapGoalTooLong     = errors.New("roadmap goal must be at most 200 characters")
	ErrRoadmapLimitReached    = errors.New("roadmap limit reached")
)

// MaxRoadmapsPerUser caps how many non-archived roadmaps one user can keep
// at once — the feature is deliberately "a few technologies at a time", and
// the cap is what keeps the push digest and menus readable.
const MaxRoadmapsPerUser = 5

// Roadmap text limits, enforced in the service layer (the DB columns are
// plain TEXT — same as learning_collections.name).
const (
	MaxRoadmapGoalLen     = 200
	MaxRoadmapCardTextLen = 300
)

// Digest caps: at most RoadmapDigestPerRoadmapCap pending cards from any one
// roadmap, and RoadmapDigestMaxCards overall, so no single roadmap
// monopolizes a push and the message stays skimmable.
const (
	RoadmapDigestPerRoadmapCap = 3
	RoadmapDigestMaxCards      = 8
)

// MaxCustomTimersPerUser caps how many custom intervals one user can keep at
// once, so the timer picker doesn't grow without bound.
const MaxCustomTimersPerUser = 5

// MinCustomTimerMinutes and MaxCustomTimerMinutes bound a valid custom
// interval; kept in sync with the chk_custom_interval_min_range DB
// constraint (migrations/0008_user_custom_timers.up.sql).
const (
	MinCustomTimerMinutes = 1
	MaxCustomTimerMinutes = 360
)
