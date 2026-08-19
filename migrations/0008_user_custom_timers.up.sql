-- хранит пользовательские интервалы таймера (кроме встроенных 15/30 мин),
-- которые пользователь добавляет вручную и может позже удалить.
CREATE TABLE IF NOT EXISTS user_custom_timers (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interval_min INTEGER     NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_custom_interval_min_range CHECK (interval_min > 0 AND interval_min <= 360),
    CONSTRAINT uq_user_custom_interval UNIQUE (user_id, interval_min)
);

CREATE INDEX IF NOT EXISTS idx_user_custom_timers_user
    ON user_custom_timers(user_id);
