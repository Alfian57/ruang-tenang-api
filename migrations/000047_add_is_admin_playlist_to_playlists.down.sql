-- Remove is_admin_playlist column from playlists table
DROP INDEX IF EXISTS idx_playlists_is_admin;
ALTER TABLE playlists DROP COLUMN IF EXISTS is_admin_playlist;
