-- Chat folders table for organizing chat sessions
CREATE TABLE chat_folders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366f1',
    icon VARCHAR(50) DEFAULT 'folder',
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_folders_user ON chat_folders(user_id);
CREATE INDEX idx_chat_folders_position ON chat_folders(user_id, position);

-- Add foreign key constraint for chat_sessions.folder_id
ALTER TABLE chat_sessions 
ADD CONSTRAINT fk_chat_sessions_folder 
FOREIGN KEY (folder_id) REFERENCES chat_folders(id) ON DELETE SET NULL;
