-- Add is_pinned field to chat_messages for pinning important AI responses
ALTER TABLE chat_messages 
ADD COLUMN is_pinned BOOLEAN DEFAULT FALSE;

-- Add type column if not exists (for backward compatibility)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'chat_messages' AND column_name = 'type'
    ) THEN
        ALTER TABLE chat_messages ADD COLUMN type VARCHAR(20) DEFAULT 'text';
    END IF;
END $$;

CREATE INDEX idx_chat_messages_pinned ON chat_messages(chat_session_id, is_pinned) WHERE is_pinned = true;
