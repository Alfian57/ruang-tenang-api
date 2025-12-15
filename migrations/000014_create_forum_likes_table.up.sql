CREATE TABLE IF NOT EXISTS forum_likes (
    id SERIAL PRIMARY KEY,
    forum_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_forum_likes_forum FOREIGN KEY (forum_id) REFERENCES forums(id) ON DELETE CASCADE,
    CONSTRAINT fk_forum_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_forum_user_like UNIQUE (forum_id, user_id)
);

CREATE INDEX idx_forum_likes_forum_id ON forum_likes(forum_id);
CREATE INDEX idx_forum_likes_user_id ON forum_likes(user_id);
