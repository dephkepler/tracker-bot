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
// Hand-written in-memory fake (see timer_test.go for the same convention) —
// keeps these tests at the bottom of the pyramid. Real SQL behavior is
// covered separately by internal/repo/*_integration_test.go against a real
// Postgres.

type fakeCollection struct {
	id       int64
	userID   int64
	name     string
	active   bool
	archived bool
}

type fakeWord struct {
	id           int64
	collectionID int64
	term         string
	translation  string
	easeFactor   float32
	intervalDays int
	repetitions  int
	nextReviewAt time.Time
	learned      bool
}

type fakeLearningRepo struct {
	nextCollectionID int64
	nextWordID       int64
	collections      map[int64]*fakeCollection
	words            map[int64]*fakeWord
	pushRows         map[int64]struct {
		intervalMin int
		nextPushAt  time.Time
		enabled     bool
	}
	reviewDates map[int64][]time.Time
}

func newFakeLearningRepo() *fakeLearningRepo {
	return &fakeLearningRepo{
		collections: map[int64]*fakeCollection{},
		words:       map[int64]*fakeWord{},
		pushRows: map[int64]struct {
			intervalMin int
			nextPushAt  time.Time
			enabled     bool
		}{},
		reviewDates: map[int64][]time.Time{},
	}
}

func (f *fakeLearningRepo) CreateCollection(_ context.Context, userID int64, name string) (int64, error) {
	for _, c := range f.collections {
		if c.userID == userID && c.name == name {
			return 0, models.ErrLearningCollectionExists
		}
	}
	f.nextCollectionID++
	id := f.nextCollectionID
	f.collections[id] = &fakeCollection{id: id, userID: userID, name: name, active: true}
	return id, nil
}

func (f *fakeLearningRepo) ListCollections(_ context.Context, userID int64, archived bool) ([]models.LearningCollectionItem, error) {
	out := make([]models.LearningCollectionItem, 0)
	for _, c := range f.collections {
		if c.userID != userID || c.archived != archived {
			continue
		}
		count := 0
		for _, w := range f.words {
			if w.collectionID == c.id {
				count++
			}
		}
		out = append(out, models.LearningCollectionItem{ID: c.id, Name: c.name, WordCount: count, Active: c.active, IsArchived: c.archived})
	}
	return out, nil
}

func (f *fakeLearningRepo) GetCollectionName(_ context.Context, userID, collectionID int64) (string, error) {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return "", models.ErrLearningCollectionNotFound
	}
	return c.name, nil
}

func (f *fakeLearningRepo) RenameCollection(_ context.Context, userID, collectionID int64, newName string) error {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return models.ErrLearningCollectionNotFound
	}
	for _, other := range f.collections {
		if other.id != collectionID && other.userID == userID && other.name == newName {
			return models.ErrLearningCollectionExists
		}
	}
	c.name = newName
	return nil
}

func (f *fakeLearningRepo) ToggleCollectionActive(_ context.Context, userID, collectionID int64) error {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return models.ErrLearningCollectionNotFound
	}
	c.active = !c.active
	return nil
}

func (f *fakeLearningRepo) ArchiveCollection(_ context.Context, userID, collectionID int64) error {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return models.ErrLearningCollectionNotFound
	}
	c.archived = true
	return nil
}

func (f *fakeLearningRepo) RestoreCollection(_ context.Context, userID, collectionID int64) error {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return models.ErrLearningCollectionNotFound
	}
	c.archived = false
	return nil
}

func (f *fakeLearningRepo) DeleteCollectionForever(_ context.Context, userID, collectionID int64) error {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return models.ErrLearningCollectionNotFound
	}
	delete(f.collections, collectionID)
	for id, w := range f.words {
		if w.collectionID == collectionID {
			delete(f.words, id)
		}
	}
	return nil
}

