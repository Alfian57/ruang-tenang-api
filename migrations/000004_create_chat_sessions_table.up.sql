-- Chat sessions table with folder and summary fields consolidated
CREATE TABLE chat_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    
    -- Folder organization
    folder_id INTEGER,
    
    -- AI summary
    summary TEXT,
    summary_generated_at TIMESTAMP WITH TIME ZONE,
    
    is_favorite BOOLEAN DEFAULT FALSE,
    is_trash BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_folder ON chat_sessions(folder_id);
CREATE INDEX idx_chat_sessions_deleted_at ON chat_sessions(deleted_at);
CREATE INDEX idx_chat_sessions_is_favorite ON chat_sessions(is_favorite) WHERE is_favorite = true;
CREATE INDEX idx_chat_sessions_is_trash ON chat_sessions(is_trash) WHERE is_trash = true;
