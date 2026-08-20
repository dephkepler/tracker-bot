//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

func cardsFrom(specs ...errlocal.RoadmapCardItem) []errlocal.RoadmapCardItem {
	return specs
}

func topic(text string, difficulty int) errlocal.RoadmapCardItem {
	return errlocal.RoadmapCardItem{Text: text, Kind: errlocal.RoadmapCardTopic, Difficulty: difficulty}
}

// TestRoadmapRepository_GoalLifecycle exercises create -> list -> rename ->
// archive -> restore -> delete against real Postgres, including the
// unique-name constraint and the rollup counts.
func TestRoadmapRepository_GoalLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if _, err := repo.CreateGoal(ctx, userID, "Mid-level"); !errors.Is(err, errlocal.ErrRoadmapGoalExists) {
		t.Fatalf("duplicate goal: got %v, want ErrRoadmapGoalExists", err)
	}

	if n, err := repo.CountGoals(ctx, userID); err != nil || n != 1 {
		t.Fatalf("count goals = (%d, %v), want (1, nil)", n, err)
	}

	if err := repo.RenameGoal(ctx, userID, goalID, "Senior"); err != nil {
		t.Fatalf("rename goal: %v", err)
	}
	goal, err := repo.GetGoal(ctx, userID, goalID)
	if err != nil || goal.Name != "Senior" {
		t.Fatalf("get goal = (%q, %v), want (Senior, nil)", goal.Name, err)
	}

	// Rollups: two technologies, three cards, one done.
	kafka, err := repo.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, err := repo.CreateRoadmap(ctx, userID, goalID, "Docker"); err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, err := repo.AddCards(ctx, kafka, cardsFrom(topic("brokers", 1), topic("partitions", 2), topic("acks", 3))); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	cards, err := repo.ListCards(ctx, userID, kafka)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if _, err := repo.ToggleCardDone(ctx, userID, cards[0].ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	goal, err = repo.GetGoal(ctx, userID, goalID)
	if err != nil {
		t.Fatalf("get goal: %v", err)
	}
	if goal.TotalRoadmaps != 2 {
		t.Errorf("goal.TotalRoadmaps = %d, want 2 (the double join must not multiply this)", goal.TotalRoadmaps)
	}
	if goal.TotalCards != 3 || goal.DoneCards != 1 {
		t.Errorf("goal rollup = %d/%d cards, want 3/1", goal.DoneCards, goal.TotalCards)
	}

	if err := repo.ArchiveGoal(ctx, userID, goalID); err != nil {
		t.Fatalf("archive goal: %v", err)
	}
	if n, err := repo.CountGoals(ctx, userID); err != nil || n != 0 {
		t.Fatalf("count after archive = (%d, %v), want (0, nil)", n, err)
	}
	archived, err := repo.ListGoals(ctx, userID, true)
	if err != nil || len(archived) != 1 {
		t.Fatalf("list archived goals = (%d, %v), want (1, nil)", len(archived), err)
	}
	if err := repo.RestoreGoal(ctx, userID, goalID); err != nil {
		t.Fatalf("restore goal: %v", err)
	}

	otherUser := testUser(t, pool)
	if _, err := repo.GetGoal(ctx, otherUser, goalID); !errors.Is(err, errlocal.ErrRoadmapGoalNotFound) {
		t.Fatalf("cross-user get goal: got %v, want ErrRoadmapGoalNotFound", err)
	}
	if err := repo.RenameGoal(ctx, otherUser, goalID, "hijack"); !errors.Is(err, errlocal.ErrRoadmapGoalNotFound) {
		t.Fatalf("cross-user rename: got %v, want ErrRoadmapGoalNotFound", err)
	}
}

