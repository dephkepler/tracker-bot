package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"tracker-bot/internal/models"
)

// --- fake repo ---------------------------------------------------------
//
// Hand-written in-memory fake, same convention as fakeLearningRepo in
// learning_test.go — real SQL behavior is covered separately by
// internal/repo/roadmap_integration_test.go against a real Postgres.

type fakeRoadmap struct {
	id       int64
	userID   int64
	name     string
	goal     string
	active   bool
	archived bool
}

type fakeCard struct {
	id        int64
	roadmapID int64
	text      string
	isDone    bool
	doneAt    *time.Time
	createdAt time.Time
}

type fakeRoadmapRepo struct {
	nextRoadmapID int64
	nextCardID    int64
	roadmaps      map[int64]*fakeRoadmap
	cards         map[int64]*fakeCard
	pushRows      map[int64]struct {
		intervalMin int
		nextPushAt  time.Time
		enabled     bool
	}
	clock time.Time
}

func newFakeRoadmapRepo() *fakeRoadmapRepo {
	return &fakeRoadmapRepo{
		roadmaps: map[int64]*fakeRoadmap{},
		cards:    map[int64]*fakeCard{},
		pushRows: map[int64]struct {
			intervalMin int
			nextPushAt  time.Time
			enabled     bool
		}{},
		clock: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func (f *fakeRoadmapRepo) CreateRoadmap(_ context.Context, userID int64, name string) (int64, error) {
	for _, r := range f.roadmaps {
		if r.userID == userID && r.name == name {
			return 0, models.ErrRoadmapExists
		}
	}
	f.nextRoadmapID++
	id := f.nextRoadmapID
	f.roadmaps[id] = &fakeRoadmap{id: id, userID: userID, name: name, active: true}
	return id, nil
}

func (f *fakeRoadmapRepo) cardCounts(roadmapID int64) (total, done int) {
	for _, c := range f.cards {
		if c.roadmapID != roadmapID {
			continue
		}
		total++
		if c.isDone {
			done++
		}
	}
	return total, done
}

func (f *fakeRoadmapRepo) item(r *fakeRoadmap) models.RoadmapItem {
	total, done := f.cardCounts(r.id)
	return models.RoadmapItem{
		ID: r.id, Name: r.name, Goal: r.goal,
		Active: r.active, IsArchived: r.archived,
		TotalCards: total, DoneCards: done,
	}
}

func (f *fakeRoadmapRepo) ListRoadmaps(_ context.Context, userID int64, archived bool) ([]models.RoadmapItem, error) {
	out := make([]models.RoadmapItem, 0)
	for _, r := range f.roadmaps {
		if r.userID != userID || r.archived != archived {
			continue
		}
		out = append(out, f.item(r))
	}
	return out, nil
}

func (f *fakeRoadmapRepo) CountRoadmaps(_ context.Context, userID int64) (int, error) {
	n := 0
	for _, r := range f.roadmaps {
		if r.userID == userID && !r.archived {
			n++
		}
	}
	return n, nil
}

func (f *fakeRoadmapRepo) find(userID, roadmapID int64) (*fakeRoadmap, error) {
	r, ok := f.roadmaps[roadmapID]
	if !ok || r.userID != userID {
		return nil, models.ErrRoadmapNotFound
	}
	return r, nil
}

func (f *fakeRoadmapRepo) GetRoadmap(_ context.Context, userID, roadmapID int64) (models.RoadmapItem, error) {
	r, err := f.find(userID, roadmapID)
	if err != nil {
		return models.RoadmapItem{}, err
	}
	return f.item(r), nil
}

func (f *fakeRoadmapRepo) RenameRoadmap(_ context.Context, userID, roadmapID int64, newName string) error {
	r, err := f.find(userID, roadmapID)
	if err != nil {
		return err
	}
	for _, other := range f.roadmaps {
		if other.userID == userID && other.id != roadmapID && other.name == newName {
			return models.ErrRoadmapExists
		}
	}
	r.name = newName
	return nil
}

func (f *fakeRoadmapRepo) SetGoal(_ context.Context, userID, roadmapID int64, goal string) error {
	r, err := f.find(userID, roadmapID)
	if err != nil {
		return err
	}
	r.goal = goal
	return nil
}

func (f *fakeRoadmapRepo) ToggleRoadmapActive(_ context.Context, userID, roadmapID int64) error {
	r, err := f.find(userID, roadmapID)
	if err != nil {
		return err
	}
	r.active = !r.active
	return nil
}

func (f *fakeRoadmapRepo) ArchiveRoadmap(_ context.Context, userID, roadmapID int64) error {
	r, err := f.find(userID, roadmapID)
	if err != nil {
		return err
	}
	r.archived = true
	return nil
}

func (f *fakeRoadmapRepo) RestoreRoadmap(_ context.Context, userID, roadmapID int64) error {
	r, err := f.find(userID, roadmapID)
	if err != nil {
		return err
	}
	r.archived = false
	return nil
}

func (f *fakeRoadmapRepo) DeleteRoadmapForever(_ context.Context, userID, roadmapID int64) error {
	if _, err := f.find(userID, roadmapID); err != nil {
		return err
	}
	delete(f.roadmaps, roadmapID)
	for id, c := range f.cards {
		if c.roadmapID == roadmapID {
			delete(f.cards, id)
		}
	}
	return nil
}

func (f *fakeRoadmapRepo) AddCards(_ context.Context, roadmapID int64, texts []string) (int, error) {
	for _, text := range texts {
		f.nextCardID++
		f.clock = f.clock.Add(time.Second)
		f.cards[f.nextCardID] = &fakeCard{id: f.nextCardID, roadmapID: roadmapID, text: text, createdAt: f.clock}
	}
	return len(texts), nil
}

func (f *fakeRoadmapRepo) ListCards(_ context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error) {
	if _, err := f.find(userID, roadmapID); err != nil {
		return nil, err
	}
	out := make([]models.RoadmapCardItem, 0)
	for _, c := range f.cards {
		if c.roadmapID != roadmapID {
			continue
		}
		out = append(out, models.RoadmapCardItem{ID: c.id, Text: c.text, IsDone: c.isDone, DoneAt: c.doneAt})
	}
	return out, nil
}

func (f *fakeRoadmapRepo) ToggleCardDone(_ context.Context, userID, cardID int64) (int64, error) {
	c, ok := f.cards[cardID]
	if !ok {
		return 0, models.ErrRoadmapCardNotFound
	}
	if _, err := f.find(userID, c.roadmapID); err != nil {
		return 0, models.ErrRoadmapCardNotFound
	}
	c.isDone = !c.isDone
	if c.isDone {
		now := f.clock
		c.doneAt = &now
	} else {
		c.doneAt = nil
	}
	return c.roadmapID, nil
}

func (f *fakeRoadmapRepo) DeleteCard(_ context.Context, userID, cardID int64) error {
	c, ok := f.cards[cardID]
	if !ok {
		return models.ErrRoadmapCardNotFound
	}
	if _, err := f.find(userID, c.roadmapID); err != nil {
		return models.ErrRoadmapCardNotFound
	}
	delete(f.cards, cardID)
	return nil
}

func (f *fakeRoadmapRepo) UpsertPushInterval(_ context.Context, userID int64, intervalMin int, nextPushAt time.Time) error {
	if intervalMin <= 0 || intervalMin > 1440 {
		return models.ErrRoadmapInvalidInterval
	}
	f.pushRows[userID] = struct {
		intervalMin int
		nextPushAt  time.Time
		enabled     bool
	}{intervalMin, nextPushAt, true}
	return nil
}

func (f *fakeRoadmapRepo) GetPushSettings(_ context.Context, userID int64) (int, time.Time, bool, error) {
	row, ok := f.pushRows[userID]
	if !ok {
		return 0, time.Time{}, false, nil
	}
	return row.intervalMin, row.nextPushAt, row.enabled, nil
}

func (f *fakeRoadmapRepo) SetNextPush(_ context.Context, userID int64, nextPushAt time.Time) error {
	row, ok := f.pushRows[userID]
	if !ok {
		return models.ErrRoadmapNotFound
	}
	row.nextPushAt = nextPushAt
	f.pushRows[userID] = row
	return nil
}

func (f *fakeRoadmapRepo) DisablePush(_ context.Context, userID int64) error {
	row, ok := f.pushRows[userID]
	if !ok {
		return nil
	}
	row.enabled = false
	f.pushRows[userID] = row
	return nil
}

func (f *fakeRoadmapRepo) ListDueUsers(_ context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error) {
	out := make([]models.RoadmapDueUser, 0)
	for userID, row := range f.pushRows {
		if !row.enabled || row.nextPushAt.IsZero() || row.nextPushAt.After(now) {
			continue
		}
		if len(out) >= limit {
			break
		}
		out = append(out, models.RoadmapDueUser{DBUserID: userID, TgUserID: userID * 10, IntervalMin: row.intervalMin})
	}
	return out, nil
}

func (f *fakeRoadmapRepo) PickDigestCards(_ context.Context, userID int64, perRoadmapCap, totalCap int) ([]models.RoadmapDigestCard, error) {
	// Mirrors the SQL's ordering: pending cards from active, non-archived
	// roadmaps, oldest first, capped per roadmap and overall.
	pending := make([]*fakeCard, 0)
	for _, c := range f.cards {
		if c.isDone {
			continue
		}
		r, ok := f.roadmaps[c.roadmapID]
		if !ok || r.userID != userID || !r.active || r.archived {
			continue
		}
		pending = append(pending, c)
	}
	for i := 1; i < len(pending); i++ {
		for j := i; j > 0 && pending[j].createdAt.Before(pending[j-1].createdAt); j-- {
			pending[j], pending[j-1] = pending[j-1], pending[j]
		}
	}

	perRoadmap := map[int64]int{}
	out := make([]models.RoadmapDigestCard, 0)
	for _, c := range pending {
		if perRoadmap[c.roadmapID] >= perRoadmapCap {
			continue
		}
		if len(out) >= totalCap {
			break
		}
		perRoadmap[c.roadmapID]++
		out = append(out, models.RoadmapDigestCard{
			ID: c.id, RoadmapID: c.roadmapID,
			RoadmapName: f.roadmaps[c.roadmapID].name, Text: c.text,
		})
	}
	return out, nil
}

func (f *fakeRoadmapRepo) CountCards(_ context.Context, userID int64) (int, int, error) {
	total, done := 0, 0
	for _, c := range f.cards {
		r, ok := f.roadmaps[c.roadmapID]
		if !ok || r.userID != userID || r.archived {
			continue
		}
		total++
		if c.isDone {
			done++
		}
	}
	return total, done, nil
}

func (f *fakeRoadmapRepo) GetRoadmapCardStats(_ context.Context, userID int64) ([]models.RoadmapCardStat, error) {
	out := make([]models.RoadmapCardStat, 0)
	for _, r := range f.roadmaps {
		if r.userID != userID || r.archived {
			continue
		}
		total, done := f.cardCounts(r.id)
		out = append(out, models.RoadmapCardStat{Name: r.name, Goal: r.goal, TotalCards: total, DoneCards: done})
	}
	return out, nil
}

// --- tests -------------------------------------------------------------

const testRoadmapUser = int64(7)

func newRoadmapServiceForTest() (RoadmapService, *fakeRoadmapRepo) {
	fake := newFakeRoadmapRepo()
	return NewRoadmapService(fake), fake
}

// TestRoadmapService_CreateEnforcesCap checks the MaxRoadmapsPerUser cap
// bites at exactly the 6th roadmap, not earlier and not later.
func TestRoadmapService_CreateEnforcesCap(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	for i := 0; i < models.MaxRoadmapsPerUser; i++ {
		name := "tech-" + string(rune('a'+i))
		if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, name); err != nil {
			t.Fatalf("create roadmap %d/%d: %v", i+1, models.MaxRoadmapsPerUser, err)
		}
	}

	_, err := srv.CreateRoadmap(ctx, testRoadmapUser, "one-too-many")
	if !errors.Is(err, models.ErrRoadmapLimitReached) {
		t.Fatalf("create beyond cap: got %v, want ErrRoadmapLimitReached", err)
	}
}

// TestRoadmapService_CapIgnoresArchived is the point of having both
// is_archived and is_active: archiving must actually free a slot.
func TestRoadmapService_CapIgnoresArchived(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	ids := make([]int64, 0, models.MaxRoadmapsPerUser)
	for i := 0; i < models.MaxRoadmapsPerUser; i++ {
		id, err := srv.CreateRoadmap(ctx, testRoadmapUser, "tech-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("create roadmap: %v", err)
		}
		ids = append(ids, id)
	}

	if err := srv.ArchiveRoadmap(ctx, testRoadmapUser, ids[0]); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, "freed-slot"); err != nil {
		t.Fatalf("create after archiving one: %v", err)
	}

	// Restoring the archived one would exceed the cap again, so it must be
	// refused rather than quietly making a 6th active roadmap.
	if err := srv.RestoreRoadmap(ctx, testRoadmapUser, ids[0]); !errors.Is(err, models.ErrRoadmapLimitReached) {
		t.Fatalf("restore over cap: got %v, want ErrRoadmapLimitReached", err)
	}
}

