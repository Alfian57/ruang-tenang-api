-- Add moderation fields to forums table
ALTER TABLE forums
ADD COLUMN trigger_warnings JSON AFTER content,
ADD COLUMN is_flagged BOOLEAN NOT NULL DEFAULT FALSE AFTER trigger_warnings,
ADD COLUMN flagged_reason TEXT AFTER is_flagged;

-- Add moderation fields to forum_posts table
ALTER TABLE forum_posts
ADD COLUMN is_flagged BOOLEAN NOT NULL DEFAULT FALSE AFTER content,
ADD COLUMN flagged_reason TEXT AFTER is_flagged;

-- Add indexes
ALTER TABLE forums ADD INDEX idx_is_flagged (is_flagged);
ALTER TABLE forum_posts ADD INDEX idx_is_flagged (is_flagged);
