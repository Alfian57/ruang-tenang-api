-- Forum post votes table
CREATE TABLE forum_post_votes (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type VARCHAR(10) NOT NULL DEFAULT 'upvote',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_post_user_vote UNIQUE (post_id, user_id),
    CONSTRAINT chk_vote_type CHECK (vote_type IN ('upvote', 'downvote'))
);

CREATE INDEX idx_forum_post_votes_post_id ON forum_post_votes(post_id);
CREATE INDEX idx_forum_post_votes_user_id ON forum_post_votes(user_id);
CREATE INDEX idx_forum_post_votes_type ON forum_post_votes(vote_type);
