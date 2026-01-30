-- Journal entries table for private journaling with AI integration
CREATE TABLE journals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    content TEXT NOT NULL,
    mood_id INTEGER REFERENCES user_moods(id) ON DELETE SET NULL,
    tags TEXT[], -- PostgreSQL array for tags
    is_private BOOLEAN NOT NULL DEFAULT TRUE, -- Privacy by default
    share_with_ai BOOLEAN NOT NULL DEFAULT FALSE, -- Explicit opt-in for AI access
    ai_accessed_at TIMESTAMP, -- Track when AI last accessed this entry
    word_count INTEGER DEFAULT 0,
    sentiment_score DECIMAL(3,2), -- -1.00 to 1.00 (negative to positive)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_journals_user_id ON journals(user_id);
CREATE INDEX idx_journals_user_id_share_with_ai ON journals(user_id, share_with_ai) WHERE share_with_ai = TRUE;
CREATE INDEX idx_journals_user_id_created_at ON journals(user_id, created_at DESC);
CREATE INDEX idx_journals_tags ON journals USING GIN(tags);

-- Journal settings table for global user preferences
CREATE TABLE journal_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    allow_ai_access BOOLEAN NOT NULL DEFAULT FALSE, -- Global toggle for AI access
    ai_context_days INTEGER DEFAULT 7, -- How many days of entries to share with AI
    ai_context_max_entries INTEGER DEFAULT 5, -- Max number of entries to share
    default_share_with_ai BOOLEAN DEFAULT FALSE, -- Default setting for new entries
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_journal_settings_user_id ON journal_settings(user_id);

-- Journal AI access log for transparency
CREATE TABLE journal_ai_access_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    journal_id INTEGER NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
    chat_session_id INTEGER REFERENCES chat_sessions(id) ON DELETE SET NULL,
    accessed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    context_type VARCHAR(50) -- 'full', 'summary', 'keyword_match'
);

CREATE INDEX idx_journal_ai_access_logs_user_id ON journal_ai_access_logs(user_id);
CREATE INDEX idx_journal_ai_access_logs_journal_id ON journal_ai_access_logs(journal_id);
