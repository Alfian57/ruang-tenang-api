DROP INDEX IF EXISTS idx_chat_sessions_folder;

ALTER TABLE chat_sessions DROP COLUMN IF EXISTS summary_generated_at;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS summary;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS folder_id;
