-- Drop unique indexes
DROP INDEX IF EXISTS idx_articles_slug;
DROP INDEX IF EXISTS idx_article_categories_slug;
DROP INDEX IF EXISTS idx_forums_slug;
DROP INDEX IF EXISTS idx_forum_categories_slug;
DROP INDEX IF EXISTS idx_songs_slug;
DROP INDEX IF EXISTS idx_song_categories_slug;
DROP INDEX IF EXISTS idx_journals_uuid;
DROP INDEX IF EXISTS idx_chat_sessions_uuid;
DROP INDEX IF EXISTS idx_chat_messages_uuid;
DROP INDEX IF EXISTS idx_chat_folders_uuid;
DROP INDEX IF EXISTS idx_playlists_uuid;
DROP INDEX IF EXISTS idx_playlist_items_uuid;
DROP INDEX IF EXISTS idx_users_username;

-- Drop columns
ALTER TABLE articles DROP COLUMN IF EXISTS slug;
ALTER TABLE article_categories DROP COLUMN IF EXISTS slug;
ALTER TABLE forums DROP COLUMN IF EXISTS slug;
ALTER TABLE forum_categories DROP COLUMN IF EXISTS slug;
ALTER TABLE songs DROP COLUMN IF EXISTS slug;
ALTER TABLE song_categories DROP COLUMN IF EXISTS slug;
ALTER TABLE journals DROP COLUMN IF EXISTS uuid;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS uuid;
ALTER TABLE chat_messages DROP COLUMN IF EXISTS uuid;
ALTER TABLE chat_folders DROP COLUMN IF EXISTS uuid;
ALTER TABLE playlists DROP COLUMN IF EXISTS uuid;
ALTER TABLE playlist_items DROP COLUMN IF EXISTS uuid;
ALTER TABLE users DROP COLUMN IF EXISTS username;
