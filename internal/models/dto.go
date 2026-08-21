package models

import "time"

type UserInput struct {
	TgUserID    int64
	UserName    *string
	PhoneNumber *string
	Email       *string
	Language    *string
	TimeZone    *string
}

type ProfileStats struct {
	TgUserID    int64
	UserName    *string
	PhoneNumber *string
	Email       *string
	Language    *string
	TimeZone    *string
}

type MainStats struct {
	CurrentActivityName string
	TodayTracked        time.Duration
	TodaySessions       int
	StreakDays          int
	// TargetMinutes is the current activity's own configured daily target;
	// nil when it hasn't been set (callers fall back to a default).
	TargetMinutes *int
}

type TrackActivityItem struct {
	ID       int64
	Name     string
	Emoji    string
	Selected bool
	// TargetMinutes is the daily time target this activity is measured
	// against on the Track main screen's progress bar; nil means not
	// configured yet (callers fall back to a default).
	TargetMinutes *int
}

type CustomTimerOption struct {
	IntervalMin int
}

type AdminUserRow struct {
	DBID      int64
	TgUserID  int64
	UserName  *string
	CreatedAt time.Time
}

type AdminOverviewStats struct {
	TotalUsers         int
	ActiveTrackTimers  int
	ActiveReviewPushes int
	TotalActivities    int
	TotalCollections   int
	TotalLearningWords int
}

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

type TimerDueUser struct {
	DBUserID    int64
	TgUserID    int64
	IntervalMin int
	// kept so a late reply still credits the original due window, not "now minus interval"
	NextPingAt time.Time
}

type ActivityDurationStat struct {
	ActivityID int64
	Name       string
	Emoji      string
	Duration   time.Duration
	Sessions   int
}

type ReportTodayStats struct {
	TotalTracked  time.Duration
	TotalSessions int
	TopActivities []ActivityDurationStat
}

type ReportPeriodStats struct {
	From          time.Time
	To            time.Time
	TotalTracked  time.Duration
	TotalSessions int
	Activities    []ActivityDurationStat
	Monthly       []MonthDurationStat
}

type MonthDurationStat struct {
	Month    time.Time
	Duration time.Duration
}

// rows sharing BucketStart are adjacent, sorted by Duration descending
type HourActivityDuration struct {
	BucketStart time.Time
	Name        string
	Emoji       string
	Duration    time.Duration
}

// Hard and Easy both count as "correct" for accuracy stats, but shift the schedule differently than Good
type LearningGrade string

const (
	LearningGradeAgain LearningGrade = "again"
	LearningGradeHard  LearningGrade = "hard"
	LearningGradeGood  LearningGrade = "good"
	LearningGradeEasy  LearningGrade = "easy"
)

type LearningStats struct {
	TotalWords    int
	DueTodayWords int
	LearnedWords  int
	StreakDays    int
	TimerActive   bool
	TimerInterval int
	NextPushIn    string // human-readable, "" when timer isn't active
}

type LearningCollectionItem struct {
	ID         int64
	Name       string
	WordCount  int
	Active     bool // included in the review rotation
	IsArchived bool
}

type LearningWordItem struct {
	ID           int64
	Term         string
	Translation  string
	Learned      bool
	NextReviewAt time.Time
	IntervalDays int
	Repetitions  int
}

type LearningDueWord struct {
	ID             int64
	CollectionName string
	Term           string
	Translation    string
}

type LearningDueUser struct {
	DBUserID    int64
	TgUserID    int64
	IntervalMin int
}

type LearningCollectionStat struct {
	Name         string
	TotalWords   int
	DueWords     int
	LearnedWords int
}

type LearningReviewEntry struct {
	Term        string
	Translation string
	Correct     bool
	ReviewedAt  time.Time
}

type LearningStatsDetail struct {
	Overall        LearningStats
	Collections    []LearningCollectionStat
	ReviewsTotal   int
	ReviewsCorrect int
}

type ChallengeDayStatus string

const (
	ChallengeDayPending ChallengeDayStatus = "pending"
	ChallengeDayDone    ChallengeDayStatus = "done"
	ChallengeDaySkipped ChallengeDayStatus = "skipped"
)

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

type ChallengeDay struct {
	Date   time.Time
	Status ChallengeDayStatus
}

type ChallengeDueUser struct {
	ChallengeID   int64
	DBUserID      int64
	TgUserID      int64
	ChallengeName string
	StartDate     time.Time
	EndDate       time.Time
}

