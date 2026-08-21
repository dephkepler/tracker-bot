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
func (f *fakeTrackSvc) GetPeriodReport(context.Context, int64, time.Time, time.Time, []int64, *time.Location) (models.ReportPeriodStats, error) {
	panic("not used yet")
}
func (f *fakeTrackSvc) GetMonthDailyTotals(context.Context, int64, time.Time, []int64, *time.Location) (map[int]time.Duration, error) {
	panic("not used yet")
}
func (f *fakeTrackSvc) GetPeriodBuckets(context.Context, int64, time.Time, time.Time, []int64, string, *time.Location) ([]time.Time, []time.Duration, error) {
	panic("not used yet")
}
func (f *fakeTrackSvc) GetHourlyBucketsByActivity(context.Context, int64, time.Time, time.Time, []int64, *time.Location) ([]models.HourActivityDuration, error) {
	panic("not used yet")
}
func (f *fakeTrackSvc) GetTrackedDaysInRange(context.Context, int64, time.Time, time.Time, *time.Location) ([]time.Time, error) {
	panic("not used yet")
}

var _ service.TrackerService = (*fakeTrackSvc)(nil)
