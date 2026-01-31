-- Add is_admin_playlist column to playlists table
ALTER TABLE playlists ADD COLUMN is_admin_playlist BOOLEAN DEFAULT false NOT NULL;

-- Create index for faster queries on public admin playlists
CREATE INDEX idx_playlists_is_admin ON playlists(is_admin_playlist) WHERE is_admin_playlist = true;
