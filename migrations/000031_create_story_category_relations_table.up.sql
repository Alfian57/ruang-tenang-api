-- Story category relations table (many-to-many)
CREATE TABLE story_category_relations (
    story_id UUID NOT NULL REFERENCES inspiring_stories(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES story_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (story_id, category_id)
);

CREATE INDEX idx_story_category_relations_story ON story_category_relations(story_id);
CREATE INDEX idx_story_category_relations_category ON story_category_relations(category_id);
