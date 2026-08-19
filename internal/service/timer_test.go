package service

import (
	"context"
	"errors"
	"slices"
	"sort"
	"testing"
	"time"
	"tracker-bot/internal/models"
)

// --- fakes -----------------------------------------------------------------
//
// These are hand-written in-memory fakes (not mocks generated from the
// interface) so the tests below stay at the bottom of the pyramid: fast,
// no network/DB, and exercising only timerService's own branching logic.
// Real SQL behavior (constraints, cascades, idempotent upserts) is covered
// separately by the *_integration_test.go files in internal/repo, which run
// against a real Postgres.

type timerRow struct {
	intervalMin int
	nextPingAt  time.Time
	enabled     bool
}

type fakeTimerRepo struct {
	rows map[int64]timerRow
}

func newFakeTimerRepo() *fakeTimerRepo {
	return &fakeTimerRepo{rows: map[int64]timerRow{}}
}

func (f *fakeTimerRepo) UpsertInterval(_ context.Context, userID int64, intervalMin int, nextPingAt time.Time) error {
	f.rows[userID] = timerRow{intervalMin: intervalMin, nextPingAt: nextPingAt, enabled: true}
	return nil
}

func (f *fakeTimerRepo) ListDueUsers(_ context.Context, now time.Time, limit int) ([]models.TimerDueUser, error) {
	return nil, nil
}

func (f *fakeTimerRepo) SetNextPing(_ context.Context, userID int64, nextPingAt time.Time) error {
	row, ok := f.rows[userID]
	if !ok {
		return errors.New("no row")
	}
	row.nextPingAt = nextPingAt
	f.rows[userID] = row
	return nil
}

func (f *fakeTimerRepo) GetInterval(_ context.Context, userID int64) (int, error) {
	row, ok := f.rows[userID]
	if !ok || !row.enabled {
		return 0, errors.New("no active timer")
	}
	return row.intervalMin, nil
}

func (f *fakeTimerRepo) GetSettings(_ context.Context, userID int64) (int, time.Time, bool, error) {
	row, ok := f.rows[userID]
	if !ok {
		return 0, time.Time{}, false, nil
	}
	return row.intervalMin, row.nextPingAt, row.enabled, nil
}

func (f *fakeTimerRepo) Disable(_ context.Context, userID int64) error {
	row, ok := f.rows[userID]
	if !ok {
		return errors.New("no row")
	}
	row.enabled = false
	f.rows[userID] = row
	return nil
}

type fakeCustomTimerRepo struct {
	byUser map[int64][]int
}

func newFakeCustomTimerRepo() *fakeCustomTimerRepo {
	return &fakeCustomTimerRepo{byUser: map[int64][]int{}}
}

func (f *fakeCustomTimerRepo) Create(_ context.Context, userID int64, intervalMin int) error {
	if slices.Contains(f.byUser[userID], intervalMin) {
		return nil // idempotent, mirrors ON CONFLICT DO NOTHING
	}
	f.byUser[userID] = append(f.byUser[userID], intervalMin)
	return nil
}

func (f *fakeCustomTimerRepo) ListByUser(_ context.Context, userID int64) ([]int, error) {
	out := append([]int(nil), f.byUser[userID]...)
	sort.Ints(out)
	return out, nil
}

func (f *fakeCustomTimerRepo) Delete(_ context.Context, userID int64, intervalMin int) error {
	items := f.byUser[userID]
	for i, m := range items {
		if m == intervalMin {
			f.byUser[userID] = append(items[:i], items[i+1:]...)
			return nil
		}
	}
	return models.ErrCustomTimerNotFound
}

func (f *fakeCustomTimerRepo) Count(_ context.Context, userID int64) (int, error) {
	return len(f.byUser[userID]), nil
}

type fakeSessionRepo struct{}

func (fakeSessionRepo) CreateRetroSession(context.Context, int64, int64, int, string) error {
	return nil
}

func newTestTimerService() (*fakeTimerRepo, *fakeCustomTimerRepo, TimerService) {
	timerRepo := newFakeTimerRepo()
	customRepo := newFakeCustomTimerRepo()
	svc := NewTimerService(timerRepo, fakeSessionRepo{}, customRepo)
	return timerRepo, customRepo, svc
}

// --- Activate: preserves a running schedule ---------------------------------

func TestTimerService_Activate_FreshSchedule(t *testing.T) {
	timerRepo, _, svc := newTestTimerService()

	before := time.Now().UTC()
	if err := svc.Activate(context.Background(), 1, 30); err != nil {
		t.Fatalf("activate: %v", err)
	}

	row := timerRepo.rows[1]
	wantMin := before.Add(30 * time.Minute)
	if row.nextPingAt.Before(wantMin) || row.nextPingAt.After(wantMin.Add(time.Second)) {
		t.Fatalf("nextPingAt = %v, want ~%v", row.nextPingAt, wantMin)
	}
}