// TestRoadmapRepository_GoalDeleteOrphansTechnologies pins the ON DELETE SET
// NULL behavior: dropping a goal must not take a pile of technologies and
// their cards with it.
func TestRoadmapRepository_GoalDeleteOrphansTechnologies(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	kafka, err := repo.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if _, err := repo.AddCards(ctx, kafka, cardsFrom(topic("brokers", 2))); err != nil {
		t.Fatalf("add cards: %v", err)
	}

	if err := repo.DeleteGoalForever(ctx, userID, goalID); err != nil {
		t.Fatalf("delete goal: %v", err)
	}

	item, err := repo.GetRoadmap(ctx, userID, kafka)
	if err != nil {
		t.Fatalf("technology should survive its goal: %v", err)
	}
	if item.GoalID != nil {
		t.Errorf("technology still points at the deleted goal: %v", *item.GoalID)
	}
	if item.TotalCards != 1 {
		t.Errorf("technology lost its cards: TotalCards = %d, want 1", item.TotalCards)
	}

	orphans, err := repo.ListRoadmaps(ctx, userID, nil, false)
	if err != nil || len(orphans) != 1 {
		t.Fatalf("list orphans = (%d, %v), want (1, nil)", len(orphans), err)
	}

	// Adoption by a new goal.
	newGoal, err := repo.CreateGoal(ctx, userID, "Senior")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if err := repo.AssignRoadmapToGoal(ctx, userID, kafka, newGoal); err != nil {
		t.Fatalf("assign to goal: %v", err)
	}
	if n, err := repo.CountRoadmapsInGoal(ctx, userID, newGoal); err != nil || n != 1 {
		t.Fatalf("count in new goal = (%d, %v), want (1, nil)", n, err)
	}

	// A goal belonging to somebody else is not a valid target.
	otherUser := testUser(t, pool)
	foreignGoal, err := repo.CreateGoal(ctx, otherUser, "Their goal")
	if err != nil {
		t.Fatalf("create foreign goal: %v", err)
	}
	if err := repo.AssignRoadmapToGoal(ctx, userID, kafka, foreignGoal); !errors.Is(err, errlocal.ErrRoadmapNotFound) {
		t.Fatalf("assign into a foreign goal: got %v, want ErrRoadmapNotFound", err)
	}
}

// TestRoadmapRepository_CreateRoadmapRequiresOwnGoal guards the WHERE EXISTS
// in CreateRoadmap — the goal id comes from a callback payload.
func TestRoadmapRepository_CreateRoadmapRequiresOwnGoal(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)
	otherUser := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if _, err := repo.CreateRoadmap(ctx, otherUser, goalID, "Kafka"); !errors.Is(err, errlocal.ErrRoadmapGoalNotFound) {
		t.Fatalf("create in a foreign goal: got %v, want ErrRoadmapGoalNotFound", err)
	}
	if _, err := repo.CreateRoadmap(ctx, userID, goalID+9999, "Kafka"); !errors.Is(err, errlocal.ErrRoadmapGoalNotFound) {
		t.Fatalf("create in a missing goal: got %v, want ErrRoadmapGoalNotFound", err)
	}
}

