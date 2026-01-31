-- Forum posts table with all fields consolidated
CREATE TABLE forum_posts (
    id SERIAL PRIMARY KEY,
    forum_id INTEGER NOT NULL REFERENCES forums(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    
    -- Moderation fields
    is_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    flagged_reason TEXT,
    
    -- Best answer and voting
    is_accepted_answer BOOLEAN DEFAULT FALSE,
    is_community_favorite BOOLEAN DEFAULT FALSE,
    upvotes_count INTEGER DEFAULT 0,
    downvotes_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_forum_posts_forum_id ON forum_posts(forum_id);
CREATE INDEX idx_forum_posts_user_id ON forum_posts(user_id);
CREATE INDEX idx_forum_posts_is_flagged ON forum_posts(is_flagged);
CREATE INDEX idx_forum_posts_accepted ON forum_posts(forum_id, is_accepted_answer) WHERE is_accepted_answer = TRUE;
CREATE INDEX idx_forum_posts_community_fav ON forum_posts(forum_id, is_community_favorite) WHERE is_community_favorite = TRUE;
CREATE INDEX idx_forum_posts_upvotes ON forum_posts(forum_id, upvotes_count DESC);
CREATE INDEX idx_forum_posts_deleted_at ON forum_posts(deleted_at);