func (f *fakeLearningRepo) AddWords(_ context.Context, collectionID int64, pairs []models.LearningWordItem) (int, error) {
	for _, p := range pairs {
		f.nextWordID++
		f.words[f.nextWordID] = &fakeWord{
			id: f.nextWordID, collectionID: collectionID, term: p.Term, translation: p.Translation,
			easeFactor: 2.5, nextReviewAt: time.Now().UTC(),
		}
	}
	return len(pairs), nil
}

func (f *fakeLearningRepo) ListWords(_ context.Context, userID, collectionID int64) ([]models.LearningWordItem, error) {
	c, ok := f.collections[collectionID]
	if !ok || c.userID != userID {
		return nil, models.ErrLearningCollectionNotFound
	}
	out := make([]models.LearningWordItem, 0)
	for _, w := range f.words {
		if w.collectionID == collectionID {
			out = append(out, models.LearningWordItem{ID: w.id, Term: w.term, Translation: w.translation, Learned: w.learned, NextReviewAt: w.nextReviewAt, IntervalDays: w.intervalDays, Repetitions: w.repetitions})
		}
	}
	return out, nil
}

func (f *fakeLearningRepo) DeleteWord(_ context.Context, userID, wordID int64) error {
	w, ok := f.words[wordID]
	if !ok {
		return models.ErrLearningWordNotFound
	}
	c, ok := f.collections[w.collectionID]
	if !ok || c.userID != userID {
		return models.ErrLearningWordNotFound
	}
	delete(f.words, wordID)
	return nil
}

func (f *fakeLearningRepo) GetWordForGrading(_ context.Context, userID, wordID int64) (string, string, string, float32, int, int, error) {
	w, ok := f.words[wordID]
	if !ok {
		return "", "", "", 0, 0, 0, models.ErrLearningWordNotFound
	}
	c, ok := f.collections[w.collectionID]
	if !ok || c.userID != userID {
		return "", "", "", 0, 0, 0, models.ErrLearningWordNotFound
	}
	return c.name, w.term, w.translation, w.easeFactor, w.intervalDays, w.repetitions, nil
}

func (f *fakeLearningRepo) UpdateWordSchedule(_ context.Context, wordID int64, easeFactor float32, intervalDays, repetitions int, nextReviewAt time.Time, learned bool) error {
	w, ok := f.words[wordID]
	if !ok {
		return models.ErrLearningWordNotFound
	}
	w.easeFactor = easeFactor
	w.intervalDays = intervalDays
	w.repetitions = repetitions
	w.nextReviewAt = nextReviewAt
	w.learned = learned
	return nil
}

func (f *fakeLearningRepo) PickDueWord(_ context.Context, userID int64, now time.Time) (*models.LearningDueWord, error) {
	var best *fakeWord
	for _, w := range f.words {
		c, ok := f.collections[w.collectionID]
		if !ok || c.userID != userID || !c.active || c.archived || w.learned {
			continue
		}
		if w.nextReviewAt.After(now) {
			continue
		}
		if best == nil || w.nextReviewAt.Before(best.nextReviewAt) {
			best = w
		}
	}
	if best == nil {
		return nil, nil
	}
	c := f.collections[best.collectionID]
	return &models.LearningDueWord{ID: best.id, CollectionName: c.name, Term: best.term, Translation: best.translation}, nil
}

func (f *fakeLearningRepo) RecordReview(_ context.Context, wordID, userID int64, correct bool, now time.Time) error {
	f.reviewDates[userID] = append(f.reviewDates[userID], now)
	return nil
}

func (f *fakeLearningRepo) ListReviewsOnDay(_ context.Context, userID int64, from, to time.Time) ([]models.LearningReviewEntry, error) {
	out := make([]models.LearningReviewEntry, 0)
	for _, d := range f.reviewDates[userID] {
		if !d.Before(from) && d.Before(to) {
			out = append(out, models.LearningReviewEntry{ReviewedAt: d})
		}
	}
	return out, nil
}

func (f *fakeLearningRepo) UpsertPushInterval(_ context.Context, userID int64, intervalMin int, nextPushAt time.Time) error {
	f.pushRows[userID] = struct {
		intervalMin int
		nextPushAt  time.Time
		enabled     bool
	}{intervalMin, nextPushAt, true}
	return nil
}