func TestTimerService_Activate_PreservesRunningSchedule_SameInterval(t *testing.T) {
	timerRepo, _, svc := newTestTimerService()
	ctx := context.Background()

	original := time.Now().UTC().Add(20 * time.Minute) // e.g. activated at 12:00 for 30 min, now it's 12:10
	timerRepo.rows[1] = timerRow{intervalMin: 30, nextPingAt: original, enabled: true}

	// User adds a new activity and re-activates with the SAME interval, as
	// described by the bug report this covers: this must not push the
	// already-scheduled prompt back by a full interval.
	if err := svc.Activate(ctx, 1, 30); err != nil {
		t.Fatalf("activate: %v", err)
	}

	got := timerRepo.rows[1].nextPingAt
	if !got.Equal(original) {
		t.Fatalf("nextPingAt = %v, want unchanged %v", got, original)
	}
}

func TestTimerService_Activate_ResetsWhenIntervalChanges(t *testing.T) {
	timerRepo, _, svc := newTestTimerService()
	ctx := context.Background()

	original := time.Now().UTC().Add(20 * time.Minute)
	timerRepo.rows[1] = timerRow{intervalMin: 30, nextPingAt: original, enabled: true}

	before := time.Now().UTC()
	if err := svc.Activate(ctx, 1, 15); err != nil {
		t.Fatalf("activate: %v", err)
	}

	got := timerRepo.rows[1].nextPingAt
	if got.Equal(original) {
		t.Fatalf("nextPingAt unchanged at %v, want recomputed for the new 15 min interval", got)
	}
	wantMin := before.Add(15 * time.Minute)
	if got.Before(wantMin) || got.After(wantMin.Add(time.Second)) {
		t.Fatalf("nextPingAt = %v, want ~%v", got, wantMin)
	}
}

func TestTimerService_Activate_ResetsWhenOverdue(t *testing.T) {
	timerRepo, _, svc := newTestTimerService()
	ctx := context.Background()

	past := time.Now().UTC().Add(-5 * time.Minute) // scheduler should have already fired this
	timerRepo.rows[1] = timerRow{intervalMin: 30, nextPingAt: past, enabled: true}

	before := time.Now().UTC()
	if err := svc.Activate(ctx, 1, 30); err != nil {
		t.Fatalf("activate: %v", err)
	}

	got := timerRepo.rows[1].nextPingAt
	wantMin := before.Add(30 * time.Minute)
	if got.Before(wantMin) {
		t.Fatalf("nextPingAt = %v, want recomputed forward from now (~%v)", got, wantMin)
	}
}

func TestTimerService_Activate_ResetsWhenDisabled(t *testing.T) {
	timerRepo, _, svc := newTestTimerService()
	ctx := context.Background()

	future := time.Now().UTC().Add(20 * time.Minute)
	timerRepo.rows[1] = timerRow{intervalMin: 30, nextPingAt: future, enabled: false}

	before := time.Now().UTC()
	if err := svc.Activate(ctx, 1, 30); err != nil {
		t.Fatalf("activate: %v", err)
	}

	row := timerRepo.rows[1]
	if !row.enabled {
		t.Fatalf("enabled = false, want true after Activate")
	}
	wantMin := before.Add(30 * time.Minute)
	if row.nextPingAt.Before(wantMin) {
		t.Fatalf("nextPingAt = %v, want recomputed forward from now (~%v)", row.nextPingAt, wantMin)
	}
}

func TestTimerService_Activate_InvalidArgs(t *testing.T) {
	_, _, svc := newTestTimerService()
	ctx := context.Background()

	if err := svc.Activate(ctx, 0, 30); err == nil {
		t.Fatal("activate with userID=0: want error")
	}
	if err := svc.Activate(ctx, 1, 0); err == nil {
		t.Fatal("activate with intervalMin=0: want error")
	}
}

// --- Custom intervals --------------------------------------------------------

func TestTimerService_AddCustomInterval_Success(t *testing.T) {
	_, customRepo, svc := newTestTimerService()
	ctx := context.Background()

	if err := svc.AddCustomInterval(ctx, 1, 1); err != nil {
		t.Fatalf("add custom interval: %v", err)
	}
	if got := customRepo.byUser[1]; len(got) != 1 || got[0] != 1 {
		t.Fatalf("byUser[1] = %v, want [1]", got)
	}
}

func TestTimerService_AddCustomInterval_OutOfRange(t *testing.T) {
	_, customRepo, svc := newTestTimerService()
	ctx := context.Background()

	for _, minutes := range []int{0, -5, 361, 1000} {
		err := svc.AddCustomInterval(ctx, 1, minutes)
		if !errors.Is(err, models.ErrCustomTimerInvalidInterval) {
			t.Fatalf("add custom interval %d: got %v, want ErrCustomTimerInvalidInterval", minutes, err)
		}
	}
	if got := customRepo.byUser[1]; len(got) != 0 {
		t.Fatalf("byUser[1] = %v, want empty (nothing should have been persisted)", got)
	}
}