// TestRoadmapService_CreateRejectsBadNames guards the same single-line 2-60
// rule Learning uses, so a pasted card list can't become a roadmap name.
func TestRoadmapService_CreateRejectsBadNames(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	for _, name := range []string{"", "x", "Go\nRust", string(make([]byte, 61))} {
		if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, name); !errors.Is(err, models.ErrRoadmapInvalidName) {
			t.Errorf("create %q: got %v, want ErrRoadmapInvalidName", name, err)
		}
	}
}

// TestRoadmapService_AddCardsFromText covers the bulk parser: one card per
// non-blank line, list markers stripped, over-long lines skipped rather
// than truncated.
func TestRoadmapService_AddCardsFromText(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	long := ""
	for len([]rune(long)) <= models.MaxRoadmapCardTextLen {
		long += "x"
	}

	text := "goroutines\n\n   \n- channels\n• select\n  context package  \n" + long
	added, skipped, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, text)
	if err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if added != 4 {
		t.Errorf("added = %d, want 4", added)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1 (the over-long line)", skipped)
	}

	cards, err := srv.ListCards(ctx, testRoadmapUser, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	got := map[string]bool{}
	for _, c := range cards {
		got[c.Text] = true
		if c.IsDone {
			t.Errorf("card %q: IsDone = true, want false on insert", c.Text)
		}
	}
	for _, want := range []string{"goroutines", "channels", "select", "context package"} {
		if !got[want] {
			t.Errorf("missing card %q (list markers should be stripped, whitespace trimmed)", want)
		}
	}
}

