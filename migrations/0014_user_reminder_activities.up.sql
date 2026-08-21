-- Reminder membership, separate from user_selected_activities (which is
-- also reused as a one-off batch-select scratchpad for Archive/Delete
-- selected). Activating reminders now ADDS the currently-checked
-- activities here and clears the checkboxes, instead of the checkboxes
-- themselves being live reminder truth — see the comment on
-- TrackerService.AddSelectedToReminders.
CREATE TABLE IF NOT EXISTS user_reminder_activities(
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_id BIGINT      NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, activity_id)
);

CREATE INDEX IF NOT EXISTS idx_user_reminder_activities_user
    ON user_reminder_activities(user_id);
