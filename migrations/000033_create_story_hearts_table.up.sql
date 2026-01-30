-- Story hearts table
CREATE TABLE story_hearts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES inspiring_stories(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(story_id, user_id)
);

CREATE INDEX idx_story_hearts_story ON story_hearts(story_id);
CREATE INDEX idx_story_hearts_user ON story_hearts(user_id);
