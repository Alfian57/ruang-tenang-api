-- Drop indexes
DROP INDEX IF EXISTS idx_playlists_user_id;
DROP INDEX IF EXISTS idx_playlists_deleted_at;

-- Drop playlists table
DROP TABLE IF EXISTS playlists;
