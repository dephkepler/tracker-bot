package dispatcher

import (
	"testing"
	"time"

	"tracker-bot/internal/models"
)

func TestPendingQuizzesTakeOnce(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	p := newPendingQuizzes()
	p.put(7, models.RoadmapQuiz{CardID: 42, Question: "Q?"}, now)

	quiz, ok := p.take(7, now)
	if !ok || quiz.CardID != 42 {
		t.Fatalf("take = %+v, %v; want the stored quiz", quiz, ok)
	}
	// A question answers once: a second message must fall through to normal
	// routing rather than be graded again.
	if _, ok := p.take(7, now); ok {
		t.Fatal("second take must report no pending quiz")
	}
}

func TestPendingQuizzesExpires(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	p := newPendingQuizzes()
	p.put(7, models.RoadmapQuiz{CardID: 42}, now)

	if _, ok := p.take(7, now.Add(quizTTL+time.Second)); ok {
		t.Fatal("an expired question must not claim the message")
	}
	// Expiry also clears the entry, so it cannot come back.
	if _, ok := p.take(7, now); ok {
		t.Fatal("expired entry was not dropped")
	}
}

func TestPendingQuizzesLatestQuestionWins(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	p := newPendingQuizzes()
	p.put(7, models.RoadmapQuiz{CardID: 1}, now)
	p.put(7, models.RoadmapQuiz{CardID: 2}, now.Add(time.Minute))

	quiz, ok := p.take(7, now.Add(2*time.Minute))
	if !ok || quiz.CardID != 2 {
		t.Fatalf("take = %+v, %v; want the second question", quiz, ok)
	}
}

func TestPendingQuizzesPerUser(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	p := newPendingQuizzes()
	p.put(7, models.RoadmapQuiz{CardID: 1}, now)
	p.put(8, models.RoadmapQuiz{CardID: 2}, now)

	if _, ok := p.take(7, now); !ok {
		t.Fatal("user 7 lost their question")
	}
	quiz, ok := p.take(8, now)
	if !ok || quiz.CardID != 2 {
		t.Fatalf("user 8 take = %+v, %v", quiz, ok)
	}
}

func TestPendingQuizzesDrop(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	p := newPendingQuizzes()
	p.put(7, models.RoadmapQuiz{CardID: 1}, now)
	p.drop(7)

	if _, ok := p.take(7, now); ok {
		t.Fatal("drop did not discard the question")
	}
}
