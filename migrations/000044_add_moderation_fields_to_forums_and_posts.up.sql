-- Add moderation fields to forums table
ALTER TABLE forums 
ADD COLUMN IF NOT EXISTS trigger_warnings JSONB,
ADD COLUMN IF NOT EXISTS is_flagged BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS flagged_reason TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS has_accepted_answer BOOLEAN DEFAULT FALSE;

-- Add moderation and voting fields to forum_posts table
ALTER TABLE forum_posts 
ADD COLUMN IF NOT EXISTS is_flagged BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS flagged_reason TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS is_accepted_answer BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS is_community_favorite BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS upvotes_count INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS downvotes_count INT DEFAULT 0;
