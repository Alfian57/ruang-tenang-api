-- Story comment hearts table
CREATE TABLE story_comment_hearts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES story_comments(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(comment_id, user_id)
);

CREATE INDEX idx_story_comment_hearts_comment ON story_comment_hearts(comment_id);
CREATE INDEX idx_story_comment_hearts_user ON story_comment_hearts(user_id);
