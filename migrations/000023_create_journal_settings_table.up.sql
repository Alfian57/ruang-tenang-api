-- Journal settings table
CREATE TABLE journal_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    allow_ai_access BOOLEAN NOT NULL DEFAULT FALSE,
    is_blocked BOOLEAN DEFAULT FALSE,
    ai_context_days INTEGER DEFAULT 7,
    ai_context_max_entries INTEGER DEFAULT 5,
    default_share_with_ai BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_journal_settings_user_id ON journal_settings(user_id);