func TestTimerService_AddCustomInterval_LimitReached(t *testing.T) {
	_, customRepo, svc := newTestTimerService()
	ctx := context.Background()

	for i := 1; i <= models.MaxCustomTimersPerUser; i++ {
		if err := svc.AddCustomInterval(ctx, 1, i); err != nil {
			t.Fatalf("add custom interval %d: %v", i, err)
		}
	}

	err := svc.AddCustomInterval(ctx, 1, 100)
	if !errors.Is(err, models.ErrCustomTimerLimitReached) {
		t.Fatalf("add custom interval past limit: got %v, want ErrCustomTimerLimitReached", err)
	}
	if got := len(customRepo.byUser[1]); got != models.MaxCustomTimersPerUser {
		t.Fatalf("byUser[1] has %d entries, want %d (the over-limit add must not persist)", got, models.MaxCustomTimersPerUser)
	}
}

func TestTimerService_AddCustomInterval_DuplicateIsNoop(t *testing.T) {
	_, customRepo, svc := newTestTimerService()
	ctx := context.Background()

	if err := svc.AddCustomInterval(ctx, 1, 45); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := svc.AddCustomInterval(ctx, 1, 45); err != nil {
		t.Fatalf("add again: %v", err)
	}

	if got := customRepo.byUser[1]; len(got) != 1 {
		t.Fatalf("byUser[1] = %v, want a single 45 (duplicate must not be stored twice)", got)
	}
}

func TestTimerService_ListCustomIntervals_SortedAscending(t *testing.T) {
	_, _, svc := newTestTimerService()
	ctx := context.Background()

	for _, m := range []int{30, 5, 1, 45} {
		if err := svc.AddCustomInterval(ctx, 1, m); err != nil {
			t.Fatalf("add %d: %v", m, err)
		}
	}

	got, err := svc.ListCustomIntervals(ctx, 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []int{1, 5, 30, 45}
	if len(got) != len(want) {
		t.Fatalf("list = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("list = %v, want %v", got, want)
		}
	}
}

func TestTimerService_RemoveCustomInterval(t *testing.T) {
	_, customRepo, svc := newTestTimerService()
	ctx := context.Background()

	if err := svc.AddCustomInterval(ctx, 1, 1); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := svc.RemoveCustomInterval(ctx, 1, 1); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if got := customRepo.byUser[1]; len(got) != 0 {
		t.Fatalf("byUser[1] = %v, want empty after remove", got)
	}

	err := svc.RemoveCustomInterval(ctx, 1, 999)
	if !errors.Is(err, models.ErrCustomTimerNotFound) {
		t.Fatalf("remove missing interval: got %v, want ErrCustomTimerNotFound", err)
	}
}

// --- Top of the pyramid: the exact user story from the bug report -----------
//
// "I create timers for 15 and 30 min; I can also set a custom one, e.g. 1
// min, and it shows up next to 15/30; later I can delete it." This walks the
// whole thing end to end against timerService, standing in for a full
// bot-level e2e test — the handler layer talks to the concrete
// *tgbotapi.BotAPI directly rather than through an interface, so it isn't
// substitutable in a test today (a real e2e would need that seam first).
func TestTimerService_CustomTimerUserScenario(t *testing.T) {
	timerRepo, _, svc := newTestTimerService()
	ctx := context.Background()
	const userID = 42

	// 1) Activate the built-in 30 min timer.
	if err := svc.Activate(ctx, userID, 30); err != nil {
		t.Fatalf("activate 30: %v", err)
	}
	runningSchedule := timerRepo.rows[userID].nextPingAt

	// 2) Later that day, add a custom 1 min timer — it must show up next to
	// 15/30 without disturbing the already-running 30 min countdown.
	if err := svc.AddCustomInterval(ctx, userID, 1); err != nil {
		t.Fatalf("add custom 1 min: %v", err)
	}
	custom, err := svc.ListCustomIntervals(ctx, userID)
	if err != nil {
		t.Fatalf("list custom: %v", err)
	}
	if len(custom) != 1 || custom[0] != 1 {
		t.Fatalf("custom intervals = %v, want [1]", custom)
	}
	if !timerRepo.rows[userID].nextPingAt.Equal(runningSchedule) {
		t.Fatalf("adding a custom interval must not touch the running 30 min schedule")
	}

	// 3) User deletes the custom timer again.
	if err := svc.RemoveCustomInterval(ctx, userID, 1); err != nil {
		t.Fatalf("remove custom 1 min: %v", err)
	}
	custom, err = svc.ListCustomIntervals(ctx, userID)
	if err != nil {
		t.Fatalf("list custom after remove: %v", err)
	}
	if len(custom) != 0 {
		t.Fatalf("custom intervals after remove = %v, want empty", custom)
	}
}
