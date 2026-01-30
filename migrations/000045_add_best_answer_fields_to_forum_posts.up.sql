-- Add Best Answer and Vote Count fields to forum_posts table
-- is_accepted_answer: OP's choice for the best answer
-- is_community_favorite: Automatically set when post has highest votes
-- upvotes_count: Cached count for performance (denormalized)
-- downvotes_count: For future downvote support (if needed)

ALTER TABLE forum_posts
ADD COLUMN is_accepted_answer BOOLEAN DEFAULT FALSE,
ADD COLUMN is_community_favorite BOOLEAN DEFAULT FALSE,
ADD COLUMN upvotes_count INTEGER DEFAULT 0,
ADD COLUMN downvotes_count INTEGER DEFAULT 0;

-- Index for efficient sorting and querying
CREATE INDEX idx_forum_posts_accepted ON forum_posts(forum_id, is_accepted_answer) WHERE is_accepted_answer = TRUE;
CREATE INDEX idx_forum_posts_community_fav ON forum_posts(forum_id, is_community_favorite) WHERE is_community_favorite = TRUE;
CREATE INDEX idx_forum_posts_upvotes ON forum_posts(forum_id, upvotes_count DESC);
