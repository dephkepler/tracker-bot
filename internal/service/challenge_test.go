package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"tracker-bot/internal/models"
)

// --- fake repo ---------------------------------------------------------

type fakeChallenge struct {
	id         int64
	userID     int64
	name       string
	start, end time.Time
	archived   bool
	nextPush   time.Time
	pushSet    bool
}

type fakeChallengeDay struct {
	challengeID int64
	date        time.Time
	status      models.ChallengeDayStatus
}

type fakeChallengeRepo struct {
	nextID     int64
	challenges map[int64]*fakeChallenge
	days       map[int64][]*fakeChallengeDay
}

func newFakeChallengeRepo() *fakeChallengeRepo {
	return &fakeChallengeRepo{challenges: map[int64]*fakeChallenge{}, days: map[int64][]*fakeChallengeDay{}}
}

func (f *fakeChallengeRepo) Create(_ context.Context, userID int64, name string, startDate, endDate time.Time) (int64, error) {
	for _, c := range f.challenges {
		if c.userID == userID && c.name == name {
			return 0, models.ErrChallengeExists
		}
	}
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if totalDays < 1 || totalDays > 100 {
		return 0, models.ErrChallengeInvalidRange
	}
	f.nextID++
	id := f.nextID
	f.challenges[id] = &fakeChallenge{id: id, userID: userID, name: name, start: startDate, end: endDate}
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		f.days[id] = append(f.days[id], &fakeChallengeDay{challengeID: id, date: d, status: models.ChallengeDayPending})
	}
	return id, nil
}

func (f *fakeChallengeRepo) ListChallenges(_ context.Context, userID int64, archived bool) ([]models.ChallengeItem, error) {
	out := make([]models.ChallengeItem, 0)
	for _, c := range f.challenges {
		if c.userID != userID || c.archived != archived {
			continue
		}
		out = append(out, f.toItem(c))
	}
	return out, nil
}

func (f *fakeChallengeRepo) toItem(c *fakeChallenge) models.ChallengeItem {
	item := models.ChallengeItem{ID: c.id, Name: c.name, StartDate: c.start, EndDate: c.end, IsArchived: c.archived}
	for _, d := range f.days[c.id] {
		item.TotalDays++
		switch d.status {
		case models.ChallengeDayDone:
			item.DoneDays++
		case models.ChallengeDaySkipped:
			item.SkippedDays++
		}
	}
	return item
}

func (f *fakeChallengeRepo) GetChallenge(_ context.Context, userID, challengeID int64) (models.ChallengeItem, error) {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return models.ChallengeItem{}, models.ErrChallengeNotFound
	}
	return f.toItem(c), nil
}

func (f *fakeChallengeRepo) ArchiveChallenge(_ context.Context, userID, challengeID int64) error {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return models.ErrChallengeNotFound
	}
	c.archived = true
	c.pushSet = false
	return nil
}

func (f *fakeChallengeRepo) RestoreChallenge(_ context.Context, userID, challengeID int64) error {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return models.ErrChallengeNotFound
	}
	c.archived = false
	return nil
}

func (f *fakeChallengeRepo) DeleteChallengeForever(_ context.Context, userID, challengeID int64) error {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return models.ErrChallengeNotFound
	}
	delete(f.challenges, challengeID)
	delete(f.days, challengeID)
	return nil
}

func (f *fakeChallengeRepo) ListDays(_ context.Context, userID, challengeID int64) ([]models.ChallengeDay, error) {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return nil, models.ErrChallengeNotFound
	}
	out := make([]models.ChallengeDay, 0)
	for _, d := range f.days[challengeID] {
		out = append(out, models.ChallengeDay{Date: d.date, Status: d.status})
	}
	return out, nil
}

func (f *fakeChallengeRepo) MarkDay(_ context.Context, userID, challengeID int64, day time.Time, status models.ChallengeDayStatus) error {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return models.ErrChallengeNotFound
	}
	for _, d := range f.days[challengeID] {
		if d.date.Equal(day) {
			d.status = status
			return nil
		}
	}
	return models.ErrChallengeDayNotFound
}

