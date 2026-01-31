-- Create chat_folders table if it doesn't exist
CREATE TABLE IF NOT EXISTS chat_folders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366f1',
    icon VARCHAR(50) DEFAULT 'folder',
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chat_folders_user ON chat_folders(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_folders_position ON chat_folders(user_id, position);

-- Add folder and summary fields to chat_sessions
ALTER TABLE chat_sessions 
ADD COLUMN IF NOT EXISTS folder_id INT,
ADD COLUMN IF NOT EXISTS summary TEXT,
ADD COLUMN IF NOT EXISTS summary_generated_at TIMESTAMP;

-- Add foreign key constraint
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_chat_sessions_folder' AND table_name = 'chat_sessions'
    ) THEN
        ALTER TABLE chat_sessions 
        ADD CONSTRAINT fk_chat_sessions_folder 
        FOREIGN KEY (folder_id) REFERENCES chat_folders(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Create index for folder_id
CREATE INDEX IF NOT EXISTS idx_chat_sessions_folder_id ON chat_sessions(folder_id);

-- Add is_pinned to chat_messages
ALTER TABLE chat_messages 
ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN DEFAULT FALSE;
