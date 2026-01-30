ALTER TABLE chat_sessions DROP CONSTRAINT IF EXISTS fk_chat_sessions_folder;
DROP TABLE IF EXISTS chat_folders CASCADE;
