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

// LearningStats contains values for learning dashboard.
type LearningStats struct {
	Language     string
	TotalWords   int
	TodayWords   int
	LearnedWords int
	NextWordIn   string
}

// SubscriptionStats contains values for subscription screen.
type SubscriptionStats struct {
	ActivePlan string
	DaysEnd    int
}
