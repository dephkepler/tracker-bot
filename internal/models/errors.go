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