// TestRoadmapService_AddCardsFromTextEmpty checks a blank paste is reported
// as such instead of silently adding nothing.
func TestRoadmapService_AddCardsFromTextEmpty(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "\n   \n\n"); !errors.Is(err, models.ErrRoadmapNoCardsParsed) {
		t.Fatalf("add blank cards: got %v, want ErrRoadmapNoCardsParsed", err)
	}
}

// TestRoadmapService_ToggleCardDone checks the flip is two-way and reports
// the owning roadmap, which is what lets a tap on a digest push re-render
// the right screen.
func TestRoadmapService_ToggleCardDone(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "goroutines"); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	cards, err := srv.ListCards(ctx, testRoadmapUser, id)
	if err != nil || len(cards) != 1 {
		t.Fatalf("list cards = (%d cards, %v), want (1, nil)", len(cards), err)
	}

	gotRoadmapID, err := srv.ToggleCardDone(ctx, testRoadmapUser, cards[0].ID)
	if err != nil {
		t.Fatalf("toggle done: %v", err)
	}
	if gotRoadmapID != id {
		t.Errorf("toggle returned roadmap id %d, want %d", gotRoadmapID, id)
	}

	stats, err := srv.GetRoadmapStats(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.DoneCards != 1 || stats.PendingCards != 0 {
		t.Errorf("after toggle: done=%d pending=%d, want 1/0", stats.DoneCards, stats.PendingCards)
	}

	if _, err := srv.ToggleCardDone(ctx, testRoadmapUser, cards[0].ID); err != nil {
		t.Fatalf("toggle back: %v", err)
	}
	stats, err = srv.GetRoadmapStats(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.DoneCards != 0 || stats.PendingCards != 1 {
		t.Errorf("after toggle back: done=%d pending=%d, want 0/1", stats.DoneCards, stats.PendingCards)
	}
}

// TestRoadmapService_ActivatePreservesNextPush mirrors the Learning
// behavior: re-activating with an unchanged interval must not push the next
// digest further away every time the user re-taps it.
func TestRoadmapService_ActivatePreservesNextPush(t *testing.T) {
	srv, fake := newRoadmapServiceForTest()
	ctx := context.Background()

	if err := srv.Activate(ctx, testRoadmapUser, 180); err != nil {
		t.Fatalf("activate: %v", err)
	}
	_, first, _, err := fake.GetPushSettings(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}

	if err := srv.Activate(ctx, testRoadmapUser, 180); err != nil {
		t.Fatalf("re-activate same interval: %v", err)
	}
	_, second, _, err := fake.GetPushSettings(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}
	if !second.Equal(first) {
		t.Errorf("re-activating with the same interval moved next_push_at %v -> %v, want unchanged", first, second)
	}

	// A different interval must reschedule.
	if err := srv.Activate(ctx, testRoadmapUser, 60); err != nil {
		t.Fatalf("activate new interval: %v", err)
	}
	_, third, _, err := fake.GetPushSettings(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}
	if third.Equal(first) {
		t.Errorf("changing the interval left next_push_at at %v, want rescheduled", third)
	}
}

// TestRoadmapService_ActivateRejectsBadInterval keeps the service in sync
// with the chk_roadmap_interval_min_range DB constraint.
func TestRoadmapService_ActivateRejectsBadInterval(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	for _, min := range []int{0, -5, 1441} {
		if err := srv.Activate(ctx, testRoadmapUser, min); !errors.Is(err, models.ErrRoadmapInvalidInterval) {
			t.Errorf("activate %d min: got %v, want ErrRoadmapInvalidInterval", min, err)
		}
	}
}

// TestRoadmapService_PickDigestCapsPerRoadmap is why the digest uses a
// window function: one long checklist must not crowd out the others.
func TestRoadmapService_PickDigestCapsPerRoadmap(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	big, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	small, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Kafka")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	// The big roadmap's cards are all older, so without the per-roadmap cap
	// they would fill the whole digest by themselves.
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, big, "a\nb\nc\nd\ne\nf"); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, small, "x\ny"); err != nil {
		t.Fatalf("add cards: %v", err)
	}

	digest, err := srv.PickDigestCards(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("pick digest: %v", err)
	}
	if len(digest) > models.RoadmapDigestMaxCards {
		t.Fatalf("digest has %d cards, want at most %d", len(digest), models.RoadmapDigestMaxCards)
	}
	perRoadmap := map[int64]int{}
	for _, c := range digest {
		perRoadmap[c.RoadmapID]++
	}
	if perRoadmap[big] > models.RoadmapDigestPerRoadmapCap {
		t.Errorf("big roadmap contributed %d cards, want at most %d", perRoadmap[big], models.RoadmapDigestPerRoadmapCap)
	}
	if perRoadmap[small] == 0 {
		t.Error("small roadmap contributed nothing — the per-roadmap cap should leave room for it")
	}

	// Deactivated roadmaps drop out of the digest entirely.
	if err := srv.ToggleRoadmapActive(ctx, testRoadmapUser, small); err != nil {
		t.Fatalf("toggle active: %v", err)
	}
	digest, err = srv.PickDigestCards(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("pick digest: %v", err)
	}
	for _, c := range digest {
		if c.RoadmapID == small {
			t.Errorf("deactivated roadmap %d still in digest", small)
		}
	}
}

