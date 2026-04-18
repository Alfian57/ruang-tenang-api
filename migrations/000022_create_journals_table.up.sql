-- Journals table
CREATE TABLE journals (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    content TEXT NOT NULL,
    summary TEXT,
    mood_id INTEGER REFERENCES user_moods(id) ON DELETE SET NULL,
    tags TEXT[],
    is_private BOOLEAN NOT NULL DEFAULT TRUE,
    share_with_ai BOOLEAN NOT NULL DEFAULT FALSE,
    ai_accessed_at TIMESTAMP,
    word_count INTEGER DEFAULT 0,
    sentiment_score DECIMAL(3,2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_journals_user_id ON journals(user_id);
CREATE INDEX idx_journals_user_id_share_with_ai ON journals(user_id, share_with_ai) WHERE share_with_ai = TRUE;
CREATE INDEX idx_journals_user_id_created_at ON journals(user_id, created_at DESC);
CREATE INDEX idx_journals_tags ON journals USING GIN(tags);
CREATE UNIQUE INDEX idx_journals_uuid ON journals(uuid);
