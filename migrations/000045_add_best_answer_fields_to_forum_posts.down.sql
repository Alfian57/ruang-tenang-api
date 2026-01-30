DROP INDEX IF EXISTS idx_forum_posts_upvotes;
DROP INDEX IF EXISTS idx_forum_posts_community_fav;
DROP INDEX IF EXISTS idx_forum_posts_accepted;

ALTER TABLE forum_posts
DROP COLUMN IF EXISTS downvotes_count,
DROP COLUMN IF EXISTS upvotes_count,
DROP COLUMN IF EXISTS is_community_favorite,
DROP COLUMN IF EXISTS is_accepted_answer;
