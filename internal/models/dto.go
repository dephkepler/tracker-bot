package models

import "time"

// UserInput is used by entry flow to create or update user profile fields.
type UserInput struct {
	TgUserID    int64
	UserName    *string
	PhoneNumber *string
	Email       *string
	Language    *string
	TimeZone    *string
}

// ProfileStats is a lightweight profile view model.
type ProfileStats struct {
	TgUserID    int64
	UserName    *string
	PhoneNumber *string
	Email       *string
	Language    *string
	TimeZone    *string
}

// MainStats contains summary values for tracking home screen.
type MainStats struct {
	CurrentActivityName string
	TodayTracked        time.Duration
	TodaySessions       int
	StreakDays          int
}

// TrackActivityItem is an activity row used in selection UIs.
type TrackActivityItem struct {
	ID       int64
	Name     string
	Emoji    string
	Selected bool
}

// CustomTimerOption is one user-defined timer interval shown next to the
// built-in 15/30 min choices in the timer picker.
type CustomTimerOption struct {
	IntervalMin int
}

// AdminUserRow is one row in the admin "who's using the bot" listing.
type AdminUserRow struct {
	DBID      int64
	TgUserID  int64
	UserName  *string
	CreatedAt time.Time
}

// AdminOverviewStats aggregates bot-wide usage numbers for the admin
// "📊 Overview" screen.
type AdminOverviewStats struct {
	TotalUsers         int
	ActiveTrackTimers  int
	ActiveReviewPushes int
	TotalActivities    int
	TotalCollections   int
	TotalLearningWords int
}

// AdminUserDetail is the admin's per-user drill-down view — profile fields
// plus cross-domain usage counts.
type AdminUserDetail struct {
	DBID             int64
	TgUserID         int64
	UserName         *string
	Language         *string
	TimeZone         *string
	CreatedAt        time.Time
	ActivitiesCount  int
	CollectionsCount int
	LearningWords    int
	TrackTimerActive bool
	ReviewsActive    bool
}

// TimerDueUser represents one user that should receive timer prompt now.
type TimerDueUser struct {
	DBUserID    int64
	TgUserID    int64
	IntervalMin int
}

// ActivityDurationStat is one activity aggregate line in reports.
type ActivityDurationStat struct {
	ActivityID int64
	Name       string
	Emoji      string
	Duration   time.Duration
	Sessions   int
}

// ReportTodayStats is a daily aggregate report.
type ReportTodayStats struct {
	TotalTracked  time.Duration
	TotalSessions int
	TopActivities []ActivityDurationStat
}

// ReportPeriodStats is an aggregate report for arbitrary date range.
type ReportPeriodStats struct {
	From          time.Time
	To            time.Time
	TotalTracked  time.Duration
	TotalSessions int
	Activities    []ActivityDurationStat
	Monthly       []MonthDurationStat
}

// MonthDurationStat stores total duration for one month bucket.
type MonthDurationStat struct {
	Month    time.Time
	Duration time.Duration
}

// HourActivityDuration is one (hour bucket, activity) total — lets the
// "By hours" report line show which activity(ies) filled each hour instead
// of just a total. Rows for the same BucketStart are adjacent, ordered by
// Duration descending.
type HourActivityDuration struct {
	BucketStart time.Time
	Name        string
	Emoji       string
	Duration    time.Duration
}

// LearningStats contains values for learning dashboard.
type LearningStats struct {
	TotalWords    int
	DueTodayWords int
	LearnedWords  int
	StreakDays    int
	TimerActive   bool
	TimerInterval int
	NextPushIn    string // human-readable, "" when timer isn't active
}

// LearningCollectionItem is one collection row in list/selection UIs.
type LearningCollectionItem struct {
	ID         int64
	Name       string
	WordCount  int
	Active     bool // included in the review rotation
	IsArchived bool
}

// LearningWordItem is one word pair row within a collection's word list.
type LearningWordItem struct {
	ID           int64
	Term         string
	Translation  string
	Learned      bool
	NextReviewAt time.Time
	IntervalDays int
	Repetitions  int
}

// LearningDueWord is a word picked for review delivery, with its owning
// collection name for display.
type LearningDueWord struct {
	ID             int64
	CollectionName string
	Term           string
	Translation    string
}

// LearningDueUser represents one user that should receive a review push now.
type LearningDueUser struct {
	DBUserID    int64
	TgUserID    int64
	IntervalMin int
}

// LearningCollectionStat is one collection's row in the detailed
// statistics breakdown.
type LearningCollectionStat struct {
	Name         string
	TotalWords   int
	DueWords     int
	LearnedWords int
}

// LearningReviewEntry is one answered review, for the heatmap day
// drill-down ("what words did I study that day").
type LearningReviewEntry struct {
	Term        string
	Translation string
	Correct     bool
	ReviewedAt  time.Time
}

// LearningStatsDetail is the "📈 Statistics" screen's full view model.
type LearningStatsDetail struct {
	Overall        LearningStats
	Collections    []LearningCollectionStat
	ReviewsTotal   int
	ReviewsCorrect int
}

// ChallengeDayStatus is one day's mark within a challenge.
type ChallengeDayStatus string

const (
	ChallengeDayPending ChallengeDayStatus = "pending"
	ChallengeDayDone    ChallengeDayStatus = "done"
	ChallengeDaySkipped ChallengeDayStatus = "skipped"
)

// ChallengeItem is one challenge row in list/selection UIs.
type ChallengeItem struct {
	ID          int64
	Name        string
	StartDate   time.Time
	EndDate     time.Time
	IsArchived  bool
	TotalDays   int
	DoneDays    int
	SkippedDays int
}

// ChallengeDay is one square in a challenge's grid.
type ChallengeDay struct {
	Date   time.Time
	Status ChallengeDayStatus
}

// ChallengeDueUser represents one user whose challenge is due for its
// daily evening push.
type ChallengeDueUser struct {
	ChallengeID   int64
	DBUserID      int64
	TgUserID      int64
	ChallengeName string
	StartDate     time.Time
	EndDate       time.Time
}

// SubscriptionStats contains values for subscription screen.
type SubscriptionStats struct {
	ActivePlan string
	DaysEnd    int
}
