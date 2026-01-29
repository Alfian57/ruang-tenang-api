-- Drop indexes
DROP INDEX IF EXISTS idx_playlist_items_playlist_id;
DROP INDEX IF EXISTS idx_playlist_items_song_id;
DROP INDEX IF EXISTS idx_playlist_items_position;
DROP INDEX IF EXISTS idx_playlist_items_deleted_at;

-- Drop playlist_items table
DROP TABLE IF EXISTS playlist_items;
