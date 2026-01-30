-- Forum likes table
CREATE TABLE forum_likes (
    id SERIAL PRIMARY KEY,
    forum_id INTEGER NOT NULL REFERENCES forums(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(forum_id, user_id)
);

CREATE INDEX idx_forum_likes_forum_id ON forum_likes(forum_id);
CREATE INDEX idx_forum_likes_user_id ON forum_likes(user_id);
