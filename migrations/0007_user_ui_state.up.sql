-- Navigation state (which "screen" a user is currently on) used to live only
-- in the bot process's memory (internal/dispatcher/dispatcher.go, userScreen
-- map). Every deploy/restart wiped it for every user mid-flow, so screen-gated
-- buttons (Activate, Timer 15/30 min, Today/Period reports, etc.) silently
-- stopped responding ("Use buttons from menu") until the user sent /start
-- again. Persisting it here lets the bot restore the right screen on restart.
CREATE TABLE IF NOT EXISTS user_ui_state (
    user_id    BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    screen     TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
