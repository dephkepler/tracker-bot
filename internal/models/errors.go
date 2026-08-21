package models

import "errors"

var (
	ErrActivityExists        = errors.New("activity already exists")
	ErrActivityNotFound      = errors.New("activity not found")
	ErrActivityTargetInvalid = errors.New("activity target must be between 1 and 1440 minutes")
	ErrForbidden             = errors.New("forbidden")

	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	ErrCustomTimerInvalidInterval = errors.New("custom timer interval must be between 1 and 360 minutes")
	ErrCustomTimerLimitReached    = errors.New("custom timer limit reached")
	ErrCustomTimerNotFound        = errors.New("custom timer not found")

	ErrLearningCollectionExists   = errors.New("collection already exists")
	ErrLearningCollectionNotFound = errors.New("collection not found")
	ErrLearningNoWordsParsed      = errors.New("no valid \"word - translation\" lines found")
	ErrLearningWordNotFound       = errors.New("word not found")
	ErrLearningInvalidInterval    = errors.New("learning push interval must be between 1 and 1440 minutes")
	ErrLearningInvalidName        = errors.New("collection name must be 2-60 characters, single line")

	ErrChallengeExists       = errors.New("challenge already exists")
	ErrChallengeNotFound     = errors.New("challenge not found")
	ErrChallengeInvalidName  = errors.New("challenge name must be 2-60 characters, single line")
	ErrChallengeInvalidRange = errors.New("challenge must be 1-100 days, end date on or after start date")
	ErrChallengeDayNotFound  = errors.New("challenge day not found")

	ErrRoadmapExists        = errors.New("roadmap already exists")
	ErrRoadmapNotFound      = errors.New("roadmap not found")
	ErrRoadmapCardNotFound  = errors.New("roadmap card not found")
	ErrRoadmapNoCardsParsed = errors.New("no non-empty card lines found")
	// ErrRoadmapAIDisabled means no LLM provider is configured. Callers
	// show the manual path instead — the feature is off, not broken.
	ErrRoadmapAIDisabled = errors.New("roadmap ai is not configured")
	// ErrRoadmapAIEmptyResult is a well-formed reply that contained nothing
	// usable, e.g. a generated plan with zero cards.
	ErrRoadmapAIEmptyResult    = errors.New("roadmap ai returned nothing usable")
	ErrRoadmapInvalidInterval  = errors.New("roadmap push interval must be between 1 and 1440 minutes")
	ErrRoadmapInvalidName      = errors.New("roadmap name must be 2-60 characters, single line")
	ErrRoadmapCriteriaTooLong  = errors.New("mastery criteria must be at most 200 characters")
	ErrRoadmapLimitReached     = errors.New("technology limit reached for this goal")
	ErrRoadmapGoalExists       = errors.New("goal already exists")
	ErrRoadmapGoalNotFound     = errors.New("goal not found")
	ErrRoadmapGoalLimitReached = errors.New("goal limit reached")
)

// Caps are per-goal, not global: five technologies is a sane plan for one
// outcome, but a global five would leave a second goal with none.
const (
	MaxRoadmapGoalsPerUser = 3
	MaxRoadmapsPerGoal     = 5
)

// Roadmap text limits, enforced only in the service layer — the DB columns
// are plain unconstrained TEXT.
const (
	MaxRoadmapCriteriaLen = 200
	MaxRoadmapCardTextLen = 300
)

// Digest caps: at most RoadmapDigestPerRoadmapCap cards from any one
// roadmap, RoadmapDigestMaxCards overall, so no roadmap crowds out the rest.
const (
	RoadmapDigestPerRoadmapCap = 3
	RoadmapDigestMaxCards      = 8
)

const MaxCustomTimersPerUser = 5

// MinCustomTimerMinutes and MaxCustomTimerMinutes: keep in sync with
// chk_custom_interval_min_range in migrations/0008_user_custom_timers.up.sql.
const (
	MinCustomTimerMinutes = 1
	MaxCustomTimerMinutes = 360
)

// MinActivityTargetMinutes and MaxActivityTargetMinutes: keep in sync with
// chk_activity_target_minutes_range in migrations/0013_activity_target.up.sql.
const (
	MinActivityTargetMinutes = 1
	MaxActivityTargetMinutes = 1440
)
