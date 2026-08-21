package web

import (
	"context"
	"time"

	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
	"tracker-bot/internal/service"
)

// Hand-written fakes, same convention as the service package's own tests. Only
// the two methods Identity calls do anything; the rest exist to satisfy the
// interfaces and would be a bug to reach from here.

type fakeEntrySvc struct {
	// dbIDs maps a Telegram id to users.id. A missing key is ErrUserNotFound,
	// which is the "never pressed /start" case.
	dbIDs map[int64]int64
	err   error
	calls int
}

func (f *fakeEntrySvc) GetDBIDByTgUserID(_ context.Context, tgUserID int64) (int64, error) {
	f.calls++
	if f.err != nil {
		return 0, f.err
	}
	id, ok := f.dbIDs[tgUserID]
	if !ok {
		return 0, models.ErrUserNotFound
	}
	return id, nil
}

func (f *fakeEntrySvc) EnsureUser(context.Context, *models.UserInput) (int64, bool, error) {
	panic("web: EnsureUser must never be called — the dashboard is read-only")
}
func (f *fakeEntrySvc) CountUsers(context.Context) (int, error) { panic("not used") }
func (f *fakeEntrySvc) ListUsersPage(context.Context, int, int) ([]models.AdminUserRow, error) {
	panic("not used")
}
func (f *fakeEntrySvc) ListAllTelegramIDs(context.Context) ([]int64, error) { panic("not used") }
func (f *fakeEntrySvc) DeleteUser(context.Context, int64) error             { panic("not used") }

type fakeProfileSvc struct {
	// stats is keyed by Telegram id, matching what profilesvc actually takes.
	stats map[int64]*models.ProfileStats
	err   error
	calls int
}

func (f *fakeProfileSvc) GetProfileStats(_ context.Context, tgUserID int64) (*models.ProfileStats, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	s, ok := f.stats[tgUserID]
	if !ok {
		return nil, models.ErrUserNotFound
	}
	return s, nil
}

func (f *fakeProfileSvc) ChangeLanguage(context.Context, int64, string) error { panic("not used") }
func (f *fakeProfileSvc) ChangeTimeZone(context.Context, int64, string) error { panic("not used") }

func strptr(s string) *string { return &s }

// knownUser is a Telegram id the fakes resolve, with a deliberately
// non-Warsaw zone so any code that quietly falls back to the app default shows
// up as a wrong timezone in the response.
const knownTgUserID int64 = 424242

func newFakes() (*fakeEntrySvc, *fakeProfileSvc) {
	return &fakeEntrySvc{dbIDs: map[int64]int64{knownTgUserID: 7}},
		&fakeProfileSvc{stats: map[int64]*models.ProfileStats{
			knownTgUserID: {
				TgUserID: knownTgUserID,
				Language: strptr("ru"),
				TimeZone: strptr("America/New_York"),
			},
		}}
}

// Compile-time proof the fakes still match the interfaces they stand in for.
var (
	_ service.EntryService   = (*fakeEntrySvc)(nil)
	_ service.ProfileService = (*fakeProfileSvc)(nil)
)

// fakeTrackSvc implements only what the handlers call. Every other method
// panics: reaching one from the web layer would mean a handler is doing
// something the dashboard has no business doing (the API is read-only), and a
// panic says so loudly instead of returning a plausible zero.
type fakeTrackSvc struct {
	main       models.MainStats
	today      models.ReportTodayStats
	activities []models.TrackActivityItem
	archived   []models.TrackActivityItem
	err        error
	// gotLoc records the timezone the handler threaded down, which is the one
	// thing about these calls worth asserting.
	gotLoc *time.Location
	// The range calls record what they were asked for, so a test can assert the
	// window and the expanded activity filter rather than just the response.
	gotFrom, gotTo  time.Time
	gotActivityIDs  []int64
	gotGranularity  string
	period          models.ReportPeriodStats
	buckets         []time.Time
	bucketDurations []time.Duration
	hourly          []models.HourActivityDuration
}

