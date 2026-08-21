package dispatcher

import (
	"sync"
	"time"

	"tracker-bot/internal/models"
)

// quizTTL is how long an unanswered question keeps claiming the user's next
// message. Without it a question the user walked away from would swallow
// their next unrelated message hours later.
const quizTTL = 15 * time.Minute

// pendingQuizzes holds the questions asked but not yet answered.
//
// This is deliberately not a userSession field: the question is produced by
// a background goroutine (the model call would otherwise freeze the serial
// update loop), and userSession is unsynchronized by design. Being its own
// synchronized store also means an unanswered quiz is state that simply
// expires, rather than a waiting flag someone has to remember to clear.
type pendingQuizzes struct {
	mu    sync.Mutex
	items map[int64]pendingQuiz
}

type pendingQuiz struct {
	quiz    models.RoadmapQuiz
	askedAt time.Time
}

func newPendingQuizzes() *pendingQuizzes {
	return &pendingQuizzes{items: make(map[int64]pendingQuiz)}
}

// put records a question. A second question for the same user replaces the
// first — only the latest one is answerable.
func (p *pendingQuizzes) put(userID int64, quiz models.RoadmapQuiz, now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items[userID] = pendingQuiz{quiz: quiz, askedAt: now}
}

// take removes and returns the pending question, if there is a live one. An
// expired entry is dropped and reported as absent, so the user's message
// falls through to normal routing.
func (p *pendingQuizzes) take(userID int64, now time.Time) (models.RoadmapQuiz, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	item, ok := p.items[userID]
	if !ok {
		return models.RoadmapQuiz{}, false
	}
	delete(p.items, userID)
	if now.Sub(item.askedAt) > quizTTL {
		return models.RoadmapQuiz{}, false
	}
	return item.quiz, true
}

// drop discards any pending question, for Cancel and for leaving the screen.
func (p *pendingQuizzes) drop(userID int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.items, userID)
}
