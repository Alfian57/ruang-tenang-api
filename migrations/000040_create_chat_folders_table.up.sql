-- Create chat folders table for organizing chat sessions (1 level hierarchy only)
CREATE TABLE chat_folders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366f1', -- Hex color code
    icon VARCHAR(50) DEFAULT 'folder',
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_folders_user ON chat_folders(user_id);
CREATE INDEX idx_chat_folders_position ON chat_folders(user_id, position);
