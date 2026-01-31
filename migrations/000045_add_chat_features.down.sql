-- Remove index and constraint
DROP INDEX IF EXISTS idx_chat_sessions_folder_id;
ALTER TABLE chat_sessions DROP CONSTRAINT IF EXISTS fk_chat_sessions_folder;

-- Remove folder and summary fields from chat_sessions
ALTER TABLE chat_sessions 
DROP COLUMN IF EXISTS folder_id,
DROP COLUMN IF EXISTS summary,
DROP COLUMN IF EXISTS summary_generated_at;

-- Remove is_pinned from chat_messages
ALTER TABLE chat_messages 
DROP COLUMN IF EXISTS is_pinned;

-- Drop folders indexes and table
DROP INDEX IF EXISTS idx_chat_folders_position;
DROP INDEX IF EXISTS idx_chat_folders_user;
DROP TABLE IF EXISTS chat_folders;
