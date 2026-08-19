//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

// wordPairs builds term/translation pairs from an alternating (term,
// translation, term, translation, ...) argument list, for AddWords calls.
func wordPairs(termsAndTranslations ...string) []errlocal.LearningWordItem {
	out := make([]errlocal.LearningWordItem, 0, len(termsAndTranslations)/2)
	for i := 0; i+1 < len(termsAndTranslations); i += 2 {
		out = append(out, errlocal.LearningWordItem{Term: termsAndTranslations[i], Translation: termsAndTranslations[i+1]})
	}
	return out
}

// TestLearningRepository_CollectionLifecycle exercises create -> list ->
// toggle -> archive -> restore -> delete forever against real Postgres,
// including the unique-name constraint.
func TestLearningRepository_CollectionLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := NewLearningRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	id, err := repo.CreateCollection(ctx, userID, "Animals")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}

	if _, err := repo.CreateCollection(ctx, userID, "Animals"); !errors.Is(err, errlocal.ErrLearningCollectionExists) {
		t.Fatalf("create duplicate: got %v, want ErrLearningCollectionExists", err)
	}

	second, err := repo.CreateCollection(ctx, userID, "Fruits")
	if err != nil {
		t.Fatalf("create second collection: %v", err)
	}
	if err := repo.RenameCollection(ctx, userID, id, "Pets"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if got, err := repo.GetCollectionName(ctx, userID, id); err != nil || got != "Pets" {
		t.Fatalf("get collection name after rename = (%q, %v), want (Pets, nil)", got, err)
	}
	if err := repo.RenameCollection(ctx, userID, second, "Pets"); !errors.Is(err, errlocal.ErrLearningCollectionExists) {
		t.Fatalf("rename to existing name: got %v, want ErrLearningCollectionExists", err)
	}
	if err := repo.DeleteCollectionForever(ctx, userID, second); err != nil {
		t.Fatalf("cleanup second collection: %v", err)
	}

	items, err := repo.ListCollections(ctx, userID, false)
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(items) != 1 || items[0].ID != id || !items[0].Active || items[0].IsArchived {
		t.Fatalf("list active = %+v, want one active non-archived collection", items)
	}

	if err := repo.ToggleCollectionActive(ctx, userID, id); err != nil {
		t.Fatalf("toggle active: %v", err)
	}
	items, _ = repo.ListCollections(ctx, userID, false)
	if items[0].Active {
		t.Fatalf("collection still active after toggle")
	}

	if err := repo.ArchiveCollection(ctx, userID, id); err != nil {
		t.Fatalf("archive: %v", err)
	}
	active, _ := repo.ListCollections(ctx, userID, false)
	if len(active) != 0 {
		t.Fatalf("active list after archive = %+v, want empty", active)
	}
	archived, _ := repo.ListCollections(ctx, userID, true)
	if len(archived) != 1 || archived[0].ID != id {
		t.Fatalf("archived list = %+v, want one entry", archived)
	}

	if err := repo.RestoreCollection(ctx, userID, id); err != nil {
		t.Fatalf("restore: %v", err)
	}
	active, _ = repo.ListCollections(ctx, userID, false)
	if len(active) != 1 {
		t.Fatalf("active list after restore = %+v, want one entry", active)
	}

	if err := repo.DeleteCollectionForever(ctx, userID, id); err != nil {
		t.Fatalf("delete forever: %v", err)
	}
	active, _ = repo.ListCollections(ctx, userID, false)
	if len(active) != 0 {
		t.Fatalf("active list after delete forever = %+v, want empty", active)
	}
}

