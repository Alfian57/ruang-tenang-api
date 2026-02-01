-- Song categories table
CREATE TABLE song_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    thumbnail VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_song_categories_name ON song_categories(name);
CREATE INDEX idx_song_categories_deleted_at ON song_categories(deleted_at);
