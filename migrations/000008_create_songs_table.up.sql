-- Songs table
CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(300) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    thumbnail VARCHAR(500),
    song_category_id INTEGER NOT NULL REFERENCES song_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_songs_title ON songs(title);
CREATE INDEX idx_songs_song_category_id ON songs(song_category_id);
CREATE INDEX idx_songs_deleted_at ON songs(deleted_at);
CREATE UNIQUE INDEX idx_songs_slug ON songs(slug) WHERE deleted_at IS NULL;
