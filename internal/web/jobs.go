package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// AI calls take tens of seconds, which is longer than a phone will hold an HTTP
// request open: backgrounding a Mini App drops the connection and the answer
// with it. So the client starts a job, gets an id, and polls.
//
// Jobs live in memory. A restart loses them, which costs a retry and no data —
// the alternative is a table and a sweeper for something whose whole lifetime is
// under two minutes.

type jobStatus string

const (
	jobPending jobStatus = "pending"
	jobDone    jobStatus = "done"
	jobFailed  jobStatus = "failed"
)

const (
	// jobTTL is how long a finished job stays readable. Long enough for a
	// client that was backgrounded mid-call to come back and collect it.
	jobTTL = 10 * time.Minute
	// jobRunTimeout bounds the work itself, above the service's own per-call
	// ceilings so theirs is the one that speaks first.
	jobRunTimeout = 3 * time.Minute
)

type job struct {
	id      string
	kind    string
	ownerID int64
	status  jobStatus
	result  any
	// errCode is client-facing; errMessage is what the user reads. The real
	// error goes to the log and no further.
	errCode    string
	errMessage string
	expiresAt  time.Time
}

// jobs is a set of running and recently finished jobs.
//
// Ownership is checked on read rather than trusted to unguessable ids: an id
// that leaks into a log or a Referer must not be enough to read someone's
// generated plan.
type jobs struct {
	// ctx is the application context, so a job outlives the request that
	// started it but not the process.
	ctx context.Context
	mu  sync.Mutex
	// items is keyed by job id.
	items map[string]*job
	now   func() time.Time
}

func newJobs(ctx context.Context) *jobs {
	return &jobs{ctx: ctx, items: make(map[string]*job), now: time.Now}
}

// start runs fn in the background and returns a snapshot of the job to poll.
//
// A snapshot, not the pointer: the goroutine below writes to the entry as soon
// as the work finishes, and handing the caller the same pointer to read from is
// a data race — one the race detector duly found.
func (j *jobs) start(kind string, ownerID int64, fn func(context.Context) (any, error)) (job, error) {
	id, err := newJobID()
	if err != nil {
		return job{}, err
	}

	entry := &job{
		id:        id,
		kind:      kind,
		ownerID:   ownerID,
		status:    jobPending,
		expiresAt: j.now().Add(jobTTL),
	}

	j.mu.Lock()
	j.sweepLocked()
	j.items[id] = entry
	snapshot := *entry
	j.mu.Unlock()

	go func() {
		// Detached from the request on purpose: the point of a job is that the
		// work survives the client going away.
		ctx, cancel := context.WithTimeout(j.ctx, jobRunTimeout)
		defer cancel()

		result, err := fn(ctx)

		j.mu.Lock()
		defer j.mu.Unlock()
		entry.expiresAt = j.now().Add(jobTTL)
		if err != nil {
			entry.status = jobFailed
			entry.errCode, entry.errMessage = aiErrorFor(err)
			log.Warn().Str("job", id).Str("kind", kind).Err(err).Msg("web ai job failed")
			return
		}
		entry.status = jobDone
		entry.result = result
	}()

	return snapshot, nil
}

// get returns a snapshot of the job, if it exists, has not expired and belongs
// to this user.
func (j *jobs) get(id string, ownerID int64) (job, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()

	entry, ok := j.items[id]
	if !ok || entry.ownerID != ownerID || j.now().After(entry.expiresAt) {
		return job{}, false
	}
	// A copy: the caller reads it outside the lock while the goroutine may
	// still be writing to the original.
	return *entry, true
}

// sweepLocked drops expired jobs. Called on each start rather than from a
// ticker: the map only grows when someone is starting jobs.
func (j *jobs) sweepLocked() {
	now := j.now()
	for id, entry := range j.items {
		if now.After(entry.expiresAt) {
			delete(j.items, id)
		}
	}
}

func newJobID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
