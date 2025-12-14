CREATE TYPE mood_type AS ENUM ('happy', 'neutral', 'angry', 'disappointed', 'sad', 'crying');

CREATE TABLE user_moods (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mood mood_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_moods_user ON user_moods(user_id);
CREATE INDEX idx_user_moods_created_at ON user_moods(created_at);