func newFakeTrackSvc() *fakeTrackSvc {
	target := 90
	return &fakeTrackSvc{
		main: models.MainStats{
			CurrentActivityID:    5,
			CurrentActivityName:  "Go",
			CurrentActivityEmoji: "🐹",
			TodayTracked:         42 * time.Minute,
			TodaySessions:        2,
			StreakDays:           7,
			TargetMinutes:        &target,
		},
		today: models.ReportTodayStats{
			TotalTracked:  2 * time.Hour,
			TotalSessions: 5,
			TopActivities: []models.ActivityDurationStat{
				{ActivityID: 5, Name: "Go", Emoji: "🐹", Duration: 90 * time.Minute, Sessions: 3},
				{ActivityID: 6, Name: "Sport", Duration: 30 * time.Minute, Sessions: 2},
			},
		},
		activities: []models.TrackActivityItem{
			{ID: 5, Name: "Go", Emoji: "🐹", Selected: true, TargetMinutes: &target},
			{ID: 6, Name: "Sport"},
		},
		archived: []models.TrackActivityItem{{ID: 9, Name: "Old habit"}},
		period: models.ReportPeriodStats{
			TotalTracked:  4 * time.Hour,
			TotalSessions: 9,
			Activities: []models.ActivityDurationStat{
				{ActivityID: 5, Name: "Go", Emoji: "🐹", Duration: 3 * time.Hour, Sessions: 6},
				{ActivityID: 6, Name: "Sport", Duration: time.Hour, Sessions: 3},
			},
			Monthly: []models.MonthDurationStat{
				// Naive values, exactly as a date_trunc column arrives: UTC
				// attached, local wall clock inside.
				{Month: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC), Duration: time.Hour},
				{Month: time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC), Duration: 3 * time.Hour},
			},
		},
		// Three days summing to the same four hours as the period above, which
		// is what lets a test assert the two agree.
		buckets: []time.Time{
			time.Date(2026, time.August, 19, 0, 0, 0, 0, time.UTC),
			time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC),
			time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC),
		},
		bucketDurations: []time.Duration{90 * time.Minute, time.Hour, 90 * time.Minute},
		hourly: []models.HourActivityDuration{
			// Ordered by bucket then duration desc, as the repository returns.
			{BucketStart: time.Date(2026, time.August, 21, 9, 0, 0, 0, time.UTC), Name: "Go", Emoji: "🐹", Duration: 45 * time.Minute},
			{BucketStart: time.Date(2026, time.August, 21, 9, 0, 0, 0, time.UTC), Name: "Sport", Duration: 15 * time.Minute},
			{BucketStart: time.Date(2026, time.August, 21, 14, 0, 0, 0, time.UTC), Name: "Go", Emoji: "🐹", Duration: 30 * time.Minute},
		},
	}
}

func (f *fakeTrackSvc) GetMainStats(_ context.Context, _ int64, loc *time.Location) (models.MainStats, error) {
	f.gotLoc = loc
	return f.main, f.err
}

func (f *fakeTrackSvc) GetTodayReport(_ context.Context, _ int64, loc *time.Location) (models.ReportTodayStats, error) {
	f.gotLoc = loc
	return f.today, f.err
}

func (f *fakeTrackSvc) ListActivities(context.Context, int64) ([]models.TrackActivityItem, error) {
	return f.activities, f.err
}

func (f *fakeTrackSvc) ListArchivedActivities(context.Context, int64) ([]models.TrackActivityItem, error) {
	return f.archived, f.err
}

