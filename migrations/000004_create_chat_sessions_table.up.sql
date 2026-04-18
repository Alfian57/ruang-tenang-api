-- Chat sessions table with folder and summary fields consolidated
CREATE TABLE chat_sessions (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    
    -- Folder organization
    folder_id INTEGER,
    
    -- AI summary
    summary TEXT,
    summary_generated_at TIMESTAMP WITH TIME ZONE,

    -- Chat context preferences
    enable_mood_context BOOLEAN NOT NULL DEFAULT TRUE,
    enable_journal_context BOOLEAN NOT NULL DEFAULT FALSE,
    enable_daily_task_context BOOLEAN NOT NULL DEFAULT TRUE,
    enable_xp_level_context BOOLEAN NOT NULL DEFAULT TRUE,
    enable_breathing_context BOOLEAN NOT NULL DEFAULT TRUE,
    enable_playlist_context BOOLEAN NOT NULL DEFAULT FALSE,
    enable_rewards_context BOOLEAN NOT NULL DEFAULT FALSE,
    enable_progress_map_context BOOLEAN NOT NULL DEFAULT FALSE,
    enable_social_context BOOLEAN NOT NULL DEFAULT FALSE,
    session_intent VARCHAR(30) NOT NULL DEFAULT 'general',
    
    is_favorite BOOLEAN DEFAULT FALSE,
    is_trash BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT chk_chat_sessions_session_intent
        CHECK (session_intent IN ('general', 'grounding', 'planning', 'reflection', 'coping'))
);

CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE UNIQUE INDEX idx_chat_sessions_uuid ON chat_sessions(uuid) WHERE deleted_at IS NULL;
CREATE INDEX idx_chat_sessions_folder ON chat_sessions(folder_id);
CREATE INDEX idx_chat_sessions_deleted_at ON chat_sessions(deleted_at);
CREATE INDEX idx_chat_sessions_is_favorite ON chat_sessions(is_favorite) WHERE is_favorite = true;
CREATE INDEX idx_chat_sessions_is_trash ON chat_sessions(is_trash) WHERE is_trash = true;
CREATE INDEX idx_chat_sessions_session_intent ON chat_sessions(session_intent);
