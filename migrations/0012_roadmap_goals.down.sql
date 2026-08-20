DROP INDEX IF EXISTS idx_roadmap_cards_pending;

CREATE INDEX IF NOT EXISTS idx_roadmap_cards_pending
    ON roadmap_cards(roadmap_id, created_at, id)
    WHERE is_done = FALSE;

ALTER TABLE roadmap_cards DROP CONSTRAINT IF EXISTS chk_roadmap_card_kind;
ALTER TABLE roadmap_cards DROP COLUMN IF EXISTS kind;
ALTER TABLE roadmap_cards DROP CONSTRAINT IF EXISTS chk_roadmap_card_difficulty;
ALTER TABLE roadmap_cards DROP COLUMN IF EXISTS difficulty;

ALTER TABLE roadmaps RENAME COLUMN mastery_criteria TO goal;

DROP INDEX IF EXISTS idx_roadmaps_goal;
ALTER TABLE roadmaps DROP COLUMN IF EXISTS goal_id;

DROP TABLE IF EXISTS roadmap_goals;
