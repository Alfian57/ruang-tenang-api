-- Add moderation fields to articles table
ALTER TABLE articles
ADD COLUMN moderation_status VARCHAR(50) NOT NULL DEFAULT 'pending' AFTER status,
ADD COLUMN moderation_notes TEXT AFTER moderation_status,
ADD COLUMN moderated_by_id BIGINT UNSIGNED AFTER moderation_notes,
ADD COLUMN moderated_at TIMESTAMP NULL AFTER moderated_by_id,
ADD COLUMN trigger_warnings JSON AFTER moderated_at,
ADD COLUMN is_user_generated BOOLEAN NOT NULL DEFAULT FALSE AFTER trigger_warnings,
ADD INDEX idx_moderation_status (moderation_status),
ADD INDEX idx_is_user_generated (is_user_generated);

-- Update existing articles to be approved (assuming they are admin-created)
UPDATE articles SET moderation_status = 'approved', is_user_generated = FALSE;