// TestRoadmapRepository_Cards covers bulk add with kind/difficulty, the
// easiest-first ordering, the done/undone flip with its done_at bookkeeping,
// the difficulty cycle, delete, and ownership scoping.
func TestRoadmapRepository_Cards(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	id, err := repo.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}

	added, err := repo.AddCards(ctx, id, cardsFrom(
		errlocal.RoadmapCardItem{Text: "internals", Kind: errlocal.RoadmapCardBook, Difficulty: errlocal.RoadmapCardHard},
		errlocal.RoadmapCardItem{Text: "brokers", Kind: errlocal.RoadmapCardTopic, Difficulty: errlocal.RoadmapCardEasy},
		errlocal.RoadmapCardItem{Text: "partitions talk", Kind: errlocal.RoadmapCardLecture, Difficulty: errlocal.RoadmapCardMedium},
	))
	if err != nil || added != 3 {
		t.Fatalf("add cards = (%d, %v), want (3, nil)", added, err)
	}

	cards, err := repo.ListCards(ctx, userID, id)
	if err != nil || len(cards) != 3 {
		t.Fatalf("list cards = (%d, %v), want (3, nil)", len(cards), err)
	}
	// Easiest-first, regardless of insertion order.
	wantOrder := []string{"brokers", "partitions talk", "internals"}
	for i, want := range wantOrder {
		if cards[i].Text != want {
			t.Fatalf("card %d = %q, want %q (list must be easiest-first)", i, cards[i].Text, want)
		}
	}
	if cards[0].Kind != errlocal.RoadmapCardTopic || cards[2].Kind != errlocal.RoadmapCardBook {
		t.Errorf("kinds did not round-trip: %q / %q", cards[0].Kind, cards[2].Kind)
	}
	for _, c := range cards {
		if c.IsDone || c.DoneAt != nil {
			t.Errorf("card %q: done=%v done_at=%v, want pending with nil done_at", c.Text, c.IsDone, c.DoneAt)
		}
	}

	gotRoadmapID, err := repo.ToggleCardDone(ctx, userID, cards[0].ID)
	if err != nil {
		t.Fatalf("toggle card done: %v", err)
	}
	if gotRoadmapID != id {
		t.Errorf("toggle returned roadmap %d, want %d", gotRoadmapID, id)
	}
	cards, err = repo.ListCards(ctx, userID, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	last := cards[len(cards)-1]
	if last.Text != "brokers" || !last.IsDone || last.DoneAt == nil {
		t.Errorf("done card = %q (done %v, at %v), want brokers last, done, stamped", last.Text, last.IsDone, last.DoneAt)
	}

	if total, done, err := repo.CountCards(ctx, userID); err != nil || total != 3 || done != 1 {
		t.Fatalf("count cards = (%d, %d, %v), want (3, 1, nil)", total, done, err)
	}

	// Flipping back clears done_at rather than leaving a stale timestamp.
	if _, err := repo.ToggleCardDone(ctx, userID, last.ID); err != nil {
		t.Fatalf("toggle back: %v", err)
	}
	cards, err = repo.ListCards(ctx, userID, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range cards {
		if c.ID == last.ID && (c.IsDone || c.DoneAt != nil) {
			t.Errorf("re-opened card: done=%v done_at=%v, want pending with nil done_at", c.IsDone, c.DoneAt)
		}
	}

	// Difficulty cycles 1 -> 2 -> 3 -> 1 and stays inside the CHECK.
	target := cards[0].ID
	for _, want := range []int{errlocal.RoadmapCardMedium, errlocal.RoadmapCardHard, errlocal.RoadmapCardEasy} {
		if _, err := repo.CycleCardDifficulty(ctx, userID, target); err != nil {
			t.Fatalf("cycle difficulty: %v", err)
		}
		updated, err := repo.ListCards(ctx, userID, id)
		if err != nil {
			t.Fatalf("list cards: %v", err)
		}
		for _, c := range updated {
			if c.ID == target && c.Difficulty != want {
				t.Fatalf("difficulty = %d, want %d", c.Difficulty, want)
			}
		}
	}

	otherUser := testUser(t, pool)
	if _, err := repo.ToggleCardDone(ctx, otherUser, target); !errors.Is(err, errlocal.ErrRoadmapCardNotFound) {
		t.Fatalf("cross-user toggle: got %v, want ErrRoadmapCardNotFound", err)
	}
	if _, err := repo.CycleCardDifficulty(ctx, otherUser, target); !errors.Is(err, errlocal.ErrRoadmapCardNotFound) {
		t.Fatalf("cross-user cycle: got %v, want ErrRoadmapCardNotFound", err)
	}
	if err := repo.DeleteCard(ctx, otherUser, target); !errors.Is(err, errlocal.ErrRoadmapCardNotFound) {
		t.Fatalf("cross-user delete: got %v, want ErrRoadmapCardNotFound", err)
	}

	if err := repo.DeleteCard(ctx, userID, target); err != nil {
		t.Fatalf("delete card: %v", err)
	}
	if err := repo.DeleteRoadmapForever(ctx, userID, id); err != nil {
		t.Fatalf("delete technology: %v", err)
	}
	if total, _, err := repo.CountCards(ctx, userID); err != nil || total != 0 {
		t.Fatalf("count after cascade = (%d, %v), want (0, nil)", total, err)
	}
}

// TestRoadmapRepository_PickDigestCards is the window function's contract:
// easiest pending cards first, capped per technology and overall, with done
// cards, deactivated technologies and archived goals excluded.
func TestRoadmapRepository_PickDigestCards(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	archivedGoal, err := repo.CreateGoal(ctx, userID, "Parked")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}

	big, err := repo.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	small, err := repo.CreateRoadmap(ctx, userID, goalID, "Docker")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	off, err := repo.CreateRoadmap(ctx, userID, goalID, "Rust")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	parked, err := repo.CreateRoadmap(ctx, userID, archivedGoal, "Elixir")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}

	// The hard cards are inserted first on purpose: without the difficulty
	// sort they would fill the digest by themselves.
	if _, err := repo.AddCards(ctx, big, cardsFrom(
		topic("hard-1", 3), topic("hard-2", 3), topic("hard-3", 3),
		topic("easy-1", 1), topic("mid-1", 2),
	)); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, err := repo.AddCards(ctx, small, cardsFrom(topic("d-easy", 1), topic("d-mid", 2))); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, err := repo.AddCards(ctx, off, cardsFrom(topic("never", 1))); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, err := repo.AddCards(ctx, parked, cardsFrom(topic("parked", 1))); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if err := repo.ToggleRoadmapActive(ctx, userID, off); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := repo.ArchiveGoal(ctx, userID, archivedGoal); err != nil {
		t.Fatalf("archive goal: %v", err)
	}

	digest, err := repo.PickDigestCards(ctx, userID, 3, 8)
	if err != nil {
		t.Fatalf("pick digest: %v", err)
	}
	if len(digest) == 0 {
		t.Fatal("digest is empty")
	}
	if digest[0].Difficulty != errlocal.RoadmapCardEasy {
		t.Errorf("digest leads with difficulty %d, want the easiest", digest[0].Difficulty)
	}
	for i := 1; i < len(digest); i++ {
		if digest[i].Difficulty < digest[i-1].Difficulty {
			t.Fatalf("digest not easiest-first: %d after %d", digest[i].Difficulty, digest[i-1].Difficulty)
		}
	}
	perRoadmap := map[int64]int{}
	for _, c := range digest {
		perRoadmap[c.RoadmapID]++
		if c.RoadmapID == off {
			t.Errorf("digest includes %q from a deactivated technology", c.Text)
		}
		if c.RoadmapID == parked {
			t.Errorf("digest includes %q from an archived goal", c.Text)
		}
		if c.RoadmapName == "" {
			t.Errorf("digest card %q has no technology name", c.Text)
		}
	}
	if perRoadmap[big] != 3 {
		t.Errorf("Kafka contributed %d cards, want exactly the cap of 3", perRoadmap[big])
	}
	if perRoadmap[small] != 2 {
		t.Errorf("Docker contributed %d cards, want 2", perRoadmap[small])
	}

	// The total cap keeps the easiest cards.
	capped, err := repo.PickDigestCards(ctx, userID, 3, 2)
	if err != nil {
		t.Fatalf("pick digest (capped): %v", err)
	}
	if len(capped) != 2 {
		t.Fatalf("digest with totalCap=2 has %d cards, want 2", len(capped))
	}
	for _, c := range capped {
		if c.Difficulty != errlocal.RoadmapCardEasy {
			t.Errorf("capped digest kept difficulty %d, want only the easiest", c.Difficulty)
		}
	}

	// Completing a technology takes it out entirely.
	cards, err := repo.ListCards(ctx, userID, small)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range cards {
		if _, err := repo.ToggleCardDone(ctx, userID, c.ID); err != nil {
			t.Fatalf("toggle: %v", err)
		}
	}
	digest, err = repo.PickDigestCards(ctx, userID, 3, 8)
	if err != nil {
		t.Fatalf("pick digest: %v", err)
	}
	for _, c := range digest {
		if c.RoadmapID == small {
			t.Errorf("digest still includes %q from a finished technology", c.Text)
		}
	}
}