func (f *fakeTrackSvc) CreateActivity(context.Context, int64, string, string) (repo.Activity, error) {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) ToggleSelectedActivity(context.Context, int64, int64) error {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) DeleteSelectedActivities(context.Context, int64) (int64, error) {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) ListSelectedActivities(context.Context, int64) ([]models.TrackActivityItem, error) {
	panic("not used")
}
func (f *fakeTrackSvc) ListReminderActivities(context.Context, int64) ([]models.TrackActivityItem, error) {
	panic("not used")
}
func (f *fakeTrackSvc) AddSelectedToReminders(context.Context, int64) (int, error) {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) RemoveFromReminders(context.Context, int64, int64) error {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) ArchiveSelectedActivities(context.Context, int64) (int64, error) {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) RestoreArchivedActivity(context.Context, int64, int64) error {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) DeleteArchivedForever(context.Context, int64, int64) error {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) SetActivityTarget(context.Context, int64, int64, int) error {
	panic("web: the dashboard is read-only")
}
func (f *fakeTrackSvc) GetTodayReportBySelected(context.Context, int64, *time.Location) (models.ReportTodayStats, error) {
	panic("not used")
}
func (f *fakeTrackSvc) GetPeriodReport(_ context.Context, _ int64, from, to time.Time, ids []int64, loc *time.Location) (models.ReportPeriodStats, error) {
	f.gotFrom, f.gotTo, f.gotActivityIDs, f.gotLoc = from, to, ids, loc
	return f.period, f.err
}
func (f *fakeTrackSvc) GetMonthDailyTotals(context.Context, int64, time.Time, []int64, *time.Location) (map[int]time.Duration, error) {
	panic("not used yet")
}
func (f *fakeTrackSvc) GetPeriodBuckets(_ context.Context, _ int64, from, to time.Time, ids []int64, granularity string, loc *time.Location) ([]time.Time, []time.Duration, error) {
	f.gotFrom, f.gotTo, f.gotActivityIDs, f.gotGranularity, f.gotLoc = from, to, ids, granularity, loc
	return f.buckets, f.bucketDurations, f.err
}
func (f *fakeTrackSvc) GetHourlyBucketsByActivity(_ context.Context, _ int64, from, to time.Time, ids []int64, loc *time.Location) ([]models.HourActivityDuration, error) {
	f.gotFrom, f.gotTo, f.gotActivityIDs, f.gotLoc = from, to, ids, loc
	return f.hourly, f.err
}
func (f *fakeTrackSvc) GetTrackedDaysInRange(context.Context, int64, time.Time, time.Time, *time.Location) ([]time.Time, error) {
	panic("not used yet")
}

var _ service.TrackerService = (*fakeTrackSvc)(nil)

// fakeRoadmapSvc embeds the interface rather than stubbing all thirty of its
// methods: an un-overridden one panics on the nil embedded value, which is the
// behaviour wanted — the dashboard has no business calling it. Same trick as
// fakeTargetRepo in the service package's own tests.
type fakeRoadmapSvc struct {
	service.RoadmapService

	goals   []models.RoadmapGoalItem
	techs   map[int64][]models.RoadmapItem
	orphans []models.RoadmapItem
	cards   map[int64][]models.RoadmapCardItem
	err     error

	// setDone records the writes, so a test can assert what was asked for
	// rather than only what came back.
	setDone []cardWrite
}

type cardWrite struct {
	cardID int64
	done   bool
}

func newFakeRoadmapSvc() *fakeRoadmapSvc {
	return &fakeRoadmapSvc{
		goals: []models.RoadmapGoalItem{
			{ID: 1, Name: "выйти на мидла", TotalRoadmaps: 1, TotalCards: 3, DoneCards: 1},
		},
		techs: map[int64][]models.RoadmapItem{
			1: {{ID: 10, Name: "Kafka", MasteryCriteria: "могу отладить отстающего консьюмера", Active: true, TotalCards: 3, DoneCards: 1}},
		},
		orphans: []models.RoadmapItem{{ID: 11, Name: "Docker", TotalCards: 1, DoneCards: 0}},
		cards: map[int64][]models.RoadmapCardItem{
			10: {
				{ID: 100, Text: "Партиции и офсеты", Kind: models.RoadmapCardTopic, Difficulty: 1, IsDone: true, DoneAt: &fakeDoneAt},
				{ID: 101, Text: "Группы консьюмеров", Kind: models.RoadmapCardTopic, Difficulty: 2},
				{ID: 102, Text: "Kafka: The Definitive Guide", Kind: models.RoadmapCardBook, Difficulty: 3},
			},
			11: {{ID: 110, Text: "Слои образов", Kind: models.RoadmapCardTopic, Difficulty: 2}},
		},
	}
}

var fakeDoneAt = time.Date(2026, time.August, 20, 9, 30, 0, 0, time.UTC)

