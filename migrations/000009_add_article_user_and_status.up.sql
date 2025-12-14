-- Add user_id and status columns to articles table
ALTER TABLE articles ADD COLUMN user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE articles ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'published';

-- Create index for user_id
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_status ON articles(status);
