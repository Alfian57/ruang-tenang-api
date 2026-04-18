-- Chat folders table for organizing chat sessions
CREATE TABLE chat_folders (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
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
CREATE UNIQUE INDEX idx_chat_folders_uuid ON chat_folders(uuid);