// TestRoadmapRepository_CardConstraints pins the CHECK constraints, so a bad
// value fails loudly here rather than silently widening later.
func TestRoadmapRepository_CardConstraints(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	id, err := repo.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}

	if _, err := repo.AddCards(ctx, id, cardsFrom(topic("out of range", 4))); err == nil {
		t.Error("difficulty 4 was accepted, want the CHECK to reject it")
	}
	if _, err := repo.AddCards(ctx, id, cardsFrom(errlocal.RoadmapCardItem{Text: "bad kind", Kind: "video", Difficulty: 2})); err == nil {
		t.Error("kind \"video\" was accepted, want the CHECK to reject it")
	}
}

// TestRoadmapRepository_PushSchedule mirrors the Learning push-scheduling
// test: upsert, read back, advance, disable, and the due-user query.
func TestRoadmapRepository_PushSchedule(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	if _, _, enabled, err := repo.GetPushSettings(ctx, userID); err != nil || enabled {
		t.Fatalf("push settings before upsert = (enabled %v, %v), want (false, nil)", enabled, err)
	}

	now := time.Now().UTC()
	if err := repo.UpsertPushInterval(ctx, userID, 180, now.Add(-time.Minute)); err != nil {
		t.Fatalf("upsert push interval: %v", err)
	}
	intervalMin, _, enabled, err := repo.GetPushSettings(ctx, userID)
	if err != nil || intervalMin != 180 || !enabled {
		t.Fatalf("push settings = (%d, %v, %v), want (180, true, nil)", intervalMin, enabled, err)
	}

	due, err := repo.ListDueUsers(ctx, now, 10)
	if err != nil {
		t.Fatalf("list due users: %v", err)
	}
	found := false
	for _, u := range due {
		if u.DBUserID == userID {
			found = true
			if u.IntervalMin != 180 {
				t.Errorf("due user interval = %d, want 180", u.IntervalMin)
			}
		}
	}
	if !found {
		t.Fatal("overdue user missing from ListDueUsers")
	}

	if err := repo.SetNextPush(ctx, userID, now.Add(time.Hour)); err != nil {
		t.Fatalf("set next push: %v", err)
	}
	due, err = repo.ListDueUsers(ctx, now, 10)
	if err != nil {
		t.Fatalf("list due users: %v", err)
	}
	for _, u := range due {
		if u.DBUserID == userID {
			t.Fatal("user still due after next_push_at moved into the future")
		}
	}

	if err := repo.DisablePush(ctx, userID); err != nil {
		t.Fatalf("disable push: %v", err)
	}
	if _, _, enabled, err := repo.GetPushSettings(ctx, userID); err != nil || enabled {
		t.Fatalf("push settings after disable = (enabled %v, %v), want (false, nil)", enabled, err)
	}

	if err := repo.UpsertPushInterval(ctx, userID, 0, now); !errors.Is(err, errlocal.ErrRoadmapInvalidInterval) {
		t.Fatalf("upsert 0 min: got %v, want ErrRoadmapInvalidInterval", err)
	}
	if err := repo.UpsertPushInterval(ctx, userID, 1441, now); !errors.Is(err, errlocal.ErrRoadmapInvalidInterval) {
		t.Fatalf("upsert 1441 min: got %v, want ErrRoadmapInvalidInterval", err)
	}
}

