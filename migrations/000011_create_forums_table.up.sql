-- Forums table with all fields consolidated
CREATE TABLE forums (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES forum_categories(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(300) NOT NULL,
    content TEXT NOT NULL,
    
    -- Moderation fields
    trigger_warnings JSONB,
    is_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    flagged_reason TEXT,
    
    -- Best answer tracking
    has_accepted_answer BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_forums_title ON forums(title);
CREATE INDEX idx_forums_user_id ON forums(user_id);
CREATE INDEX idx_forums_category_id ON forums(category_id);
CREATE INDEX idx_forums_is_flagged ON forums(is_flagged);
CREATE INDEX idx_forums_has_accepted ON forums(has_accepted_answer);
CREATE INDEX idx_forums_deleted_at ON forums(deleted_at);
CREATE UNIQUE INDEX idx_forums_slug ON forums(slug) WHERE deleted_at IS NULL;
