-- Articles table with moderation fields consolidated
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    thumbnail VARCHAR(500),
    content TEXT NOT NULL,
    article_category_id INTEGER NOT NULL REFERENCES article_categories(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'published', 'draft', 'rejected'
    
    -- Moderation fields
    is_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    flag_reason TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_articles_title ON articles(title);
CREATE INDEX idx_articles_article_category_id ON articles(article_category_id);
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_is_flagged ON articles(is_flagged);
CREATE INDEX idx_articles_deleted_at ON articles(deleted_at);
