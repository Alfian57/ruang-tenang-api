CREATE TABLE story_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50) DEFAULT '📖',
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_story_categories_slug ON story_categories(slug);
CREATE INDEX idx_story_categories_active ON story_categories(is_active);

-- Insert default categories
INSERT INTO story_categories (name, slug, description, icon, display_order) VALUES
('Recovery Journey', 'recovery-journey', 'Share your path to recovery and healing', '🌱', 1),
('Overcoming Depression', 'overcoming-depression', 'Stories about battling and overcoming depression', '☀️', 2),
('Anxiety Management', 'anxiety-management', 'Experiences with managing anxiety', '🧘', 3),
('Healing from Trauma', 'healing-from-trauma', 'Journey of healing from traumatic experiences', '💚', 4),
('Finding Hope', 'finding-hope', 'Stories about finding hope in difficult times', '✨', 5),
('Self-Care Journey', 'self-care-journey', 'Personal self-care practices and discoveries', '🌸', 6),
('Professional Help Experience', 'professional-help', 'Experiences with therapy, counseling, or treatment', '🏥', 7),
('Other', 'other', 'Other mental health related stories', '📝', 8);
