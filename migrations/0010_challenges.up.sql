-- "Challenges" feature: a user-defined day-range plan (e.g. "100 days of
-- reading") with one square per day, marked done/skipped, plus a daily
-- evening push nudging the user to mark today.

CREATE TABLE IF NOT EXISTS challenges (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    start_date   DATE        NOT NULL,
    end_date     DATE        NOT NULL,
    is_archived  BOOLEAN     NOT NULL DEFAULT FALSE,
    -- Daily evening push schedule — same shape as user_timer_settings'
    -- next_ping_at, just one absolute UTC instant computed from the user's
    -- own timezone (see internal/scheduler/challenge.go).
    next_push_at TIMESTAMPTZ NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_challenge_name UNIQUE (user_id, name),
    CONSTRAINT chk_challenge_date_range CHECK (end_date >= start_date),
    -- Keeps the day-grid renderable as inline-keyboard buttons in one message.
    CONSTRAINT chk_challenge_max_length CHECK (end_date - start_date < 100)
);

CREATE INDEX IF NOT EXISTS idx_challenges_user
    ON challenges(user_id)
    WHERE is_archived = FALSE;

CREATE INDEX IF NOT EXISTS idx_challenges_push_due
    ON challenges(next_push_at)
    WHERE is_archived = FALSE AND next_push_at IS NOT NULL;

-- One row per day in the plan's range, pre-populated when the challenge is
-- created so the grid and progress math never need to compute "which days
-- exist" on the fly.
CREATE TABLE IF NOT EXISTS challenge_days (
    id           BIGSERIAL PRIMARY KEY,
    challenge_id BIGINT      NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    day_date     DATE        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending',
    marked_at    TIMESTAMPTZ NULL,

    CONSTRAINT uq_challenge_day UNIQUE (challenge_id, day_date),
    CONSTRAINT chk_challenge_day_status CHECK (status IN ('pending', 'done', 'skipped'))
);

CREATE INDEX IF NOT EXISTS idx_challenge_days_challenge
    ON challenge_days(challenge_id);
