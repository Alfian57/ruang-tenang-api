-- Forum categories table
CREATE TABLE forum_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(300) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_forum_categories_name ON forum_categories(name);
CREATE INDEX idx_forum_categories_deleted_at ON forum_categories(deleted_at);
CREATE UNIQUE INDEX idx_forum_categories_slug ON forum_categories(slug) WHERE deleted_at IS NULL;