// TestLearningRepository_WordsAndDuePicking checks words are inserted due
// immediately, PickDueWord only returns words from active/non-archived
// collections, and deleting a collection cascades to its words.
func TestLearningRepository_WordsAndDuePicking(t *testing.T) {
	pool := testPool(t)
	repo := NewLearningRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	activeID, err := repo.CreateCollection(ctx, userID, "Active")
	if err != nil {
		t.Fatalf("create active collection: %v", err)
	}
	inactiveID, err := repo.CreateCollection(ctx, userID, "Inactive")
	if err != nil {
		t.Fatalf("create inactive collection: %v", err)
	}
	if err := repo.ToggleCollectionActive(ctx, userID, inactiveID); err != nil {
		t.Fatalf("toggle inactive: %v", err)
	}

	n, err := repo.AddWords(ctx, activeID, wordPairs("cat", "кот", "dog", "собака"))
	if err != nil {
		t.Fatalf("add words active: %v", err)
	}
	if n != 2 {
		t.Fatalf("added = %d, want 2", n)
	}
	if _, err := repo.AddWords(ctx, inactiveID, wordPairs("fox", "лиса")); err != nil {
		t.Fatalf("add words inactive: %v", err)
	}

	words, err := repo.ListWords(ctx, userID, activeID)
	if err != nil {
		t.Fatalf("list words: %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("list words = %+v, want 2", words)
	}

	due, err := repo.PickDueWord(ctx, userID, time.Now().UTC())
	if err != nil {
		t.Fatalf("pick due word: %v", err)
	}
	if due == nil {
		t.Fatal("pick due word = nil, want a word from the active collection")
	}
	if due.CollectionName != "Active" {
		t.Fatalf("due.CollectionName = %q, want %q (must not pick from the inactive collection)", due.CollectionName, "Active")
	}

	if err := repo.DeleteWord(ctx, userID, words[0].ID); err != nil {
		t.Fatalf("delete word: %v", err)
	}
	words, _ = repo.ListWords(ctx, userID, activeID)
	if len(words) != 1 {
		t.Fatalf("list words after delete = %+v, want 1", words)
	}

	// Deleting the collection must cascade to its remaining word — ListWords
	// joins through the collection, so a gone collection just yields an
	// empty list rather than an error.
	if err := repo.DeleteCollectionForever(ctx, userID, activeID); err != nil {
		t.Fatalf("delete collection forever: %v", err)
	}
	words, err = repo.ListWords(ctx, userID, activeID)
	if err != nil {
		t.Fatalf("list words after collection delete: %v", err)
	}
	if len(words) != 0 {
		t.Fatalf("list words after collection delete = %+v, want empty (cascade should have removed it)", words)
	}
}

// TestLearningRepository_PushSettings exercises the push-scheduling rows
// (mirrors user_timer_settings), including the CHECK constraint.
func TestLearningRepository_PushSettings(t *testing.T) {
	pool := testPool(t)
	repo := NewLearningRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	next := time.Now().UTC().Add(60 * time.Minute)
	if err := repo.UpsertPushInterval(ctx, userID, 60, next); err != nil {
		t.Fatalf("upsert push interval: %v", err)
	}

	interval, nextPushAt, enabled, err := repo.GetPushSettings(ctx, userID)
	if err != nil {
		t.Fatalf("get push settings: %v", err)
	}
	if interval != 60 || !enabled || nextPushAt.Sub(next).Abs() > time.Second {
		t.Fatalf("got (interval=%d enabled=%v nextPushAt=%v), want (60 true ~%v)", interval, enabled, nextPushAt, next)
	}

	due, err := repo.ListDueUsers(ctx, next.Add(time.Minute), 100)
	if err != nil {
		t.Fatalf("list due users: %v", err)
	}
	found := false
	for _, u := range due {
		if u.DBUserID == userID {
			found = true
		}
	}
	if !found {
		t.Fatalf("list due users = %+v, want to include userID %d", due, userID)
	}

	if err := repo.DisablePush(ctx, userID); err != nil {
		t.Fatalf("disable push: %v", err)
	}
	_, _, enabled, _ = repo.GetPushSettings(ctx, userID)
	if enabled {
		t.Fatal("enabled = true after DisablePush, want false")
	}

	if err := repo.UpsertPushInterval(ctx, userID, 5000, next); !errors.Is(err, errlocal.ErrLearningInvalidInterval) {
		t.Fatalf("upsert out-of-range interval: got %v, want ErrLearningInvalidInterval", err)
	}
}

// TestLearningRepository_StatsDetail checks the per-collection breakdown
// and accuracy queries behind the "📈 Statistics" screen.
func TestLearningRepository_StatsDetail(t *testing.T) {
	pool := testPool(t)
	repo := NewLearningRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	collID, err := repo.CreateCollection(ctx, userID, "Animals")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, err := repo.AddWords(ctx, collID, wordPairs("cat", "кот", "dog", "собака")); err != nil {
		t.Fatalf("add words: %v", err)
	}

	stats, err := repo.GetCollectionStats(ctx, userID)
	if err != nil {
		t.Fatalf("get collection stats: %v", err)
	}
	if len(stats) != 1 || stats[0].Name != "Animals" || stats[0].TotalWords != 2 || stats[0].DueWords != 2 {
		t.Fatalf("collection stats = %+v, want one Animals row with TotalWords=2 DueWords=2", stats)
	}

	words, err := repo.ListWords(ctx, userID, collID)
	if err != nil || len(words) != 2 {
		t.Fatalf("list words: %v, %+v", err, words)
	}
	if err := repo.RecordReview(ctx, words[0].ID, userID, true, time.Now().UTC()); err != nil {
		t.Fatalf("record review (correct): %v", err)
	}
	if err := repo.RecordReview(ctx, words[1].ID, userID, false, time.Now().UTC()); err != nil {
		t.Fatalf("record review (incorrect): %v", err)
	}

	correct, total, err := repo.GetAccuracy(ctx, userID)
	if err != nil {
		t.Fatalf("get accuracy: %v", err)
	}
	if correct != 1 || total != 2 {
		t.Fatalf("accuracy = (%d, %d), want (1, 2)", correct, total)
	}

	now := time.Now().UTC()
	entries, err := repo.ListReviewsOnDay(ctx, userID, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("list reviews on day: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("list reviews on day = %+v, want 2 entries", entries)
	}
	entries, err = repo.ListReviewsOnDay(ctx, userID, now.Add(-48*time.Hour), now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("list reviews on day (out of range): %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("list reviews on day (out of range) = %+v, want empty", entries)
	}
}
