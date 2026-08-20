-- "Roadmap" feature: up to 5 technologies the user is learning, each with a
-- free-text mastery goal and a checklist of freeform cards (topics, tasks,
-- articles). A periodic push delivers a digest of what's still pending —
-- same interval-in-minutes schedule shape as user_learning_settings
-- (migrations/0009) rather than a fixed time of day.

CREATE TABLE IF NOT EXISTS roadmaps (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    goal        TEXT        NOT NULL DEFAULT '',
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,  -- included in the push digest
    is_archived BOOLEAN     NOT NULL DEFAULT FALSE, -- archiving is what frees a slot in the 5-cap
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_roadmap_name UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_roadmaps_user
    ON roadmaps(user_id)
    WHERE is_archived = FALSE;

-- One row per checklist card. Freeform text, no kind/type field — a card is
-- whatever the user typed on that line.
CREATE TABLE IF NOT EXISTS roadmap_cards (
    id         BIGSERIAL PRIMARY KEY,
    roadmap_id BIGINT      NOT NULL REFERENCES roadmaps(id) ON DELETE CASCADE,
    text       TEXT        NOT NULL,
    is_done    BOOLEAN     NOT NULL DEFAULT FALSE,
    done_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_roadmap_cards_roadmap
    ON roadmap_cards(roadmap_id);

-- Speeds up the digest query ("oldest pending cards per roadmap"). Includes
-- id because a bulk paste inserts every card in one transaction and so with
-- one identical created_at — id is the tiebreaker that keeps the digest's
-- ordering deterministic (see repo.PickDigestCards).
CREATE INDEX IF NOT EXISTS idx_roadmap_cards_pending
    ON roadmap_cards(roadmap_id, created_at, id)
    WHERE is_done = FALSE;

-- Push-delivery schedule, one row per user — same shape as
-- user_learning_settings so the scheduler pattern is reused as-is.
CREATE TABLE IF NOT EXISTS user_roadmap_settings (
    user_id      BIGINT      PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    interval_min INTEGER     NOT NULL DEFAULT 180,
    next_push_at TIMESTAMPTZ NULL,
    enabled      BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_roadmap_interval_min_range CHECK (interval_min > 0 AND interval_min <= 1440)
);

CREATE INDEX IF NOT EXISTS idx_roadmap_push_due
    ON user_roadmap_settings(next_push_at)
    WHERE enabled = TRUE;
