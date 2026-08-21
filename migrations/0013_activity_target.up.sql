-- Per-activity daily time target, replacing the hardcoded 120-minute
-- constant the Track main screen's progress bar used to compare against.
-- NULL means "not configured yet" — callers fall back to the old default.
ALTER TABLE activities ADD COLUMN target_minutes INTEGER NULL
    CONSTRAINT chk_activity_target_minutes_range
    CHECK (target_minutes IS NULL OR (target_minutes > 0 AND target_minutes <= 1440));