// TestRoadmapService_SetGoal covers the length cap and the newline collapse
// that keeps a goal renderable as one line.
func TestRoadmapService_SetGoal(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	if err := srv.SetGoal(ctx, testRoadmapUser, id, " write\na service "); err != nil {
		t.Fatalf("set goal: %v", err)
	}
	item, err := srv.Roadmap(ctx, testRoadmapUser, id)
	if err != nil {
		t.Fatalf("get roadmap: %v", err)
	}
	if item.Goal != "write a service" {
		t.Errorf("goal = %q, want %q", item.Goal, "write a service")
	}

	long := ""
	for len([]rune(long)) <= models.MaxRoadmapGoalLen {
		long += "y"
	}
	if err := srv.SetGoal(ctx, testRoadmapUser, id, long); !errors.Is(err, models.ErrRoadmapGoalTooLong) {
		t.Fatalf("set over-long goal: got %v, want ErrRoadmapGoalTooLong", err)
	}
}

// TestRoadmapService_GetStatsDetail checks the per-roadmap breakdown lines
// up with the overall totals.
func TestRoadmapService_GetStatsDetail(t *testing.T) {
	srv, _ := newRoadmapServiceForTest()
	ctx := context.Background()

	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "a\nb\nc"); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	cards, err := srv.ListCards(ctx, testRoadmapUser, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if _, err := srv.ToggleCardDone(ctx, testRoadmapUser, cards[0].ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	detail, err := srv.GetStatsDetail(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("stats detail: %v", err)
	}
	if detail.Overall.TotalRoadmaps != 1 || detail.Overall.TotalCards != 3 || detail.Overall.DoneCards != 1 {
		t.Errorf("overall = %+v, want 1 roadmap / 3 cards / 1 done", detail.Overall)
	}
	if len(detail.Roadmaps) != 1 {
		t.Fatalf("per-roadmap rows = %d, want 1", len(detail.Roadmaps))
	}
	if detail.Roadmaps[0].TotalCards != 3 || detail.Roadmaps[0].DoneCards != 1 {
		t.Errorf("roadmap row = %+v, want 3 cards / 1 done", detail.Roadmaps[0])
	}
}
