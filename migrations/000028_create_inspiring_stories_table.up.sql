CREATE TABLE inspiring_stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    cover_image VARCHAR(500),
    is_anonymous BOOLEAN DEFAULT false,
    has_trigger_warning BOOLEAN DEFAULT false,
    trigger_warning_text VARCHAR(500),
    
    -- Moderation
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'revision_requested')),
    moderator_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    moderation_feedback TEXT,
    moderated_at TIMESTAMP WITH TIME ZONE,
    
    -- Stats
    view_count INTEGER DEFAULT 0,
    heart_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    
    -- Feature flags
    is_featured BOOLEAN DEFAULT false,
    featured_at TIMESTAMP WITH TIME ZONE,
    featured_until TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);

-- Story to category relationship (many-to-many)
CREATE TABLE story_category_relations (
    story_id UUID NOT NULL REFERENCES inspiring_stories(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES story_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (story_id, category_id)
);

-- Story tags for custom tags
CREATE TABLE story_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES inspiring_stories(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_inspiring_stories_author ON inspiring_stories(author_id);
CREATE INDEX idx_inspiring_stories_status ON inspiring_stories(status);
CREATE INDEX idx_inspiring_stories_featured ON inspiring_stories(is_featured);
CREATE INDEX idx_inspiring_stories_published ON inspiring_stories(published_at);
CREATE INDEX idx_inspiring_stories_hearts ON inspiring_stories(heart_count);
CREATE INDEX idx_story_tags_story ON story_tags(story_id);
CREATE INDEX idx_story_tags_tag ON story_tags(tag);
