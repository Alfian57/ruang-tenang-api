-- Article categories table
CREATE TABLE article_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(300) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_article_categories_name ON article_categories(name);
CREATE INDEX idx_article_categories_deleted_at ON article_categories(deleted_at);
CREATE UNIQUE INDEX idx_article_categories_slug ON article_categories(slug) WHERE deleted_at IS NULL;
