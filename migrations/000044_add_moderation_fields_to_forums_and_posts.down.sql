-- Remove moderation fields from forums table
ALTER TABLE forums 
DROP COLUMN IF EXISTS trigger_warnings,
DROP COLUMN IF EXISTS is_flagged,
DROP COLUMN IF EXISTS flagged_reason,
DROP COLUMN IF EXISTS has_accepted_answer;

-- Remove moderation and voting fields from forum_posts table
ALTER TABLE forum_posts 
DROP COLUMN IF EXISTS is_flagged,
DROP COLUMN IF EXISTS flagged_reason,
DROP COLUMN IF EXISTS is_accepted_answer,
DROP COLUMN IF EXISTS is_community_favorite,
DROP COLUMN IF EXISTS upvotes_count,
DROP COLUMN IF EXISTS downvotes_count;
