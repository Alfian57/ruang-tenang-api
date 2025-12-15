ALTER TABLE forums ADD COLUMN category_id INTEGER;
ALTER TABLE forums ADD CONSTRAINT fk_forums_category FOREIGN KEY (category_id) REFERENCES forum_categories(id);
CREATE INDEX idx_forums_category_id ON forums(category_id);