func (f *fakeLearningRepo) GetPushSettings(_ context.Context, userID int64) (int, time.Time, bool, error) {
	row, ok := f.pushRows[userID]
	if !ok {
		return 0, time.Time{}, false, nil
	}
	return row.intervalMin, row.nextPushAt, row.enabled, nil
}

func (f *fakeLearningRepo) SetNextPush(_ context.Context, userID int64, nextPushAt time.Time) error {
	row, ok := f.pushRows[userID]
	if !ok {
		return errors.New("no row")
	}
	row.nextPushAt = nextPushAt
	f.pushRows[userID] = row
	return nil
}

func (f *fakeLearningRepo) DisablePush(_ context.Context, userID int64) error {
	row := f.pushRows[userID]
	row.enabled = false
	f.pushRows[userID] = row
	return nil
}

func (f *fakeLearningRepo) ListDueUsers(_ context.Context, now time.Time, limit int) ([]models.LearningDueUser, error) {
	return nil, nil
}

func (f *fakeLearningRepo) CountWords(_ context.Context, userID int64) (int, int, int, error) {
	total, due, learned := 0, 0, 0
	now := time.Now().UTC()
	for _, w := range f.words {
		c, ok := f.collections[w.collectionID]
		if !ok || c.userID != userID || c.archived {
			continue
		}
		total++
		if w.learned {
			learned++
		} else if !w.nextReviewAt.After(now) {
			due++
		}
	}
	return total, due, learned, nil
}