// ChallengeDayDetail is the view model behind a day-square tap: that day's
// own status plus streak/trend context. Total/done/skipped/pending counts
// aren't duplicated here — callers already have them via ChallengeItem.
type ChallengeDayDetail struct {
	Day           time.Time
	Status        ChallengeDayStatus
	CurrentStreak int
	BestStreak    int
	// Trend holds up to a fixed lookback window of days ending at Day,
	// oldest first, clipped to the challenge's own start.
	Trend []ChallengeDayStatus
}

// RoadmapCardKind is what a card actually is — a topic to study, or a
// concrete resource. Freeform text stays the card's body; this only drives
// the icon and lets the user see at a glance what a checklist is made of.
type RoadmapCardKind string

const (
	RoadmapCardTopic   RoadmapCardKind = "topic"
	RoadmapCardArticle RoadmapCardKind = "article"
	RoadmapCardBook    RoadmapCardKind = "book"
	RoadmapCardLecture RoadmapCardKind = "lecture"
)

// Card difficulty, 1-3 (kept in sync with chk_roadmap_card_difficulty).
// This is what makes the plan walkable easiest-first instead of in paste
// order — the digest offers the simplest pending card next.
const (
	RoadmapCardEasy   = 1
	RoadmapCardMedium = 2
	RoadmapCardHard   = 3
)

// RoadmapGoalItem is one outcome the user is working toward ("reach
// mid-level"), aggregating every technology and card underneath it.
type RoadmapGoalItem struct {
	ID            int64
	Name          string
	IsArchived    bool
	TotalRoadmaps int
	TotalCards    int
	DoneCards     int
}

// RoadmapItem is one technology inside a goal, with its card counts for the
// "done/total" display. GoalID is nil for a technology not attached to any
// goal (a v1 leftover, or one whose goal was deleted).
type RoadmapItem struct {
	ID              int64
	GoalID          *int64
	Name            string
	MasteryCriteria string // free text: what "I know this" means for the user
	Active          bool   // included in the push digest
	IsArchived      bool
	TotalCards      int
	DoneCards       int
}

// RoadmapCardItem is one checklist card within a technology.
type RoadmapCardItem struct {
	ID         int64
	Text       string
	Kind       RoadmapCardKind
	Difficulty int
	IsDone     bool
	DoneAt     *time.Time // nil while pending
}

// RoadmapDueUser represents one user that should receive a roadmap digest
// push now.
type RoadmapDueUser struct {
	DBUserID    int64
	TgUserID    int64
	IntervalMin int
}

// RoadmapDigestCard is one pending card picked for a digest push, with its
// owning technology's name for display.
type RoadmapDigestCard struct {
	ID          int64
	RoadmapID   int64
	RoadmapName string
	Text        string
	Kind        RoadmapCardKind
	Difficulty  int
}

// RoadmapQuiz is one generated question about a card, carried through the
// UI between asking and grading. Nothing about a quiz is persisted: the
// question lives in the session while the user types an answer, and the only
// lasting effect a quiz can have is the user choosing to tick the card done.
type RoadmapQuiz struct {
	CardID   int64
	CardText string
	Question string
}

// RoadmapQuizVerdict is how an answer was judged. Three levels rather than a
// pass/fail bit, because "you got the main idea but missed X" is the common
// case and the most useful thing to say back.
type RoadmapQuizVerdict string

const (
	RoadmapQuizCorrect RoadmapQuizVerdict = "correct"
	RoadmapQuizPartial RoadmapQuizVerdict = "partial"
	RoadmapQuizWrong   RoadmapQuizVerdict = "wrong"
)

// RoadmapQuizGrade is the judgement of one answer, in the user's language.
type RoadmapQuizGrade struct {
	Verdict  RoadmapQuizVerdict
	Feedback string
}

// RoadmapCardStat is one technology's row in the statistics breakdown.
type RoadmapCardStat struct {
	GoalName        string
	Name            string
	MasteryCriteria string
	TotalCards      int
	DoneCards       int
}

// RoadmapStats contains values for the Roadmap dashboard.
type RoadmapStats struct {
	TotalGoals    int
	TotalRoadmaps int
	TotalCards    int
	DoneCards     int
	PendingCards  int
	TimerActive   bool
	TimerInterval int
	NextPushIn    string // human-readable, "" when the push isn't active
}

// RoadmapStatsDetail is the progress screen's full view model.
type RoadmapStatsDetail struct {
	Overall  RoadmapStats
	Goals    []RoadmapGoalItem
	Roadmaps []RoadmapCardStat
}

type SubscriptionStats struct {
	ActivePlan string
	DaysEnd    int
}
