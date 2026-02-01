-- User moods table
CREATE TABLE user_moods (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mood VARCHAR(50) NOT NULL,
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_user_moods_user_id ON user_moods(user_id);
CREATE INDEX idx_user_moods_mood ON user_moods(mood);
CREATE INDEX idx_user_moods_created_at ON user_moods(created_at);
CREATE INDEX idx_user_moods_deleted_at ON user_moods(deleted_at);
