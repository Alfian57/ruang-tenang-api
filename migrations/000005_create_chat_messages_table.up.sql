-- Chat messages table with pinned and type fields consolidated
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    chat_session_id INTEGER NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- 'user', 'ai'
    content TEXT NOT NULL,
    
    -- Additional fields
    is_pinned BOOLEAN DEFAULT FALSE,
    type VARCHAR(20) DEFAULT 'text', -- 'text', 'voice', etc.
    is_liked BOOLEAN DEFAULT FALSE,
    is_disliked BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_messages_chat_session_id ON chat_messages(chat_session_id);
CREATE INDEX idx_chat_messages_role ON chat_messages(role);
CREATE INDEX idx_chat_messages_pinned ON chat_messages(chat_session_id, is_pinned) WHERE is_pinned = true;