func (f *fakeRoadmapSvc) ListGoals(context.Context, int64) ([]models.RoadmapGoalItem, error) {
	return f.goals, f.err
}

func (f *fakeRoadmapSvc) ListRoadmaps(_ context.Context, _ int64, goalID int64) ([]models.RoadmapItem, error) {
	return f.techs[goalID], f.err
}

func (f *fakeRoadmapSvc) ListOrphanRoadmaps(context.Context, int64) ([]models.RoadmapItem, error) {
	return f.orphans, f.err
}

func (f *fakeRoadmapSvc) ListCards(_ context.Context, _ int64, roadmapID int64) ([]models.RoadmapCardItem, error) {
	return f.cards[roadmapID], f.err
}

func (f *fakeRoadmapSvc) SetCardDone(_ context.Context, _ int64, cardID int64, done bool) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	for _, list := range f.cards {
		for i := range list {
			if list[i].ID == cardID {
				list[i].IsDone = done
				f.setDone = append(f.setDone, cardWrite{cardID: cardID, done: done})
				return 10, nil
			}
		}
	}
	return 0, models.ErrRoadmapCardNotFound
}

// fakeRoadmapAISvc answers without a provider. Its interface is small enough to
// implement outright, and every method here is one the dashboard calls.
type fakeRoadmapAISvc struct {
	enabled  bool
	err      error
	added    int
	rejected int
	question string
	grade    models.RoadmapQuizGrade

	// gotLang records the language threaded down, and gotAnswer what was sent
	// for grading.
	gotLang   string
	gotAnswer string
}

func newFakeRoadmapAISvc() *fakeRoadmapAISvc {
	return &fakeRoadmapAISvc{
		enabled:  true,
		added:    6,
		rejected: 1,
		question: "Как ребалансировка влияет на порядок обработки?",
		grade:    models.RoadmapQuizGrade{Verdict: models.RoadmapQuizPartial, Feedback: "почти"},
	}
}

func (f *fakeRoadmapAISvc) Enabled() bool { return f.enabled }

func (f *fakeRoadmapAISvc) GeneratePlan(_ context.Context, _, _ int64, lang string) (int, int, error) {
	f.gotLang = lang
	if f.err != nil {
		return 0, 0, f.err
	}
	return f.added, f.rejected, nil
}

func (f *fakeRoadmapAISvc) AddCardsFromTextAI(_ context.Context, _, _ int64, _, lang string) (int, int, error) {
	f.gotLang = lang
	return f.added, f.rejected, f.err
}

func (f *fakeRoadmapAISvc) DigestAdvice(_ context.Context, _ int64, _ []models.RoadmapDigestCard, lang string) (string, error) {
	f.gotLang = lang
	return "совет", f.err
}

func (f *fakeRoadmapAISvc) QuizCard(_ context.Context, _, cardID int64, lang string) (models.RoadmapQuiz, error) {
	f.gotLang = lang
	if f.err != nil {
		return models.RoadmapQuiz{}, f.err
	}
	return models.RoadmapQuiz{CardID: cardID, CardText: "Группы консьюмеров", Question: f.question}, nil
}

func (f *fakeRoadmapAISvc) GradeQuizAnswer(_ context.Context, _ models.RoadmapQuiz, answer, lang string) (models.RoadmapQuizGrade, error) {
	f.gotLang, f.gotAnswer = lang, answer
	return f.grade, f.err
}

// testDeps wires a full set of fakes. Tests that need to inspect one reach for
// it through the returned struct rather than building their own set.
func testDeps() Deps {
	entrysvc, profilesvc := newFakes()
	return Deps{
		BotToken:  testBotToken,
		Entry:     entrysvc,
		Profile:   profilesvc,
		Tracker:   newFakeTrackSvc(),
		Roadmap:   newFakeRoadmapSvc(),
		RoadmapAI: newFakeRoadmapAISvc(),
	}
}

var (
	_ service.RoadmapService   = (*fakeRoadmapSvc)(nil)
	_ service.RoadmapAIService = (*fakeRoadmapAISvc)(nil)
)
