-- Story comments table
CREATE TABLE story_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES inspiring_stories(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(500) NOT NULL,
    heart_count INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT false,
    hidden_reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_story_comments_story ON story_comments(story_id);
CREATE INDEX idx_story_comments_user ON story_comments(user_id);
