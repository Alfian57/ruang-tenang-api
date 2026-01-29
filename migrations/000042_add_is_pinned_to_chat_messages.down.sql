DROP INDEX IF EXISTS idx_chat_messages_pinned;

ALTER TABLE chat_messages DROP COLUMN IF EXISTS is_pinned;
