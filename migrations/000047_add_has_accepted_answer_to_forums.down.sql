DROP INDEX IF EXISTS idx_forums_has_accepted;

ALTER TABLE forums
DROP COLUMN IF EXISTS has_accepted_answer;
