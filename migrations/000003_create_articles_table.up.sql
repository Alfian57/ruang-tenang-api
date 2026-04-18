-- Articles table with moderation fields consolidated
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(300) NOT NULL,
    thumbnail VARCHAR(500),
    content TEXT NOT NULL,
    article_category_id INTEGER NOT NULL REFERENCES article_categories(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'published', -- 'draft', 'published', 'blocked'
    
    -- Moderation fields
    moderation_status VARCHAR(50) DEFAULT 'pending',
    moderation_notes TEXT,
    moderated_by_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    moderated_at TIMESTAMP,
    trigger_warnings JSONB,
    is_user_generated BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_articles_title ON articles(title);
CREATE INDEX idx_articles_article_category_id ON articles(article_category_id);
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_moderation_status ON articles(moderation_status);
CREATE INDEX idx_articles_is_user_generated ON articles(is_user_generated);
CREATE INDEX idx_articles_deleted_at ON articles(deleted_at);
CREATE UNIQUE INDEX idx_articles_slug ON articles(slug) WHERE deleted_at IS NULL;
