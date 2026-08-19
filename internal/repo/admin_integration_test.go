//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

// TestAdminRepository_GetOverviewStats checks the bot-wide counts move as
// expected when a user activates a timer, review pushes, and creates
// tracked/learning data.
func TestAdminRepository_GetOverviewStats(t *testing.T) {
	pool := testPool(t)
	adminRepo := NewAdminRepository(pool)
	trackRepo := NewTrackerRepository(pool)
	timerRepo := NewTimerRepository(pool)
	learningRepo := NewLearningRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	before, err := adminRepo.GetOverviewStats(ctx)
	if err != nil {
		t.Fatalf("get overview stats (before): %v", err)
	}

	if _, err := trackRepo.Create(ctx, userID, "Reading", "📖"); err != nil {
		t.Fatalf("create activity: %v", err)
	}
	if err := timerRepo.UpsertInterval(ctx, userID, 30, time.Now().UTC().Add(30*time.Minute)); err != nil {
		t.Fatalf("upsert timer interval: %v", err)
	}
	collID, err := learningRepo.CreateCollection(ctx, userID, "Words")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if err := learningRepo.UpsertPushInterval(ctx, userID, 60, time.Now().UTC().Add(60*time.Minute)); err != nil {
		t.Fatalf("upsert push interval: %v", err)
	}
	if _, err := learningRepo.AddWords(ctx, collID, wordPairs("cat", "кот")); err != nil {
		t.Fatalf("add words: %v", err)
	}

	after, err := adminRepo.GetOverviewStats(ctx)
	if err != nil {
		t.Fatalf("get overview stats (after): %v", err)
	}

	if after.TotalUsers != before.TotalUsers {
		t.Fatalf("TotalUsers = %d, want unchanged %d (the test user was created before this snapshot)", after.TotalUsers, before.TotalUsers)
	}
	if after.ActiveTrackTimers != before.ActiveTrackTimers+1 {
		t.Fatalf("ActiveTrackTimers = %d, want %d", after.ActiveTrackTimers, before.ActiveTrackTimers+1)
	}
	if after.ActiveReviewPushes != before.ActiveReviewPushes+1 {
		t.Fatalf("ActiveReviewPushes = %d, want %d", after.ActiveReviewPushes, before.ActiveReviewPushes+1)
	}
	if after.TotalActivities != before.TotalActivities+1 {
		t.Fatalf("TotalActivities = %d, want %d", after.TotalActivities, before.TotalActivities+1)
	}
	if after.TotalCollections != before.TotalCollections+1 {
		t.Fatalf("TotalCollections = %d, want %d", after.TotalCollections, before.TotalCollections+1)
	}
	if after.TotalLearningWords != before.TotalLearningWords+1 {
		t.Fatalf("TotalLearningWords = %d, want %d", after.TotalLearningWords, before.TotalLearningWords+1)
	}
}

// TestAdminRepository_GetUserDetail checks the per-user drill-down counts
// and active flags.
func TestAdminRepository_GetUserDetail(t *testing.T) {
	pool := testPool(t)
	adminRepo := NewAdminRepository(pool)
	trackRepo := NewTrackerRepository(pool)
	timerRepo := NewTimerRepository(pool)
	learningRepo := NewLearningRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	if _, err := trackRepo.Create(ctx, userID, "Reading", "📖"); err != nil {
		t.Fatalf("create activity: %v", err)
	}
	if err := timerRepo.UpsertInterval(ctx, userID, 30, time.Now().UTC().Add(30*time.Minute)); err != nil {
		t.Fatalf("upsert timer interval: %v", err)
	}
	collID, err := learningRepo.CreateCollection(ctx, userID, "Words")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, err := learningRepo.AddWords(ctx, collID, wordPairs("cat", "кот", "dog", "собака")); err != nil {
		t.Fatalf("add words: %v", err)
	}

	detail, err := adminRepo.GetUserDetail(ctx, userID)
	if err != nil {
		t.Fatalf("get user detail: %v", err)
	}
	if detail.DBID != userID {
		t.Fatalf("DBID = %d, want %d", detail.DBID, userID)
	}
	if detail.ActivitiesCount != 1 {
		t.Fatalf("ActivitiesCount = %d, want 1", detail.ActivitiesCount)
	}
	if detail.CollectionsCount != 1 {
		t.Fatalf("CollectionsCount = %d, want 1", detail.CollectionsCount)
	}
	if detail.LearningWords != 2 {
		t.Fatalf("LearningWords = %d, want 2", detail.LearningWords)
	}
	if !detail.TrackTimerActive {
		t.Fatal("TrackTimerActive = false, want true")
	}
	if detail.ReviewsActive {
		t.Fatal("ReviewsActive = true, want false (never activated)")
	}
}

// TestAdminRepository_GetUserDetail_NotFound checks a missing user surfaces
// the domain error instead of a zero-value detail.
func TestAdminRepository_GetUserDetail_NotFound(t *testing.T) {
	pool := testPool(t)
	adminRepo := NewAdminRepository(pool)
	ctx := context.Background()

	_, err := adminRepo.GetUserDetail(ctx, -1)
	if !errors.Is(err, errlocal.ErrUserNotFound) {
		t.Fatalf("get user detail for missing user: got %v, want ErrUserNotFound", err)
	}
}

// TestEntryRepository_ListAllTelegramIDs_And_Delete exercises the two new
// EntryRepository methods used by the admin broadcast/delete features.
func TestEntryRepository_ListAllTelegramIDs_And_Delete(t *testing.T) {
	pool := testPool(t)
	entryRepo := NewEntryRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	ids, err := entryRepo.ListAllTelegramIDs(ctx)
	if err != nil {
		t.Fatalf("list all telegram ids: %v", err)
	}
	found := false
	for _, id := range ids {
		if id != 0 {
			found = true
		}
	}
	if !found || len(ids) == 0 {
		t.Fatalf("list all telegram ids = %v, want at least the freshly created test user", ids)
	}

	if err := entryRepo.Delete(ctx, userID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	// Deleting again must report "not found" rather than silently
	// succeeding — proves the first delete actually removed the row.
	if err := entryRepo.Delete(ctx, userID); !errors.Is(err, errlocal.ErrUserNotFound) {
		t.Fatalf("delete already-deleted user: got %v, want ErrUserNotFound", err)
	}
}