func (f *fakeChallengeRepo) GetDayStatus(_ context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error) {
	c, ok := f.challenges[challengeID]
	if !ok || c.userID != userID {
		return "", models.ErrChallengeNotFound
	}
	for _, d := range f.days[challengeID] {
		if d.date.Equal(day) {
			return d.status, nil
		}
	}
	return "", models.ErrChallengeDayNotFound
}

func (f *fakeChallengeRepo) UpsertPushSchedule(_ context.Context, challengeID int64, nextPushAt time.Time) error {
	c, ok := f.challenges[challengeID]
	if !ok {
		return models.ErrChallengeNotFound
	}
	c.nextPush = nextPushAt
	c.pushSet = true
	return nil
}

func (f *fakeChallengeRepo) ClearPushSchedule(_ context.Context, challengeID int64) error {
	c, ok := f.challenges[challengeID]
	if !ok {
		return models.ErrChallengeNotFound
	}
	c.pushSet = false
	return nil
}

func (f *fakeChallengeRepo) ListDueChallenges(_ context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error) {
	out := make([]models.ChallengeDueUser, 0)
	for _, c := range f.challenges {
		if !c.archived && c.pushSet && !c.nextPush.After(now) {
			out = append(out, models.ChallengeDueUser{ChallengeID: c.id, DBUserID: c.userID, ChallengeName: c.name, StartDate: c.start, EndDate: c.end})
		}
	}
	return out, nil
}

func newTestChallengeService() (*fakeChallengeRepo, ChallengeService) {
	repo := newFakeChallengeRepo()
	return repo, NewChallengeService(repo)
}

// --- CreateChallenge validation ------------------------------------------

func TestChallengeService_CreateChallenge_ValidatesName(t *testing.T) {
	_, svc := newTestChallengeService()
	ctx := context.Background()
	loc := time.UTC
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 9)

	cases := []string{"", "a", "multi\nline"}
	for _, name := range cases {
		if _, err := svc.CreateChallenge(ctx, 1, name, start, end, loc); !errors.Is(err, models.ErrChallengeInvalidName) {
			t.Errorf("name %q: got %v, want ErrChallengeInvalidName", name, err)
		}
	}
}

func TestChallengeService_CreateChallenge_ValidatesRange(t *testing.T) {
	_, svc := newTestChallengeService()
	ctx := context.Background()
	loc := time.UTC
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// End before start.
	if _, err := svc.CreateChallenge(ctx, 1, "Too short", start, start.AddDate(0, 0, -1), loc); !errors.Is(err, models.ErrChallengeInvalidRange) {
		t.Errorf("end before start: got %v, want ErrChallengeInvalidRange", err)
	}
	// 101 days — over the cap.
	if _, err := svc.CreateChallenge(ctx, 1, "Too long", start, start.AddDate(0, 0, 100), loc); !errors.Is(err, models.ErrChallengeInvalidRange) {
		t.Errorf("101 days: got %v, want ErrChallengeInvalidRange", err)
	}
	// Exactly 100 days — allowed.
	if _, err := svc.CreateChallenge(ctx, 1, "Exactly 100", start, start.AddDate(0, 0, 99), loc); err != nil {
		t.Errorf("100 days: got %v, want nil", err)
	}
}

func TestChallengeService_CreateChallenge_DuplicateName(t *testing.T) {
	_, svc := newTestChallengeService()
	ctx := context.Background()
	loc := time.UTC
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 9)

	if _, err := svc.CreateChallenge(ctx, 1, "100 days", start, end, loc); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := svc.CreateChallenge(ctx, 1, "100 days", start, end, loc); !errors.Is(err, models.ErrChallengeExists) {
		t.Fatalf("duplicate: got %v, want ErrChallengeExists", err)
	}
}

