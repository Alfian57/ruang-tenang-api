-- Add is_favorite and is_trash to chat_sessions
ALTER TABLE chat_sessions
ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS is_trash BOOLEAN DEFAULT FALSE;

-- Add is_liked and is_disliked to chat_messages
ALTER TABLE chat_messages
ADD COLUMN IF NOT EXISTS is_liked BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS is_disliked BOOLEAN DEFAULT FALSE;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_chat_sessions_is_favorite ON chat_sessions(is_favorite) WHERE is_favorite = true;
CREATE INDEX IF NOT EXISTS idx_chat_sessions_is_trash ON chat_sessions(is_trash) WHERE is_trash = true;
