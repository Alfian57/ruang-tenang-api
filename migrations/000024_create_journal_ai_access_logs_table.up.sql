-- Journal AI access logs table
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
