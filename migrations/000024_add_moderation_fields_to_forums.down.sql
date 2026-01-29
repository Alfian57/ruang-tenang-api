ALTER TABLE forums
DROP INDEX idx_is_flagged,
DROP COLUMN trigger_warnings,
DROP COLUMN is_flagged,
DROP COLUMN flagged_reason;

ALTER TABLE forum_posts
DROP INDEX idx_is_flagged,
DROP COLUMN is_flagged,
DROP COLUMN flagged_reason;