// --- nextChallengeFireTime / nextDailyFireTime ---------------------------

func TestNextChallengeFireTime_BeforeTodays21_UsesToday(t *testing.T) {
	loc := time.UTC
	from := time.Date(2026, 8, 19, 10, 0, 0, 0, loc) // 10:00, before 21:00
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, loc)
	end := time.Date(2026, 8, 25, 0, 0, 0, 0, loc)

	got, ok := nextChallengeFireTime(loc, from, start, end)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	want := time.Date(2026, 8, 19, 21, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v (today's 21:00)", got, want)
	}
}

func TestNextChallengeFireTime_After21_UsesTomorrow(t *testing.T) {
	loc := time.UTC
	from := time.Date(2026, 8, 19, 22, 0, 0, 0, loc) // 22:00, after 21:00
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, loc)
	end := time.Date(2026, 8, 25, 0, 0, 0, 0, loc)

	got, ok := nextChallengeFireTime(loc, from, start, end)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	want := time.Date(2026, 8, 20, 21, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v (tomorrow's 21:00)", got, want)
	}
}

func TestNextChallengeFireTime_StartInFuture_UsesStartDate(t *testing.T) {
	loc := time.UTC
	from := time.Date(2026, 8, 19, 10, 0, 0, 0, loc)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, loc) // starts later
	end := time.Date(2026, 9, 10, 0, 0, 0, 0, loc)

	got, ok := nextChallengeFireTime(loc, from, start, end)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	want := time.Date(2026, 9, 1, 21, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v (start date's 21:00, not today)", got, want)
	}
}

func TestNextChallengeFireTime_RangeAlreadyEnded(t *testing.T) {
	loc := time.UTC
	from := time.Date(2026, 8, 19, 10, 0, 0, 0, loc)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, loc)
	end := time.Date(2026, 7, 10, 0, 0, 0, 0, loc) // long over

	if _, ok := nextChallengeFireTime(loc, from, start, end); ok {
		t.Fatal("ok = true, want false for an already-ended range")
	}
}

func TestNextDailyFireTime_AdvancesOneDay(t *testing.T) {
	loc := time.UTC
	firedAt := time.Date(2026, 8, 19, 21, 0, 0, 0, loc)
	end := time.Date(2026, 8, 25, 0, 0, 0, 0, loc)

	got, ok := nextDailyFireTime(loc, firedAt, end)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	want := time.Date(2026, 8, 20, 21, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestNextDailyFireTime_StopsAfterEndDate(t *testing.T) {
	loc := time.UTC
	firedAt := time.Date(2026, 8, 25, 21, 0, 0, 0, loc) // fired on the last day
	end := time.Date(2026, 8, 25, 0, 0, 0, 0, loc)

	if _, ok := nextDailyFireTime(loc, firedAt, end); ok {
		t.Fatal("ok = true, want false (no next day within range)")
	}
}

// --- MarkDay / GetDayStatus -----------------------------------------------

func TestChallengeService_MarkDay_UpdatesStatus(t *testing.T) {
	repo, svc := newTestChallengeService()
	ctx := context.Background()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 4)
	id, err := svc.CreateChallenge(ctx, 1, "5 days", start, end, time.UTC)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.MarkDay(ctx, 1, id, start, true); err != nil {
		t.Fatalf("mark done: %v", err)
	}
	status, err := svc.GetDayStatus(ctx, 1, id, start)
	if err != nil {
		t.Fatalf("get day status: %v", err)
	}
	if status != models.ChallengeDayDone {
		t.Fatalf("status = %v, want done", status)
	}

	if err := svc.MarkDay(ctx, 1, id, start.AddDate(0, 0, 1), false); err != nil {
		t.Fatalf("mark skipped: %v", err)
	}
	status, _ = svc.GetDayStatus(ctx, 1, id, start.AddDate(0, 0, 1))
	if status != models.ChallengeDaySkipped {
		t.Fatalf("status = %v, want skipped", status)
	}
	_ = repo
}
