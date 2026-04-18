-- Song categories table
CREATE TABLE song_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(300) NOT NULL,
    thumbnail VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_song_categories_name ON song_categories(name);
CREATE INDEX idx_song_categories_deleted_at ON song_categories(deleted_at);
CREATE UNIQUE INDEX idx_song_categories_slug ON song_categories(slug) WHERE deleted_at IS NULL;