func (f *fakeLearningRepo) ListReviewDates(_ context.Context, userID int64, since time.Time) ([]time.Time, error) {
	out := make([]time.Time, 0)
	for _, d := range f.reviewDates[userID] {
		if !d.Before(since) {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *fakeLearningRepo) GetCollectionStats(_ context.Context, userID int64) ([]models.LearningCollectionStat, error) {
	out := make([]models.LearningCollectionStat, 0)
	for _, c := range f.collections {
		if c.userID != userID || c.archived {
			continue
		}
		var s models.LearningCollectionStat
		s.Name = c.name
		now := time.Now().UTC()
		for _, w := range f.words {
			if w.collectionID != c.id {
				continue
			}
			s.TotalWords++
			if w.learned {
				s.LearnedWords++
			} else if !w.nextReviewAt.After(now) {
				s.DueWords++
			}
		}
		out = append(out, s)
	}
	return out, nil
}

func (f *fakeLearningRepo) GetAccuracy(_ context.Context, userID int64) (int, int, error) {
	correct, total := 0, 0
	for uid, dates := range f.reviewDates {
		if uid != userID {
			continue
		}
		total += len(dates)
	}
	return correct, total, nil
}

func newTestLearningService() (*fakeLearningRepo, LearningService) {
	repo := newFakeLearningRepo()
	return repo, NewLearningService(repo)
}

// --- gradeSchedule (Anki-style four-grade SM-2 variant) ------------------

func TestGradeSchedule_FirstGood(t *testing.T) {
	ease, interval, reps, learned := gradeSchedule(2.5, 0, 0, models.LearningGradeGood)
	if interval != 1 || reps != 1 || learned {
		t.Fatalf("got (ease=%v interval=%d reps=%d learned=%v), want interval=1 reps=1 learned=false", ease, interval, reps, learned)
	}
}

func TestGradeSchedule_SecondGood(t *testing.T) {
	_, interval, reps, learned := gradeSchedule(2.6, 1, 1, models.LearningGradeGood)
	if interval != 6 || reps != 2 || learned {
		t.Fatalf("got (interval=%d reps=%d learned=%v), want interval=6 reps=2 learned=false", interval, reps, learned)
	}
}

func TestGradeSchedule_GoodGrowsWithEase(t *testing.T) {
	_, interval, reps, _ := gradeSchedule(2.5, 6, 2, models.LearningGradeGood)
	if interval <= 6 {
		t.Fatalf("interval = %d, want > 6 (must grow past the previous interval)", interval)
	}
	if reps != 3 {
		t.Fatalf("reps = %d, want 3", reps)
	}
}

func TestGradeSchedule_GraduatesToLearned(t *testing.T) {
	// Interval already at 15 days with a high ease factor — the next Good
	// answer should push it past the 21-day learned threshold.
	_, interval, _, learned := gradeSchedule(2.5, 15, 3, models.LearningGradeGood)
	if interval < 21 || !learned {
		t.Fatalf("got (interval=%d learned=%v), want interval>=21 learned=true", interval, learned)
	}
}

func TestGradeSchedule_AgainResets(t *testing.T) {
	ease, interval, reps, learned := gradeSchedule(2.5, 40, 5, models.LearningGradeAgain)
	if interval != 1 || reps != 0 || learned {
		t.Fatalf("got (interval=%d reps=%d learned=%v), want interval=1 reps=0 learned=false", interval, reps, learned)
	}
	if ease >= 2.5 {
		t.Fatalf("ease = %v, want lower than before Again", ease)
	}
}

func TestGradeSchedule_HardGrowsSlowerThanGood(t *testing.T) {
	_, hardInterval, hardReps, _ := gradeSchedule(2.5, 10, 3, models.LearningGradeHard)
	_, goodInterval, _, _ := gradeSchedule(2.5, 10, 3, models.LearningGradeGood)
	if hardInterval >= goodInterval {
		t.Fatalf("hard interval = %d, want < good interval %d", hardInterval, goodInterval)
	}
	if hardInterval <= 10 {
		t.Fatalf("hard interval = %d, want to still grow past the previous interval", hardInterval)
	}
	if hardReps != 4 {
		t.Fatalf("hard reps = %d, want 4", hardReps)
	}
}

func TestGradeSchedule_HardOnNewWordStaysShort(t *testing.T) {
	_, interval, _, _ := gradeSchedule(2.5, 0, 0, models.LearningGradeHard)
	if interval != 1 {
		t.Fatalf("hard interval on a brand-new word = %d, want 1", interval)
	}
}

func TestGradeSchedule_EasyGrowsFasterThanGood(t *testing.T) {
	_, easyInterval, easyReps, _ := gradeSchedule(2.5, 10, 3, models.LearningGradeEasy)
	_, goodInterval, _, _ := gradeSchedule(2.5, 10, 3, models.LearningGradeGood)
	if easyInterval <= goodInterval {
		t.Fatalf("easy interval = %d, want > good interval %d", easyInterval, goodInterval)
	}
	if easyReps != 4 {
		t.Fatalf("easy reps = %d, want 4", easyReps)
	}
}

func TestGradeSchedule_EasyOnNewWordSkipsAhead(t *testing.T) {
	_, interval, _, _ := gradeSchedule(2.5, 0, 0, models.LearningGradeEasy)
	if interval != 4 {
		t.Fatalf("easy interval on a brand-new word = %d, want 4", interval)
	}
}

func TestGradeSchedule_EaseFactorBounds(t *testing.T) {
	// Many consecutive Again grades must not push ease below SM-2's floor.
	ease := float32(1.3)
	for i := 0; i < 10; i++ {
		ease, _, _, _ = gradeSchedule(ease, 10, 3, models.LearningGradeAgain)
	}
	if ease < 1.3 {
		t.Fatalf("ease = %v, want >= 1.3 (SM-2 floor)", ease)
	}

	// Many consecutive Easy grades must not push ease above the cap.
	ease = 2.5
	interval, reps := 1, 0
	for i := 0; i < 10; i++ {
		ease, interval, reps, _ = gradeSchedule(ease, interval, reps, models.LearningGradeEasy)
	}
	if ease > 2.5 {
		t.Fatalf("ease = %v, want <= 2.5 (cap)", ease)
	}
}

// --- parseWordLine / AddWordsFromText ------------------------------------

func TestParseWordLine(t *testing.T) {
	cases := []struct {
		line            string
		wantTerm        string
		wantTranslation string
		wantOK          bool
	}{
		{"apple - яблоко", "apple", "яблоко", true},
		{"apple – яблоко", "apple", "яблоко", true},
		{"apple: яблоко", "apple", "яблоко", true},
		{"apple-яблоко", "apple", "яблоко", true},
		// Mobile "smart punctuation" autocorrect turns a typed "-" into an
		// en/em dash *without* re-adding the surrounding spaces — this must
		// still parse (see the prod bug report it guards against).
		{"apple—яблоко", "apple", "яблоко", true},
		{"apple–яблоко", "apple", "яблоко", true},
		{"just some text", "", "", false},
		{"- яблоко", "", "", false},
		{"apple -", "", "", false},
		{"   ", "", "", false},
	}
	for _, tc := range cases {
		term, translation, ok := parseWordLine(tc.line)
		if ok != tc.wantOK {
			t.Errorf("parseWordLine(%q): ok = %v, want %v", tc.line, ok, tc.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if term != tc.wantTerm || translation != tc.wantTranslation {
			t.Errorf("parseWordLine(%q) = (%q, %q), want (%q, %q)", tc.line, term, translation, tc.wantTerm, tc.wantTranslation)
		}
	}
}

func TestLearningService_AddWordsFromText_MixedValidAndInvalid(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()
	id, err := svc.CreateCollection(ctx, 1, "Animals")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}

	text := "cat - кот\nnot a valid line\ndog - собака\n\n"
	added, skipped, err := svc.AddWordsFromText(ctx, 1, id, text)
	if err != nil {
		t.Fatalf("add words: %v", err)
	}
	if added != 2 || skipped != 1 {
		t.Fatalf("added=%d skipped=%d, want added=2 skipped=1", added, skipped)
	}
	if len(repo.words) != 2 {
		t.Fatalf("repo has %d words, want 2", len(repo.words))
	}
}

func TestLearningService_AddWordsFromText_NoValidLines(t *testing.T) {
	_, svc := newTestLearningService()
	ctx := context.Background()
	id, _ := svc.CreateCollection(ctx, 1, "Animals")

	_, _, err := svc.AddWordsFromText(ctx, 1, id, "nothing here\nor here")
	if !errors.Is(err, models.ErrLearningNoWordsParsed) {
		t.Fatalf("got %v, want ErrLearningNoWordsParsed", err)
	}
}

// --- RenameCollection --------------------------------------------------

func TestLearningService_RenameCollection_Success(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()
	id, _ := svc.CreateCollection(ctx, 1, "Animals")

	if err := svc.RenameCollection(ctx, 1, id, "  Pets  "); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if got := repo.collections[id].name; got != "Pets" {
		t.Fatalf("collection name = %q, want %q (trimmed)", got, "Pets")
	}
}

func TestLearningService_RenameCollection_RejectsBadInput(t *testing.T) {
	_, svc := newTestLearningService()
	ctx := context.Background()
	id, _ := svc.CreateCollection(ctx, 1, "Animals")

	cases := []string{"", "a", "multi\nline", string(make([]byte, 61))}
	for _, name := range cases {
		if err := svc.RenameCollection(ctx, 1, id, name); !errors.Is(err, models.ErrLearningInvalidName) {
			t.Errorf("rename to %q: got %v, want ErrLearningInvalidName", name, err)
		}
	}
}

func TestLearningService_RenameCollection_DuplicateName(t *testing.T) {
	_, svc := newTestLearningService()
	ctx := context.Background()
	_, _ = svc.CreateCollection(ctx, 1, "Animals")
	id2, _ := svc.CreateCollection(ctx, 1, "Fruits")

	if err := svc.RenameCollection(ctx, 1, id2, "Animals"); !errors.Is(err, models.ErrLearningCollectionExists) {
		t.Fatalf("rename to existing name: got %v, want ErrLearningCollectionExists", err)
	}
}

// --- GetStatsDetail ------------------------------------------------------

func TestLearningService_GetStatsDetail_AggregatesEverything(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()
	id, _ := svc.CreateCollection(ctx, 1, "Animals")
	if _, _, err := svc.AddWordsFromText(ctx, 1, id, "cat - кот\ndog - собака"); err != nil {
		t.Fatalf("add words: %v", err)
	}
	repo.reviewDates[1] = []time.Time{time.Now().UTC()}

	detail, err := svc.GetStatsDetail(ctx, 1, time.UTC)
	if err != nil {
		t.Fatalf("get stats detail: %v", err)
	}
	if detail.Overall.TotalWords != 2 {
		t.Fatalf("Overall.TotalWords = %d, want 2", detail.Overall.TotalWords)
	}
	if len(detail.Collections) != 1 || detail.Collections[0].TotalWords != 2 {
		t.Fatalf("Collections = %+v, want one entry with TotalWords=2", detail.Collections)
	}
	if detail.ReviewsTotal != 1 {
		t.Fatalf("ReviewsTotal = %d, want 1", detail.ReviewsTotal)
	}
}

// --- Activate: preserves a running schedule (mirrors TimerService) ------

func TestLearningService_Activate_PreservesRunningSchedule_SameInterval(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()

	original := time.Now().UTC().Add(20 * time.Minute)
	repo.pushRows[1] = struct {
		intervalMin int
		nextPushAt  time.Time
		enabled     bool
	}{60, original, true}

	if err := svc.Activate(ctx, 1, 60); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := repo.pushRows[1].nextPushAt; !got.Equal(original) {
		t.Fatalf("nextPushAt = %v, want unchanged %v", got, original)
	}
}

func TestLearningService_Activate_InvalidInterval(t *testing.T) {
	_, svc := newTestLearningService()
	if err := svc.Activate(context.Background(), 1, 0); !errors.Is(err, models.ErrLearningInvalidInterval) {
		t.Fatalf("got %v, want ErrLearningInvalidInterval", err)
	}
}

// --- PickDueWord / GradeAnswer end-to-end via the service ----------------

func TestLearningService_PickDueWord_OnlyActiveCollections(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()

	activeID, _ := svc.CreateCollection(ctx, 1, "Active")
	inactiveID, _ := svc.CreateCollection(ctx, 1, "Inactive")
	_ = svc.ToggleCollectionActive(ctx, 1, inactiveID)

	if _, _, err := svc.AddWordsFromText(ctx, 1, activeID, "a - a1"); err != nil {
		t.Fatalf("add words active: %v", err)
	}
	if _, _, err := svc.AddWordsFromText(ctx, 1, inactiveID, "b - b1"); err != nil {
		t.Fatalf("add words inactive: %v", err)
	}

	due, err := svc.PickDueWord(ctx, 1)
	if err != nil {
		t.Fatalf("pick due word: %v", err)
	}
	if due == nil || due.Term != "a" {
		t.Fatalf("due = %+v, want word from the active collection only", due)
	}
	_ = repo
}

// --- computeStreak ---------------------------------------------------------

func TestComputeStreak_NoReviews(t *testing.T) {
	repo, svc := newTestLearningService()
	got := mustStreak(t, svc, repo, 1)
	if got != 0 {
		t.Fatalf("streak = %d, want 0", got)
	}
}

func TestComputeStreak_ConsecutiveDaysEndingToday(t *testing.T) {
	repo, svc := newTestLearningService()
	today := truncDay(time.Now().UTC())
	repo.reviewDates[1] = []time.Time{
		today,
		today.AddDate(0, 0, -1),
		today.AddDate(0, 0, -2),
	}
	if got := mustStreak(t, svc, repo, 1); got != 3 {
		t.Fatalf("streak = %d, want 3", got)
	}
}

func TestComputeStreak_StillLiveIfLastReviewWasYesterday(t *testing.T) {
	repo, svc := newTestLearningService()
	today := truncDay(time.Now().UTC())
	repo.reviewDates[1] = []time.Time{
		today.AddDate(0, 0, -1),
		today.AddDate(0, 0, -2),
	}
	if got := mustStreak(t, svc, repo, 1); got != 2 {
		t.Fatalf("streak = %d, want 2 (streak stays alive through the day it's checked)", got)
	}
}

func TestComputeStreak_BrokenByGap(t *testing.T) {
	repo, svc := newTestLearningService()
	today := truncDay(time.Now().UTC())
	repo.reviewDates[1] = []time.Time{
		today,
		today.AddDate(0, 0, -1),
		// gap at -2
		today.AddDate(0, 0, -3),
	}
	if got := mustStreak(t, svc, repo, 1); got != 2 {
		t.Fatalf("streak = %d, want 2 (must stop at the gap)", got)
	}
}

func TestComputeStreak_ExpiredTwoDaysAgo(t *testing.T) {
	repo, svc := newTestLearningService()
	today := truncDay(time.Now().UTC())
	repo.reviewDates[1] = []time.Time{today.AddDate(0, 0, -2)}
	if got := mustStreak(t, svc, repo, 1); got != 0 {
		t.Fatalf("streak = %d, want 0 (last review too long ago to still be live)", got)
	}
}

// TestComputeStreak_BucketsByUserLocalDayNotUTC is a regression test for a
// bug where computeStreak bucketed reviews by their raw UTC calendar day
// instead of the caller's own timezone (now.Location()): two reviews made on
// the *same* local calendar day, but on opposite sides of the UTC midnight
// that falls in the middle of that local day (common far from UTC, e.g.
// UTC+14), used to land in two different UTC-day buckets and inflate the
// streak. Calls computeStreak directly (unlike mustStreak below) so the
// review/now instants can be pinned exactly — real wall-clock time can't
// deterministically hit this case.
func TestComputeStreak_BucketsByUserLocalDayNotUTC(t *testing.T) {
	repo, svc := newTestLearningService()
	ls := svc.(*learningService)

	loc := time.FixedZone("UTC+14", 14*60*60)
	// Same local day (2024-03-10) for both, but review1's instant is before
	// that day's UTC midnight-in-local-time (08:00 local = 2024-03-09 18:00
	// UTC) while review2's is after (22:00 local = 2024-03-10 08:00 UTC).
	// .UTC() mirrors how pgx actually hands back TIMESTAMPTZ rows in
	// production — without it, both instants already carry loc and the bug
	// this test targets can't reproduce.
	review1 := time.Date(2024, 3, 10, 8, 0, 0, 0, loc).UTC()
	review2 := time.Date(2024, 3, 10, 22, 0, 0, 0, loc).UTC()
	repo.reviewDates[1] = []time.Time{review2, review1} // most-recent-first

	now := time.Date(2024, 3, 10, 23, 0, 0, 0, loc)
	got, err := ls.computeStreak(context.Background(), 1, now)
	if err != nil {
		t.Fatalf("compute streak: %v", err)
	}
	if got != 1 {
		t.Fatalf("streak = %d, want 1 (both reviews fall on the same local calendar day)", got)
	}
}

// mustStreak reads the streak back out via GetLearningStats (computeStreak
// is unexported and only reachable through the service's public surface).
func mustStreak(t *testing.T, svc LearningService, _ *fakeLearningRepo, userID int64) int {
	t.Helper()
	stats, err := svc.GetLearningStats(context.Background(), userID, time.UTC)
	if err != nil {
		t.Fatalf("get learning stats: %v", err)
	}
	return stats.StreakDays
}

func TestLearningService_GradeAnswer_UpdatesScheduleAndRecordsReview(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()

	id, err := svc.CreateCollection(ctx, 1, "Coll")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, _, err := svc.AddWordsFromText(ctx, 1, id, "a - a1"); err != nil {
		t.Fatalf("add words: %v", err)
	}
	var wordID int64
	for wid := range repo.words {
		wordID = wid
	}

	nextReviewAt, learned, err := svc.GradeAnswer(ctx, 1, wordID, models.LearningGradeGood)
	if err != nil {
		t.Fatalf("grade answer: %v", err)
	}
	if learned {
		t.Fatalf("learned = true, want false for a first Good answer")
	}
	wantAt := time.Now().UTC().Add(24 * time.Hour)
	if nextReviewAt.Before(wantAt.Add(-time.Minute)) || nextReviewAt.After(wantAt.Add(time.Minute)) {
		t.Fatalf("nextReviewAt = %v, want ~1 day from now (%v)", nextReviewAt, wantAt)
	}
	if len(repo.reviewDates[1]) != 1 {
		t.Fatalf("review history has %d entries, want 1", len(repo.reviewDates[1]))
	}
	if repo.words[wordID].nextReviewAt.Before(time.Now().UTC().Add(23 * time.Hour)) {
		t.Fatalf("nextReviewAt = %v, want ~1 day from now", repo.words[wordID].nextReviewAt)
	}
}

// TestLearningService_GradeAnswer_AgainUsesMinuteStep checks Again schedules
// a short 10-minute step, not a full day — the bug report this guards
// against showed a flat "next review tomorrow" regardless of grade.
func TestLearningService_GradeAnswer_AgainUsesMinuteStep(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()
	id, err := svc.CreateCollection(ctx, 1, "Coll")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, _, err := svc.AddWordsFromText(ctx, 1, id, "a - a1"); err != nil {
		t.Fatalf("add words: %v", err)
	}
	var wordID int64
	for wid := range repo.words {
		wordID = wid
	}

	nextReviewAt, learned, err := svc.GradeAnswer(ctx, 1, wordID, models.LearningGradeAgain)
	if err != nil {
		t.Fatalf("grade answer: %v", err)
	}
	if learned {
		t.Fatal("learned = true, want false for Again")
	}
	until := time.Until(nextReviewAt)
	if until <= 0 || until > 15*time.Minute {
		t.Fatalf("next review in %v, want a short (~10 min) step", until)
	}
}

// TestLearningService_GradeAnswer_HardOnNewWordUsesMinuteStep checks Hard
// on a brand-new word (never graduated) also uses a short step, matching
// the user's explicit request for Anki-style "in 15 minutes" behavior.
func TestLearningService_GradeAnswer_HardOnNewWordUsesMinuteStep(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()
	id, err := svc.CreateCollection(ctx, 1, "Coll")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, _, err := svc.AddWordsFromText(ctx, 1, id, "a - a1"); err != nil {
		t.Fatalf("add words: %v", err)
	}
	var wordID int64
	for wid := range repo.words {
		wordID = wid
	}

	nextReviewAt, _, err := svc.GradeAnswer(ctx, 1, wordID, models.LearningGradeHard)
	if err != nil {
		t.Fatalf("grade answer: %v", err)
	}
	until := time.Until(nextReviewAt)
	if until <= 0 || until > 20*time.Minute {
		t.Fatalf("next review in %v, want a short (~15 min) step for Hard on a new word", until)
	}
}

// TestLearningService_GradeAnswer_HardOnGraduatedWordUsesDays checks Hard
// falls back to the normal day-based growth once a word has already
// graduated to daily reviews (repetitions > 0) — the minute-level step is
// only for still-new words.
func TestLearningService_GradeAnswer_HardOnGraduatedWordUsesDays(t *testing.T) {
	repo, svc := newTestLearningService()
	ctx := context.Background()
	id, err := svc.CreateCollection(ctx, 1, "Coll")
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, _, err := svc.AddWordsFromText(ctx, 1, id, "a - a1"); err != nil {
		t.Fatalf("add words: %v", err)
	}
	var wordID int64
	for wid := range repo.words {
		wordID = wid
	}
	// Graduate it once with Good first (repetitions 0 -> 1).
	if _, _, err := svc.GradeAnswer(ctx, 1, wordID, models.LearningGradeGood); err != nil {
		t.Fatalf("first grade: %v", err)
	}

	nextReviewAt, _, err := svc.GradeAnswer(ctx, 1, wordID, models.LearningGradeHard)
	if err != nil {
		t.Fatalf("second grade: %v", err)
	}
	until := time.Until(nextReviewAt)
	if until < 20*time.Hour {
		t.Fatalf("next review in %v, want at least ~1 day (word already graduated)", until)
	}
}
