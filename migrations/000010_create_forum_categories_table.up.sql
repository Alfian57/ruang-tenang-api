-- Forum categories table
CREATE TABLE forum_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_forum_categories_name ON forum_categories(name);
CREATE INDEX idx_forum_categories_deleted_at ON forum_categories(deleted_at);
