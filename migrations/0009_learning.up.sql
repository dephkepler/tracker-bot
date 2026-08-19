-- Word-learning feature: collections of word pairs reviewed on a spaced-
-- repetition schedule, delivered via periodic pushes (mirrors
-- user_timer_settings' push-scheduling shape, see migrations/0004).

CREATE TABLE IF NOT EXISTS learning_collections (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    is_archived BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE, -- included in the review rotation
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_learning_collection_name UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_learning_collections_user
    ON learning_collections(user_id)
    WHERE is_archived = FALSE;

-- One row per word pair; SRS scheduling state (SM-2-lite) lives directly on
-- the word since each word progresses independently.
CREATE TABLE IF NOT EXISTS learning_words (
    id             BIGSERIAL PRIMARY KEY,
    collection_id  BIGINT      NOT NULL REFERENCES learning_collections(id) ON DELETE CASCADE,
    term           TEXT        NOT NULL,
    translation    TEXT        NOT NULL,
    ease_factor    REAL        NOT NULL DEFAULT 2.5,
    interval_days  INTEGER     NOT NULL DEFAULT 0,
    repetitions    INTEGER     NOT NULL DEFAULT 0,
    next_review_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    learned        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_learning_words_collection
    ON learning_words(collection_id);

-- Speeds up "pick the most overdue due word" queries.
CREATE INDEX IF NOT EXISTS idx_learning_words_due
    ON learning_words(collection_id, next_review_at)
    WHERE learned = FALSE;

-- Push-delivery schedule, one row per user — same shape as
-- user_timer_settings so the scheduler pattern can be reused as-is.
CREATE TABLE IF NOT EXISTS user_learning_settings (
    user_id      BIGINT      PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    interval_min INTEGER     NOT NULL DEFAULT 60,
    next_push_at TIMESTAMPTZ NULL,
    enabled      BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_learning_interval_min_range CHECK (interval_min > 0 AND interval_min <= 1440)
);

CREATE INDEX IF NOT EXISTS idx_learning_push_due
    ON user_learning_settings(next_push_at)
    WHERE enabled = TRUE;

-- Review history, used to compute stats (today/streak) without touching the
-- live SRS state on learning_words.
CREATE TABLE IF NOT EXISTS learning_reviews (
    id          BIGSERIAL PRIMARY KEY,
    word_id     BIGINT      NOT NULL REFERENCES learning_words(id) ON DELETE CASCADE,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    correct     BOOLEAN     NOT NULL,
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_learning_reviews_user_date
    ON learning_reviews(user_id, reviewed_at);
