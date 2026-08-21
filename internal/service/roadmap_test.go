package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"tracker-bot/internal/models"
)

// Hand-written in-memory fake, same convention as fakeLearningRepo in
// learning_test.go — real SQL behavior is covered separately by
// internal/repo/roadmap_integration_test.go against a real Postgres.

type fakeGoal struct {
	id       int64
	userID   int64
	name     string
	archived bool
	seq      int64
}

type fakeRoadmap struct {
	id       int64
	userID   int64
	goalID   *int64
	name     string
	criteria string
	active   bool
	archived bool
	seq      int64
}

type fakeCard struct {
	id         int64
	roadmapID  int64
	text       string
	kind       models.RoadmapCardKind
	difficulty int
	isDone     bool
	doneAt     *time.Time
	seq        int64
}

type fakePushRow struct {
	intervalMin int
	nextPushAt  time.Time
	enabled     bool
}

type fakeRoadmapRepo struct {
	nextID   int64
	seq      int64
	goals    map[int64]*fakeGoal
	roadmaps map[int64]*fakeRoadmap
	cards    map[int64]*fakeCard
	pushRows map[int64]fakePushRow
	now      time.Time
}

func newFakeRoadmapRepo() *fakeRoadmapRepo {
	return &fakeRoadmapRepo{
		goals:    map[int64]*fakeGoal{},
		roadmaps: map[int64]*fakeRoadmap{},
		cards:    map[int64]*fakeCard{},
		pushRows: map[int64]fakePushRow{},
		now:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func (f *fakeRoadmapRepo) id() int64  { f.nextID++; return f.nextID }
func (f *fakeRoadmapRepo) ord() int64 { f.seq++; return f.seq }

func (f *fakeRoadmapRepo) CreateGoal(_ context.Context, userID int64, name string) (int64, error) {
	for _, g := range f.goals {
		if g.userID == userID && g.name == name {
			return 0, models.ErrRoadmapGoalExists
		}
	}
	id := f.id()
	f.goals[id] = &fakeGoal{id: id, userID: userID, name: name, seq: f.ord()}
	return id, nil
}

func (f *fakeRoadmapRepo) goalItem(g *fakeGoal) models.RoadmapGoalItem {
	item := models.RoadmapGoalItem{ID: g.id, Name: g.name, IsArchived: g.archived}
	for _, r := range f.roadmaps {
		if r.goalID == nil || *r.goalID != g.id || r.archived {
			continue
		}
		item.TotalRoadmaps++
		for _, c := range f.cards {
			if c.roadmapID != r.id {
				continue
			}
			item.TotalCards++
			if c.isDone {
				item.DoneCards++
			}
		}
	}
	return item
}

func (f *fakeRoadmapRepo) ListGoals(_ context.Context, userID int64, archived bool) ([]models.RoadmapGoalItem, error) {
	out := make([]models.RoadmapGoalItem, 0)
	for _, g := range sortedGoals(f.goals) {
		if g.userID != userID || g.archived != archived {
			continue
		}
		out = append(out, f.goalItem(g))
	}
	return out, nil
}

func (f *fakeRoadmapRepo) CountGoals(_ context.Context, userID int64) (int, error) {
	n := 0
	for _, g := range f.goals {
		if g.userID == userID && !g.archived {
			n++
		}
	}
	return n, nil
}

func (f *fakeRoadmapRepo) findGoal(userID, goalID int64) (*fakeGoal, error) {
	g, ok := f.goals[goalID]
	if !ok || g.userID != userID {
		return nil, models.ErrRoadmapGoalNotFound
	}
	return g, nil
}

func (f *fakeRoadmapRepo) GetGoal(_ context.Context, userID, goalID int64) (models.RoadmapGoalItem, error) {
	g, err := f.findGoal(userID, goalID)
	if err != nil {
		return models.RoadmapGoalItem{}, err
	}
	return f.goalItem(g), nil
}

func (f *fakeRoadmapRepo) RenameGoal(_ context.Context, userID, goalID int64, newName string) error {
	g, err := f.findGoal(userID, goalID)
	if err != nil {
		return err
	}
	for _, other := range f.goals {
		if other.userID == userID && other.id != goalID && other.name == newName {
			return models.ErrRoadmapGoalExists
		}
	}
	g.name = newName
	return nil
}

func (f *fakeRoadmapRepo) ArchiveGoal(_ context.Context, userID, goalID int64) error {
	g, err := f.findGoal(userID, goalID)
	if err != nil {
		return err
	}
	g.archived = true
	return nil
}

func (f *fakeRoadmapRepo) RestoreGoal(_ context.Context, userID, goalID int64) error {
	g, err := f.findGoal(userID, goalID)
	if err != nil {
		return err
	}
	g.archived = false
	return nil
}

func (f *fakeRoadmapRepo) DeleteGoalForever(_ context.Context, userID, goalID int64) error {
	if _, err := f.findGoal(userID, goalID); err != nil {
		return err
	}
	delete(f.goals, goalID)
	// Mirrors ON DELETE SET NULL: technologies survive, unattached.
	for _, r := range f.roadmaps {
		if r.goalID != nil && *r.goalID == goalID {
			r.goalID = nil
		}
	}
	return nil
}

func (f *fakeRoadmapRepo) CreateRoadmap(_ context.Context, userID, goalID int64, name string) (int64, error) {
	if _, err := f.findGoal(userID, goalID); err != nil {
		return 0, err
	}
	for _, r := range f.roadmaps {
		if r.userID == userID && r.name == name {
			return 0, models.ErrRoadmapExists
		}
	}
	id := f.id()
	gid := goalID
	f.roadmaps[id] = &fakeRoadmap{id: id, userID: userID, goalID: &gid, name: name, active: true, seq: f.ord()}
	return id, nil
}

func (f *fakeRoadmapRepo) roadmapItem(r *fakeRoadmap) models.RoadmapItem {
	item := models.RoadmapItem{
		ID: r.id, GoalID: r.goalID, Name: r.name, MasteryCriteria: r.criteria,
		Active: r.active, IsArchived: r.archived,
	}
	for _, c := range f.cards {
		if c.roadmapID != r.id {
			continue
		}
		item.TotalCards++
		if c.isDone {
			item.DoneCards++
		}
	}
	return item
}

func (f *fakeRoadmapRepo) ListRoadmaps(_ context.Context, userID int64, goalID *int64, archived bool) ([]models.RoadmapItem, error) {
	out := make([]models.RoadmapItem, 0)
	for _, r := range sortedRoadmaps(f.roadmaps) {
		if r.userID != userID || r.archived != archived {
			continue
		}
		switch {
		case goalID == nil && r.goalID != nil:
			continue
		case goalID != nil && (r.goalID == nil || *r.goalID != *goalID):
			continue
		}
		out = append(out, f.roadmapItem(r))
	}
	return out, nil
}

func (f *fakeRoadmapRepo) ListRoadmapsAnyGoal(_ context.Context, userID int64, archived bool) ([]models.RoadmapItem, error) {
	out := make([]models.RoadmapItem, 0)
	for _, r := range sortedRoadmaps(f.roadmaps) {
		if r.userID != userID || r.archived != archived {
			continue
		}
		out = append(out, f.roadmapItem(r))
	}
	return out, nil
}

func (f *fakeRoadmapRepo) CountRoadmapsInGoal(_ context.Context, userID, goalID int64) (int, error) {
	n := 0
	for _, r := range f.roadmaps {
		if r.userID == userID && !r.archived && r.goalID != nil && *r.goalID == goalID {
			n++
		}
	}
	return n, nil
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

func (f *fakeRoadmapRepo) findRoadmap(userID, roadmapID int64) (*fakeRoadmap, error) {
	r, ok := f.roadmaps[roadmapID]
	if !ok || r.userID != userID {
		return nil, models.ErrRoadmapNotFound
	}
	return r, nil
}

func (f *fakeRoadmapRepo) GetRoadmap(_ context.Context, userID, roadmapID int64) (models.RoadmapItem, error) {
	r, err := f.findRoadmap(userID, roadmapID)
	if err != nil {
		return models.RoadmapItem{}, err
	}
	return f.roadmapItem(r), nil
}

func (f *fakeRoadmapRepo) RenameRoadmap(_ context.Context, userID, roadmapID int64, newName string) error {
	r, err := f.findRoadmap(userID, roadmapID)
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

func (f *fakeRoadmapRepo) SetMasteryCriteria(_ context.Context, userID, roadmapID int64, criteria string) error {
	r, err := f.findRoadmap(userID, roadmapID)
	if err != nil {
		return err
	}
	r.criteria = criteria
	return nil
}

func (f *fakeRoadmapRepo) AssignRoadmapToGoal(_ context.Context, userID, roadmapID, goalID int64) error {
	r, err := f.findRoadmap(userID, roadmapID)
	if err != nil {
		return err
	}
	if _, err := f.findGoal(userID, goalID); err != nil {
		return models.ErrRoadmapNotFound
	}
	gid := goalID
	r.goalID = &gid
	return nil
}

func (f *fakeRoadmapRepo) ToggleRoadmapActive(_ context.Context, userID, roadmapID int64) error {
	r, err := f.findRoadmap(userID, roadmapID)
	if err != nil {
		return err
	}
	r.active = !r.active
	return nil
}

func (f *fakeRoadmapRepo) ArchiveRoadmap(_ context.Context, userID, roadmapID int64) error {
	r, err := f.findRoadmap(userID, roadmapID)
	if err != nil {
		return err
	}
	r.archived = true
	return nil
}

func (f *fakeRoadmapRepo) RestoreRoadmap(_ context.Context, userID, roadmapID int64) error {
	r, err := f.findRoadmap(userID, roadmapID)
	if err != nil {
		return err
	}
	r.archived = false
	return nil
}

func (f *fakeRoadmapRepo) DeleteRoadmapForever(_ context.Context, userID, roadmapID int64) error {
	if _, err := f.findRoadmap(userID, roadmapID); err != nil {
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

func (f *fakeRoadmapRepo) AddCards(_ context.Context, roadmapID int64, cards []models.RoadmapCardItem) (int, error) {
	for _, c := range cards {
		id := f.id()
		f.cards[id] = &fakeCard{
			id: id, roadmapID: roadmapID, text: c.Text,
			kind: c.Kind, difficulty: c.Difficulty, seq: f.ord(),
		}
	}
	return len(cards), nil
}

// Mirrors the SQL's ORDER BY is_done, difficulty, created_at, id.
func (f *fakeRoadmapRepo) ListCards(_ context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error) {
	if _, err := f.findRoadmap(userID, roadmapID); err != nil {
		return nil, err
	}
	picked := make([]*fakeCard, 0)
	for _, c := range f.cards {
		if c.roadmapID == roadmapID {
			picked = append(picked, c)
		}
	}
	sortCards(picked, true)

	out := make([]models.RoadmapCardItem, 0, len(picked))
	for _, c := range picked {
		out = append(out, models.RoadmapCardItem{
			ID: c.id, Text: c.text, Kind: c.kind, Difficulty: c.difficulty,
			IsDone: c.isDone, DoneAt: c.doneAt,
		})
	}
	return out, nil
}

func (f *fakeRoadmapRepo) findCard(userID, cardID int64) (*fakeCard, error) {
	c, ok := f.cards[cardID]
	if !ok {
		return nil, models.ErrRoadmapCardNotFound
	}
	if _, err := f.findRoadmap(userID, c.roadmapID); err != nil {
		return nil, models.ErrRoadmapCardNotFound
	}
	return c, nil
}

func (f *fakeRoadmapRepo) SetCardDone(_ context.Context, userID, cardID int64, done bool) (int64, error) {
	c, err := f.findCard(userID, cardID)
	if err != nil {
		return 0, err
	}
	c.isDone = done
	if done {
		now := f.now
		c.doneAt = &now
	} else {
		c.doneAt = nil
	}
	return c.roadmapID, nil
}

func (f *fakeRoadmapRepo) ToggleCardDone(_ context.Context, userID, cardID int64) (int64, error) {
	c, err := f.findCard(userID, cardID)
	if err != nil {
		return 0, err
	}
	c.isDone = !c.isDone
	if c.isDone {
		now := f.now
		c.doneAt = &now
	} else {
		c.doneAt = nil
	}
	return c.roadmapID, nil
}

func (f *fakeRoadmapRepo) CycleCardDifficulty(_ context.Context, userID, cardID int64) (int64, error) {
	c, err := f.findCard(userID, cardID)
	if err != nil {
		return 0, err
	}
	c.difficulty = (c.difficulty % 3) + 1
	return c.roadmapID, nil
}

func (f *fakeRoadmapRepo) DeleteCard(_ context.Context, userID, cardID int64) error {
	c, err := f.findCard(userID, cardID)
	if err != nil {
		return err
	}
	delete(f.cards, c.id)
	return nil
}

func (f *fakeRoadmapRepo) UpsertPushInterval(_ context.Context, userID int64, intervalMin int, nextPushAt time.Time) error {
	if intervalMin <= 0 || intervalMin > 1440 {
		return models.ErrRoadmapInvalidInterval
	}
	f.pushRows[userID] = fakePushRow{intervalMin, nextPushAt, true}
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
		if !row.enabled || row.nextPushAt.IsZero() || row.nextPushAt.After(now) || len(out) >= limit {
			continue
		}
		out = append(out, models.RoadmapDueUser{DBUserID: userID, TgUserID: userID * 10, IntervalMin: row.intervalMin})
	}
	return out, nil
}

// Mirrors the window function: easiest pending first, capped per technology
// and overall, archived goals excluded.
func (f *fakeRoadmapRepo) PickDigestCards(_ context.Context, userID int64, perRoadmapCap, totalCap int) ([]models.RoadmapDigestCard, error) {
	pending := make([]*fakeCard, 0)
	for _, c := range f.cards {
		if c.isDone {
			continue
		}
		r, ok := f.roadmaps[c.roadmapID]
		if !ok || r.userID != userID || !r.active || r.archived {
			continue
		}
		if r.goalID != nil {
			if g, ok := f.goals[*r.goalID]; ok && g.archived {
				continue
			}
		}
		pending = append(pending, c)
	}
	sortCards(pending, false)

	perRoadmap := map[int64]int{}
	out := make([]models.RoadmapDigestCard, 0)
	for _, c := range pending {
		if perRoadmap[c.roadmapID] >= perRoadmapCap || len(out) >= totalCap {
			continue
		}
		perRoadmap[c.roadmapID]++
		out = append(out, models.RoadmapDigestCard{
			ID: c.id, RoadmapID: c.roadmapID, RoadmapName: f.roadmaps[c.roadmapID].name,
			Text: c.text, Kind: c.kind, Difficulty: c.difficulty,
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
	for _, r := range sortedRoadmaps(f.roadmaps) {
		if r.userID != userID || r.archived {
			continue
		}
		goalName := ""
		if r.goalID != nil {
			if g, ok := f.goals[*r.goalID]; ok {
				goalName = g.name
			}
		}
		item := f.roadmapItem(r)
		out = append(out, models.RoadmapCardStat{
			GoalName: goalName, Name: r.name, MasteryCriteria: r.criteria,
			TotalCards: item.TotalCards, DoneCards: item.DoneCards,
		})
	}
	return out, nil
}

// --- ordering helpers (map iteration is random, the SQL is not) ----------

func sortCards(cards []*fakeCard, byDone bool) {
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0 && cardLess(cards[j], cards[j-1], byDone); j-- {
			cards[j], cards[j-1] = cards[j-1], cards[j]
		}
	}
}

func cardLess(a, b *fakeCard, byDone bool) bool {
	if byDone && a.isDone != b.isDone {
		return !a.isDone
	}
	if a.difficulty != b.difficulty {
		return a.difficulty < b.difficulty
	}
	return a.seq < b.seq
}

func sortedGoals(m map[int64]*fakeGoal) []*fakeGoal {
	out := make([]*fakeGoal, 0, len(m))
	for _, g := range m {
		out = append(out, g)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].seq < out[j-1].seq; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func sortedRoadmaps(m map[int64]*fakeRoadmap) []*fakeRoadmap {
	out := make([]*fakeRoadmap, 0, len(m))
	for _, r := range m {
		out = append(out, r)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].seq < out[j-1].seq; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// --- tests --------------------------------------------------------------

const testRoadmapUser = int64(7)

func newRoadmapServiceForTest(t *testing.T) (RoadmapService, *fakeRoadmapRepo, int64) {
	t.Helper()
	fake := newFakeRoadmapRepo()
	srv := NewRoadmapService(fake)
	goalID, err := srv.CreateGoal(context.Background(), testRoadmapUser, "Выйти на мидла")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	return srv, fake, goalID
}

// The goal cap must bite at the 4th, not earlier and not later.
func TestRoadmapService_GoalCap(t *testing.T) {
	srv, _, _ := newRoadmapServiceForTest(t)
	ctx := context.Background()

	for i := 1; i < models.MaxRoadmapGoalsPerUser; i++ {
		if _, err := srv.CreateGoal(ctx, testRoadmapUser, "goal-"+string(rune('a'+i))); err != nil {
			t.Fatalf("create goal %d: %v", i+1, err)
		}
	}
	if _, err := srv.CreateGoal(ctx, testRoadmapUser, "one-too-many"); !errors.Is(err, models.ErrRoadmapGoalLimitReached) {
		t.Fatalf("create beyond goal cap: got %v, want ErrRoadmapGoalLimitReached", err)
	}
}

// The technology cap is per goal: filling one goal must not block another.
func TestRoadmapService_TechnologyCapIsPerGoal(t *testing.T) {
	srv, _, firstGoal := newRoadmapServiceForTest(t)
	ctx := context.Background()

	for i := 0; i < models.MaxRoadmapsPerGoal; i++ {
		if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, firstGoal, "tech-a"+string(rune('a'+i))); err != nil {
			t.Fatalf("create technology %d: %v", i+1, err)
		}
	}
	if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, firstGoal, "overflow"); !errors.Is(err, models.ErrRoadmapLimitReached) {
		t.Fatalf("create beyond per-goal cap: got %v, want ErrRoadmapLimitReached", err)
	}

	secondGoal, err := srv.CreateGoal(ctx, testRoadmapUser, "Выйти на синьора")
	if err != nil {
		t.Fatalf("create second goal: %v", err)
	}
	if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, secondGoal, "tech-in-second-goal"); err != nil {
		t.Fatalf("a full goal must not block a fresh one: %v", err)
	}
}

// A technology can only be created inside a goal the user owns — the goal id
// arrives from a callback payload.
func TestRoadmapService_CreateRoadmapChecksGoalOwnership(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()

	if _, err := srv.CreateRoadmap(ctx, testRoadmapUser+1, goalID, "Kafka"); !errors.Is(err, models.ErrRoadmapGoalNotFound) {
		t.Fatalf("cross-user create: got %v, want ErrRoadmapGoalNotFound", err)
	}
}

// Archiving a goal frees a slot; restoring over the cap must be refused
// rather than quietly making one goal too many.
func TestRoadmapService_ArchiveFreesGoalSlot(t *testing.T) {
	srv, _, firstGoal := newRoadmapServiceForTest(t)
	ctx := context.Background()

	for i := 1; i < models.MaxRoadmapGoalsPerUser; i++ {
		if _, err := srv.CreateGoal(ctx, testRoadmapUser, "goal-"+string(rune('a'+i))); err != nil {
			t.Fatalf("create goal: %v", err)
		}
	}
	if err := srv.ArchiveGoal(ctx, testRoadmapUser, firstGoal); err != nil {
		t.Fatalf("archive goal: %v", err)
	}
	if _, err := srv.CreateGoal(ctx, testRoadmapUser, "freed-slot"); err != nil {
		t.Fatalf("create after archiving one: %v", err)
	}
	if err := srv.RestoreGoal(ctx, testRoadmapUser, firstGoal); !errors.Is(err, models.ErrRoadmapGoalLimitReached) {
		t.Fatalf("restore over cap: got %v, want ErrRoadmapGoalLimitReached", err)
	}
}

// Deleting a goal must not take its technologies with it — they resurface as
// unattached, which is what ON DELETE SET NULL buys.
func TestRoadmapService_DeletingGoalOrphansItsTechnologies(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()

	if _, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka"); err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if err := srv.DeleteGoalForever(ctx, testRoadmapUser, goalID); err != nil {
		t.Fatalf("delete goal: %v", err)
	}

	orphans, err := srv.ListOrphanRoadmaps(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("list orphans: %v", err)
	}
	if len(orphans) != 1 || orphans[0].Name != "Kafka" {
		t.Fatalf("orphans = %+v, want the one surviving technology", orphans)
	}
	if orphans[0].GoalID != nil {
		t.Errorf("orphan still points at a goal: %v", *orphans[0].GoalID)
	}

	// And it can be adopted by another goal.
	newGoal, err := srv.CreateGoal(ctx, testRoadmapUser, "Выйти на синьора")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if err := srv.AssignRoadmapToGoal(ctx, testRoadmapUser, orphans[0].ID, newGoal); err != nil {
		t.Fatalf("assign orphan to goal: %v", err)
	}
	inGoal, err := srv.ListRoadmaps(ctx, testRoadmapUser, newGoal)
	if err != nil || len(inGoal) != 1 {
		t.Fatalf("list in new goal = (%d, %v), want (1, nil)", len(inGoal), err)
	}
}

// The paste flow is the main input path, so its tag parsing carries weight:
// tags set kind/difficulty and must not survive into the stored text.
func TestRoadmapService_AddCardsParsesTags(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()
	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}

	long := ""
	for len([]rune(long)) <= models.MaxRoadmapCardTextLen {
		long += "x"
	}

	text := "что такое брокер !easy\n" +
		"- Kafka internals #book !hard\n" +
		"https://kafka.apache.org/docs\n" +
		"#lecture запись доклада про партиции\n" +
		"\n   \n" +
		"#book !hard\n" +
		long
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
	got := map[string]models.RoadmapCardItem{}
	for _, c := range cards {
		got[c.Text] = c
	}

	if c, ok := got["что такое брокер"]; !ok {
		t.Error("missing easy topic card (tag should be stripped from the text)")
	} else if c.Difficulty != models.RoadmapCardEasy || c.Kind != models.RoadmapCardTopic {
		t.Errorf("easy topic = (kind %q, difficulty %d), want (topic, easy)", c.Kind, c.Difficulty)
	}

	if c, ok := got["Kafka internals"]; !ok {
		t.Error("missing book card (list marker and both tags should be stripped)")
	} else if c.Kind != models.RoadmapCardBook || c.Difficulty != models.RoadmapCardHard {
		t.Errorf("book card = (kind %q, difficulty %d), want (book, hard)", c.Kind, c.Difficulty)
	}

	if c, ok := got["https://kafka.apache.org/docs"]; !ok {
		t.Error("missing link card")
	} else if c.Kind != models.RoadmapCardArticle {
		t.Errorf("bare link kind = %q, want article (a pasted URL is a resource, not a topic)", c.Kind)
	} else if c.Difficulty != models.RoadmapCardMedium {
		t.Errorf("untagged difficulty = %d, want medium", c.Difficulty)
	}

	if c, ok := got["запись доклада про партиции"]; !ok {
		t.Error("missing lecture card (leading tag should be stripped)")
	} else if c.Kind != models.RoadmapCardLecture {
		t.Errorf("lecture kind = %q, want lecture", c.Kind)
	}

	if _, ok := got[""]; ok {
		t.Error("a tags-only line produced an empty card")
	}
}

func TestRoadmapService_AddCardsRejectsBlankPaste(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()
	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "\n  \n#book\n"); !errors.Is(err, models.ErrRoadmapNoCardsParsed) {
		t.Fatalf("blank paste: got %v, want ErrRoadmapNoCardsParsed", err)
	}
}

// The checklist must read as "what to do next": pending easiest-first, done
// at the bottom.
func TestRoadmapService_CardsListEasiestFirst(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()
	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "hard one !hard\neasy one !easy\nmid one"); err != nil {
		t.Fatalf("add cards: %v", err)
	}

	cards, err := srv.ListCards(ctx, testRoadmapUser, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	want := []string{"easy one", "mid one", "hard one"}
	for i, w := range want {
		if cards[i].Text != w {
			t.Fatalf("card order = %q..., want %q at index %d", cards[i].Text, w, i)
		}
	}

	// Ticking the easiest moves it to the bottom.
	if _, err := srv.ToggleCardDone(ctx, testRoadmapUser, cards[0].ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}
	cards, err = srv.ListCards(ctx, testRoadmapUser, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if cards[len(cards)-1].Text != "easy one" || !cards[len(cards)-1].IsDone {
		t.Errorf("done card = %q (done %v), want it last and done", cards[len(cards)-1].Text, cards[len(cards)-1].IsDone)
	}
}

func TestRoadmapService_CycleCardDifficulty(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()
	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "one !easy"); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	cards, err := srv.ListCards(ctx, testRoadmapUser, id)
	if err != nil || len(cards) != 1 {
		t.Fatalf("list cards = (%d, %v), want (1, nil)", len(cards), err)
	}

	for _, want := range []int{models.RoadmapCardMedium, models.RoadmapCardHard, models.RoadmapCardEasy} {
		gotRoadmapID, err := srv.CycleCardDifficulty(ctx, testRoadmapUser, cards[0].ID)
		if err != nil {
			t.Fatalf("cycle difficulty: %v", err)
		}
		if gotRoadmapID != id {
			t.Errorf("cycle returned roadmap %d, want %d", gotRoadmapID, id)
		}
		updated, err := srv.ListCards(ctx, testRoadmapUser, id)
		if err != nil {
			t.Fatalf("list cards: %v", err)
		}
		if updated[0].Difficulty != want {
			t.Fatalf("difficulty = %d, want %d", updated[0].Difficulty, want)
		}
	}
}

// The digest exists to propose the next realistic step, so it must lead with
// the easiest pending cards and still leave room for every technology.
func TestRoadmapService_DigestOffersEasiestFirst(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()

	big, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	small, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Docker")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, big, "a !hard\nb !hard\nc !hard\nd !easy\ne !hard\nf !hard"); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, small, "x !easy\ny !mid"); err != nil {
		t.Fatalf("add cards: %v", err)
	}

	digest, err := srv.PickDigestCards(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("pick digest: %v", err)
	}
	if len(digest) == 0 {
		t.Fatal("digest is empty")
	}
	if digest[0].Difficulty != models.RoadmapCardEasy {
		t.Errorf("digest leads with difficulty %d, want the easiest (%d)", digest[0].Difficulty, models.RoadmapCardEasy)
	}
	for i := 1; i < len(digest); i++ {
		if digest[i].Difficulty < digest[i-1].Difficulty {
			t.Fatalf("digest is not easiest-first: %d after %d", digest[i].Difficulty, digest[i-1].Difficulty)
		}
	}

	perRoadmap := map[int64]int{}
	for _, c := range digest {
		perRoadmap[c.RoadmapID]++
	}
	if perRoadmap[big] > models.RoadmapDigestPerRoadmapCap {
		t.Errorf("Kafka contributed %d cards, want at most %d", perRoadmap[big], models.RoadmapDigestPerRoadmapCap)
	}
	if perRoadmap[small] == 0 {
		t.Error("Docker contributed nothing — the per-technology cap should leave room for it")
	}
}

// Archiving the goal must take its whole plan out of the reminder rotation,
// not just hide it on screen.
func TestRoadmapService_ArchivedGoalDropsOutOfDigest(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()

	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, _, err := srv.AddCardsFromText(ctx, testRoadmapUser, id, "brokers\npartitions"); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if digest, err := srv.PickDigestCards(ctx, testRoadmapUser); err != nil || len(digest) == 0 {
		t.Fatalf("digest before archiving = (%d cards, %v), want some", len(digest), err)
	}

	if err := srv.ArchiveGoal(ctx, testRoadmapUser, goalID); err != nil {
		t.Fatalf("archive goal: %v", err)
	}
	digest, err := srv.PickDigestCards(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("pick digest: %v", err)
	}
	if len(digest) != 0 {
		t.Errorf("digest still has %d cards from an archived goal", len(digest))
	}
}

// Re-tapping the same interval must not push the next digest further away.
func TestRoadmapService_ActivatePreservesNextPush(t *testing.T) {
	srv, fake, _ := newRoadmapServiceForTest(t)
	ctx := context.Background()

	if err := srv.Activate(ctx, testRoadmapUser, 180); err != nil {
		t.Fatalf("activate: %v", err)
	}
	_, first, _, err := fake.GetPushSettings(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}

	if err := srv.Activate(ctx, testRoadmapUser, 180); err != nil {
		t.Fatalf("re-activate: %v", err)
	}
	_, second, _, err := fake.GetPushSettings(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}
	if !second.Equal(first) {
		t.Errorf("same interval moved next_push_at %v -> %v, want unchanged", first, second)
	}

	if err := srv.Activate(ctx, testRoadmapUser, 60); err != nil {
		t.Fatalf("activate new interval: %v", err)
	}
	_, third, _, err := fake.GetPushSettings(ctx, testRoadmapUser)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}
	if third.Equal(first) {
		t.Error("changing the interval left next_push_at untouched, want rescheduled")
	}
}

func TestRoadmapService_ActivateRejectsBadInterval(t *testing.T) {
	srv, _, _ := newRoadmapServiceForTest(t)
	ctx := context.Background()
	for _, min := range []int{0, -5, 1441} {
		if err := srv.Activate(ctx, testRoadmapUser, min); !errors.Is(err, models.ErrRoadmapInvalidInterval) {
			t.Errorf("activate %d min: got %v, want ErrRoadmapInvalidInterval", min, err)
		}
	}
}

func TestRoadmapService_SetMasteryCriteria(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()
	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}

	if err := srv.SetMasteryCriteria(ctx, testRoadmapUser, id, " поднять\nкластер "); err != nil {
		t.Fatalf("set criteria: %v", err)
	}
	item, err := srv.Roadmap(ctx, testRoadmapUser, id)
	if err != nil {
		t.Fatalf("get technology: %v", err)
	}
	if item.MasteryCriteria != "поднять кластер" {
		t.Errorf("criteria = %q, want %q", item.MasteryCriteria, "поднять кластер")
	}

	long := ""
	for len([]rune(long)) <= models.MaxRoadmapCriteriaLen {
		long += "y"
	}
	if err := srv.SetMasteryCriteria(ctx, testRoadmapUser, id, long); !errors.Is(err, models.ErrRoadmapCriteriaTooLong) {
		t.Fatalf("over-long criteria: got %v, want ErrRoadmapCriteriaTooLong", err)
	}
}

func TestRoadmapService_StatsDetail(t *testing.T) {
	srv, _, goalID := newRoadmapServiceForTest(t)
	ctx := context.Background()
	id, err := srv.CreateRoadmap(ctx, testRoadmapUser, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
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
	if detail.Overall.TotalGoals != 1 || detail.Overall.TotalRoadmaps != 1 {
		t.Errorf("overall = %+v, want 1 goal / 1 technology", detail.Overall)
	}
	if detail.Overall.TotalCards != 3 || detail.Overall.DoneCards != 1 || detail.Overall.PendingCards != 2 {
		t.Errorf("overall cards = %+v, want 3/1/2", detail.Overall)
	}
	if len(detail.Goals) != 1 || detail.Goals[0].TotalCards != 3 || detail.Goals[0].DoneCards != 1 {
		t.Errorf("goal rollup = %+v, want the goal aggregating its technologies' cards", detail.Goals)
	}
	if len(detail.Roadmaps) != 1 || detail.Roadmaps[0].GoalName != "Выйти на мидла" {
		t.Errorf("per-technology rows = %+v, want one row naming its goal", detail.Roadmaps)
	}
}
