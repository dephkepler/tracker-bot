//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

// TestRoadmapRepository_RoadmapLifecycle exercises create -> list -> goal ->
// rename -> toggle -> archive -> restore -> delete forever against real
// Postgres, including the unique-name constraint.
func TestRoadmapRepository_RoadmapLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	id, err := repo.CreateRoadmap(ctx, userID, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	if _, err := repo.CreateRoadmap(ctx, userID, "Go"); !errors.Is(err, errlocal.ErrRoadmapExists) {
		t.Fatalf("create duplicate: got %v, want ErrRoadmapExists", err)
	}

	second, err := repo.CreateRoadmap(ctx, userID, "Kafka")
	if err != nil {
		t.Fatalf("create second roadmap: %v", err)
	}

	if n, err := repo.CountRoadmaps(ctx, userID); err != nil || n != 2 {
		t.Fatalf("count roadmaps = (%d, %v), want (2, nil)", n, err)
	}

	if err := repo.SetGoal(ctx, userID, id, "ship a service in production"); err != nil {
		t.Fatalf("set goal: %v", err)
	}
	if err := repo.RenameRoadmap(ctx, userID, id, "Golang"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	item, err := repo.GetRoadmap(ctx, userID, id)
	if err != nil {
		t.Fatalf("get roadmap: %v", err)
	}
	if item.Name != "Golang" || item.Goal != "ship a service in production" {
		t.Fatalf("get roadmap = (%q, %q), want (Golang, ship a service in production)", item.Name, item.Goal)
	}
	if !item.Active || item.IsArchived {
		t.Fatalf("new roadmap = (active %v, archived %v), want (true, false)", item.Active, item.IsArchived)
	}

	if err := repo.RenameRoadmap(ctx, userID, second, "Golang"); !errors.Is(err, errlocal.ErrRoadmapExists) {
		t.Fatalf("rename to existing name: got %v, want ErrRoadmapExists", err)
	}

	if err := repo.ToggleRoadmapActive(ctx, userID, id); err != nil {
		t.Fatalf("toggle active: %v", err)
	}
	if item, err := repo.GetRoadmap(ctx, userID, id); err != nil || item.Active {
		t.Fatalf("after toggle: active = %v (err %v), want false", item.Active, err)
	}

	if err := repo.ArchiveRoadmap(ctx, userID, second); err != nil {
		t.Fatalf("archive: %v", err)
	}
	// Archiving must free a slot in the MaxRoadmapsPerUser cap, which is
	// exactly what CountRoadmaps' is_archived = FALSE filter is for.
	if n, err := repo.CountRoadmaps(ctx, userID); err != nil || n != 1 {
		t.Fatalf("count after archive = (%d, %v), want (1, nil)", n, err)
	}
	archived, err := repo.ListRoadmaps(ctx, userID, true)
	if err != nil || len(archived) != 1 || archived[0].ID != second {
		t.Fatalf("list archived = (%d items, %v), want the one archived roadmap", len(archived), err)
	}
	if err := repo.RestoreRoadmap(ctx, userID, second); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if n, err := repo.CountRoadmaps(ctx, userID); err != nil || n != 2 {
		t.Fatalf("count after restore = (%d, %v), want (2, nil)", n, err)
	}

	if err := repo.DeleteRoadmapForever(ctx, userID, second); err != nil {
		t.Fatalf("delete forever: %v", err)
	}
	if _, err := repo.GetRoadmap(ctx, userID, second); !errors.Is(err, errlocal.ErrRoadmapNotFound) {
		t.Fatalf("get deleted roadmap: got %v, want ErrRoadmapNotFound", err)
	}

	// Another user must not see or touch this one's roadmaps.
	otherUser := testUser(t, pool)
	if _, err := repo.GetRoadmap(ctx, otherUser, id); !errors.Is(err, errlocal.ErrRoadmapNotFound) {
		t.Fatalf("cross-user get: got %v, want ErrRoadmapNotFound", err)
	}
	if err := repo.ArchiveRoadmap(ctx, otherUser, id); !errors.Is(err, errlocal.ErrRoadmapNotFound) {
		t.Fatalf("cross-user archive: got %v, want ErrRoadmapNotFound", err)
	}
}

// TestRoadmapRepository_Cards covers bulk add, listing order (pending
// before done), the done/undone flip with its done_at bookkeeping, delete,
// and ownership scoping.
func TestRoadmapRepository_Cards(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	id, err := repo.CreateRoadmap(ctx, userID, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	added, err := repo.AddCards(ctx, id, []string{"goroutines", "channels", "select"})
	if err != nil || added != 3 {
		t.Fatalf("add cards = (%d, %v), want (3, nil)", added, err)
	}

	cards, err := repo.ListCards(ctx, userID, id)
	if err != nil || len(cards) != 3 {
		t.Fatalf("list cards = (%d, %v), want (3, nil)", len(cards), err)
	}
	if cards[0].Text != "goroutines" {
		t.Errorf("first card = %q, want goroutines (insertion order)", cards[0].Text)
	}
	for _, c := range cards {
		if c.IsDone || c.DoneAt != nil {
			t.Errorf("card %q: done=%v done_at=%v, want pending with nil done_at", c.Text, c.IsDone, c.DoneAt)
		}
	}

	// Tick the first card: it must report its roadmap, stamp done_at, and
	// sort behind the still-pending ones.
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
	if cards[len(cards)-1].Text != "goroutines" || !cards[len(cards)-1].IsDone {
		t.Errorf("done card = %q (done %v), want it last and done", cards[len(cards)-1].Text, cards[len(cards)-1].IsDone)
	}
	if cards[len(cards)-1].DoneAt == nil {
		t.Error("done card has nil done_at, want it stamped")
	}

	if total, done, err := repo.CountCards(ctx, userID); err != nil || total != 3 || done != 1 {
		t.Fatalf("count cards = (%d, %d, %v), want (3, 1, nil)", total, done, err)
	}

	// Flipping back must clear done_at rather than leave a stale timestamp.
	doneID := cards[len(cards)-1].ID
	if _, err := repo.ToggleCardDone(ctx, userID, doneID); err != nil {
		t.Fatalf("toggle card back: %v", err)
	}
	cards, err = repo.ListCards(ctx, userID, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range cards {
		if c.ID == doneID && (c.IsDone || c.DoneAt != nil) {
			t.Errorf("re-opened card: done=%v done_at=%v, want pending with nil done_at", c.IsDone, c.DoneAt)
		}
	}

	otherUser := testUser(t, pool)
	if _, err := repo.ToggleCardDone(ctx, otherUser, doneID); !errors.Is(err, errlocal.ErrRoadmapCardNotFound) {
		t.Fatalf("cross-user toggle: got %v, want ErrRoadmapCardNotFound", err)
	}
	if err := repo.DeleteCard(ctx, otherUser, doneID); !errors.Is(err, errlocal.ErrRoadmapCardNotFound) {
		t.Fatalf("cross-user delete: got %v, want ErrRoadmapCardNotFound", err)
	}

	if err := repo.DeleteCard(ctx, userID, doneID); err != nil {
		t.Fatalf("delete card: %v", err)
	}
	if cards, err := repo.ListCards(ctx, userID, id); err != nil || len(cards) != 2 {
		t.Fatalf("list after delete = (%d, %v), want (2, nil)", len(cards), err)
	}

	// Deleting the roadmap cascades to its cards.
	if err := repo.DeleteRoadmapForever(ctx, userID, id); err != nil {
		t.Fatalf("delete roadmap forever: %v", err)
	}
	if total, _, err := repo.CountCards(ctx, userID); err != nil || total != 0 {
		t.Fatalf("count cards after cascade = (%d, %v), want (0, nil)", total, err)
	}
}

// TestRoadmapRepository_PickDigestCards is the window-function query's
// contract: oldest pending cards first, at most perRoadmapCap from any one
// roadmap, totalCap overall, and nothing from inactive/archived roadmaps or
// done cards.
func TestRoadmapRepository_PickDigestCards(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	big, err := repo.CreateRoadmap(ctx, userID, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	small, err := repo.CreateRoadmap(ctx, userID, "Kafka")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	off, err := repo.CreateRoadmap(ctx, userID, "Rust")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	archivedRoadmap, err := repo.CreateRoadmap(ctx, userID, "Elixir")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	if _, err := repo.AddCards(ctx, big, []string{"a", "b", "c", "d", "e"}); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, err := repo.AddCards(ctx, small, []string{"x", "y"}); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, err := repo.AddCards(ctx, off, []string{"never"}); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if _, err := repo.AddCards(ctx, archivedRoadmap, []string{"gone"}); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	if err := repo.ToggleRoadmapActive(ctx, userID, off); err != nil {
		t.Fatalf("deactivate roadmap: %v", err)
	}
	if err := repo.ArchiveRoadmap(ctx, userID, archivedRoadmap); err != nil {
		t.Fatalf("archive roadmap: %v", err)
	}

	digest, err := repo.PickDigestCards(ctx, userID, 3, 8)
	if err != nil {
		t.Fatalf("pick digest cards: %v", err)
	}
	perRoadmap := map[int64]int{}
	for _, c := range digest {
		perRoadmap[c.RoadmapID]++
		if c.RoadmapID == off || c.RoadmapID == archivedRoadmap {
			t.Errorf("digest includes card %q from an inactive/archived roadmap", c.Text)
		}
		if c.RoadmapName == "" {
			t.Errorf("digest card %q has empty roadmap name", c.Text)
		}
	}
	if perRoadmap[big] != 3 {
		t.Errorf("big roadmap contributed %d cards, want exactly the per-roadmap cap of 3", perRoadmap[big])
	}
	if perRoadmap[small] != 2 {
		t.Errorf("small roadmap contributed %d cards, want 2", perRoadmap[small])
	}

	// The total cap truncates, keeping the oldest cards.
	capped, err := repo.PickDigestCards(ctx, userID, 3, 2)
	if err != nil {
		t.Fatalf("pick digest cards (capped): %v", err)
	}
	if len(capped) != 2 {
		t.Fatalf("digest with totalCap=2 has %d cards, want 2", len(capped))
	}
	if capped[0].Text != "a" || capped[1].Text != "b" {
		t.Errorf("capped digest = %q, %q, want the two oldest (a, b)", capped[0].Text, capped[1].Text)
	}

	// Done cards drop out.
	cards, err := repo.ListCards(ctx, userID, small)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range cards {
		if _, err := repo.ToggleCardDone(ctx, userID, c.ID); err != nil {
			t.Fatalf("toggle card done: %v", err)
		}
	}
	digest, err = repo.PickDigestCards(ctx, userID, 3, 8)
	if err != nil {
		t.Fatalf("pick digest cards: %v", err)
	}
	for _, c := range digest {
		if c.RoadmapID == small {
			t.Errorf("digest still includes %q from a fully-completed roadmap", c.Text)
		}
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
		t.Fatalf("get push settings before any upsert = (enabled %v, %v), want (false, nil)", enabled, err)
	}

	now := time.Now().UTC()
	if err := repo.UpsertPushInterval(ctx, userID, 180, now.Add(-time.Minute)); err != nil {
		t.Fatalf("upsert push interval: %v", err)
	}
	intervalMin, nextPushAt, enabled, err := repo.GetPushSettings(ctx, userID)
	if err != nil || intervalMin != 180 || !enabled {
		t.Fatalf("get push settings = (%d, %v, %v, %v), want (180, ..., true, nil)", intervalMin, nextPushAt, enabled, err)
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

	// Advancing the schedule takes the user out of the due set.
	if err := repo.SetNextPush(ctx, userID, now.Add(time.Hour)); err != nil {
		t.Fatalf("set next push: %v", err)
	}
	due, err = repo.ListDueUsers(ctx, now, 10)
	if err != nil {
		t.Fatalf("list due users: %v", err)
	}
	for _, u := range due {
		if u.DBUserID == userID {
			t.Fatal("user still due after next_push_at was moved into the future")
		}
	}

	if err := repo.DisablePush(ctx, userID); err != nil {
		t.Fatalf("disable push: %v", err)
	}
	if _, _, enabled, err := repo.GetPushSettings(ctx, userID); err != nil || enabled {
		t.Fatalf("get push settings after disable = (enabled %v, %v), want (false, nil)", enabled, err)
	}

	// The interval range is enforced by the DB CHECK, surfaced as a typed
	// error rather than a raw pg error.
	if err := repo.UpsertPushInterval(ctx, userID, 0, now); !errors.Is(err, errlocal.ErrRoadmapInvalidInterval) {
		t.Fatalf("upsert 0 min: got %v, want ErrRoadmapInvalidInterval", err)
	}
	if err := repo.UpsertPushInterval(ctx, userID, 1441, now); !errors.Is(err, errlocal.ErrRoadmapInvalidInterval) {
		t.Fatalf("upsert 1441 min: got %v, want ErrRoadmapInvalidInterval", err)
	}
}

// TestRoadmapRepository_CardStats checks the per-roadmap statistics rows
// include goal text and exclude archived roadmaps.
func TestRoadmapRepository_CardStats(t *testing.T) {
	pool := testPool(t)
	repo := NewRoadmapRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	id, err := repo.CreateRoadmap(ctx, userID, "Go")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	if err := repo.SetGoal(ctx, userID, id, "ship it"); err != nil {
		t.Fatalf("set goal: %v", err)
	}
	if _, err := repo.AddCards(ctx, id, []string{"a", "b"}); err != nil {
		t.Fatalf("add cards: %v", err)
	}
	cards, err := repo.ListCards(ctx, userID, id)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if _, err := repo.ToggleCardDone(ctx, userID, cards[0].ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	archivedID, err := repo.CreateRoadmap(ctx, userID, "Elixir")
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	if err := repo.ArchiveRoadmap(ctx, userID, archivedID); err != nil {
		t.Fatalf("archive: %v", err)
	}

	stats, err := repo.GetRoadmapCardStats(ctx, userID)
	if err != nil {
		t.Fatalf("get roadmap card stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("stats rows = %d, want 1 (archived excluded)", len(stats))
	}
	if stats[0].Name != "Go" || stats[0].Goal != "ship it" || stats[0].TotalCards != 2 || stats[0].DoneCards != 1 {
		t.Errorf("stats row = %+v, want Go/ship it/2/1", stats[0])
	}
}
