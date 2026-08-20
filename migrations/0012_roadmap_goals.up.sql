-- Roadmap v2: a third level above technologies.
--
-- v1 modelled "a technology with a checklist". What the feature is actually
-- for is bigger: an outcome the user is working toward ("reach mid-level"),
-- which several technologies feed into. So goals become their own entity,
-- technologies hang off a goal, and progress is reported against the goal.
--
-- Cards also gain a difficulty and a kind, which is what lets the checklist
-- be walked easiest-first instead of in paste order.

CREATE TABLE IF NOT EXISTS roadmap_goals (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    is_archived BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_roadmap_goal_name UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_roadmap_goals_user
    ON roadmap_goals(user_id)
    WHERE is_archived = FALSE;

-- ON DELETE SET NULL, not CASCADE: dropping a goal must not silently take a
-- pile of technologies and their cards with it. Unattached technologies stay
-- visible in their own "no goal" group instead.
ALTER TABLE roadmaps
    ADD COLUMN IF NOT EXISTS goal_id BIGINT NULL REFERENCES roadmap_goals(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_roadmaps_goal
    ON roadmaps(goal_id)
    WHERE is_archived = FALSE;

-- v1's roadmaps.goal was free text meaning "what does knowing this
-- technology mean for me" — with real goals now being their own table, that
-- name would be two different things at once. Renamed, data preserved.
ALTER TABLE roadmaps RENAME COLUMN goal TO mastery_criteria;

-- 1 = easy, 2 = medium, 3 = hard. Medium is the default so a plain pasted
-- line lands in the middle rather than pretending to be graded.
ALTER TABLE roadmap_cards
    ADD COLUMN IF NOT EXISTS difficulty SMALLINT NOT NULL DEFAULT 2;

ALTER TABLE roadmap_cards
    ADD CONSTRAINT chk_roadmap_card_difficulty CHECK (difficulty BETWEEN 1 AND 3);

ALTER TABLE roadmap_cards
    ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'topic';

ALTER TABLE roadmap_cards
    ADD CONSTRAINT chk_roadmap_card_kind CHECK (kind IN ('topic', 'article', 'book', 'lecture'));

-- Replaces the v1 index: the digest and the checklist now walk pending cards
-- easiest-first, so difficulty leads the sort. id still tiebreaks, since a
-- bulk paste shares one created_at (see repo.PickDigestCards).
DROP INDEX IF EXISTS idx_roadmap_cards_pending;

CREATE INDEX IF NOT EXISTS idx_roadmap_cards_pending
    ON roadmap_cards(roadmap_id, difficulty, created_at, id)
    WHERE is_done = FALSE;
