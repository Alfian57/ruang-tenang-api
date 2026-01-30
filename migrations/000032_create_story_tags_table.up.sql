-- Story tags table
CREATE TABLE story_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES inspiring_stories(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_story_tags_story ON story_tags(story_id);
CREATE INDEX idx_story_tags_tag ON story_tags(tag);
