-- Remove added columns
ALTER TABLE chat_sessions
DROP COLUMN IF EXISTS is_favorite,
DROP COLUMN IF EXISTS is_trash;

ALTER TABLE chat_messages
DROP COLUMN IF EXISTS is_liked,
DROP COLUMN IF EXISTS is_disliked;