// TestRoadmapRepository_CardStats checks the per-technology rows carry their
// goal's name, keep unattached technologies, and exclude archived ones.
func TestRoadmapRepository_CardStats(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	goalID, err := repo.CreateGoal(ctx, userID, "Mid-level")
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	kafka, err := repo.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if err := repo.SetMasteryCriteria(ctx, userID, kafka, "run a cluster"); err != nil {
		t.Fatalf("set criteria: %v", err)
	}
	if _, err := repo.AddCards(ctx, kafka, cardsFrom(topic("a", 1), topic("b", 2))); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	cards, err := repo.ListCards(ctx, userID, kafka)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if _, err := repo.ToggleCardDone(ctx, userID, cards[0].ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	gone, err := repo.CreateRoadmap(ctx, userID, goalID, "Elixir")
	if err != nil {
		t.Fatalf("create technology: %v", err)
	}
	if err := repo.ArchiveRoadmap(ctx, userID, gone); err != nil {
		t.Fatalf("archive technology: %v", err)
	}

	stats, err := repo.GetRoadmapCardStats(ctx, userID)
	if err != nil {
		t.Fatalf("get card stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("stats rows = %d, want 1 (archived excluded)", len(stats))
	}
	if stats[0].GoalName != "Mid-level" || stats[0].Name != "Kafka" {
		t.Errorf("stats row = %+v, want Kafka under Mid-level", stats[0])
	}
	if stats[0].MasteryCriteria != "run a cluster" || stats[0].TotalCards != 2 || stats[0].DoneCards != 1 {
		t.Errorf("stats row = %+v, want criteria + 2/1 cards", stats[0])
	}

	// An unattached technology still shows up, with an empty goal name.
	if err := repo.DeleteGoalForever(ctx, userID, goalID); err != nil {
		t.Fatalf("delete goal: %v", err)
	}
	stats, err = repo.GetRoadmapCardStats(ctx, userID)
	if err != nil {
		t.Fatalf("get card stats: %v", err)
	}
	if len(stats) != 1 || stats[0].GoalName != "" {
		t.Errorf("stats after goal delete = %+v, want the orphan with an empty goal name", stats)
	}
}
