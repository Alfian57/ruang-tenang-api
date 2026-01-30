-- Forum Post Votes Table
-- Supports upvote system for forum posts (comments/answers)
-- Using upvote-only approach (recommended for mental health community)
CREATE TABLE IF NOT EXISTS forum_post_votes (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    vote_type VARCHAR(10) NOT NULL DEFAULT 'upvote', -- 'upvote' or 'downvote' (for future use)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_post_votes_post FOREIGN KEY (post_id) REFERENCES forum_posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_votes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_post_user_vote UNIQUE (post_id, user_id),
    CONSTRAINT chk_vote_type CHECK (vote_type IN ('upvote', 'downvote'))
);

CREATE INDEX idx_forum_post_votes_post_id ON forum_post_votes(post_id);
CREATE INDEX idx_forum_post_votes_user_id ON forum_post_votes(user_id);
CREATE INDEX idx_forum_post_votes_type ON forum_post_votes(vote_type);
