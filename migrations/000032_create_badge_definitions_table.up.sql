-- Badge definitions
CREATE TABLE badge_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    badge_key VARCHAR(100) NOT NULL UNIQUE,
    badge_name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    category VARCHAR(50) DEFAULT 'general',
    requirement_type VARCHAR(50) NOT NULL, -- 'streak', 'activity_count', 'level', 'manual', 'story'
    requirement_value INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User badges (earned badges)
CREATE TABLE user_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id UUID NOT NULL REFERENCES badge_definitions(id) ON DELETE CASCADE,
    earned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_showcased BOOLEAN DEFAULT false, -- User can showcase up to 5 badges
    UNIQUE(user_id, badge_id)
);

CREATE INDEX idx_badge_definitions_category ON badge_definitions(category);
CREATE INDEX idx_user_badges_user ON user_badges(user_id);
CREATE INDEX idx_user_badges_showcased ON user_badges(user_id, is_showcased);

-- Insert default badge definitions
INSERT INTO badge_definitions (badge_key, badge_name, description, icon, category, requirement_type, requirement_value, display_order) VALUES
-- Streak badges
('streak_7', '7-Day Streak', 'Active for 7 consecutive days', '🔥', 'engagement', 'streak', 7, 1),
('streak_14', '14-Day Streak', 'Active for 14 consecutive days', '🔥', 'engagement', 'streak', 14, 2),
('streak_30', '30-Day Streak', 'Active for 30 consecutive days', '🔥', 'engagement', 'streak', 30, 3),
('streak_60', '60-Day Streak', 'Active for 60 consecutive days', '🔥', 'engagement', 'streak', 60, 4),
('streak_100', '100-Day Streak', 'Active for 100 consecutive days', '🔥', 'engagement', 'streak', 100, 5),

-- Activity count badges
('activities_10', 'Getting Started', 'Completed 10 activities', '🌟', 'engagement', 'activity_count', 10, 10),
('activities_50', 'Regular User', 'Completed 50 activities', '⭐', 'engagement', 'activity_count', 50, 11),
('activities_100', 'Active Member', 'Completed 100 activities', '💫', 'engagement', 'activity_count', 100, 12),
('activities_500', 'Power User', 'Completed 500 activities', '✨', 'engagement', 'activity_count', 500, 13),

-- Contribution badges
('first_article', 'First Article', 'Published your first article', '📝', 'contribution', 'manual', 1, 20),
('articles_5', 'Article Writer', 'Published 5 articles', '✍️', 'contribution', 'activity_count', 5, 21),
('helpful_commenter', 'Helpful Commenter', 'Received 10 hearts on comments', '💬', 'contribution', 'manual', 10, 22),
('top_contributor', 'Top Contributor', 'Made significant contributions to community', '🏆', 'contribution', 'manual', 0, 23),

-- Level badges
('level_5', 'Level 5', 'Reached Level 5', '🌱', 'milestone', 'level', 5, 30),
('level_10', 'Level 10', 'Reached Level 10', '🌳', 'milestone', 'level', 10, 31),

-- XP badges
('xp_1000', '1K XP', 'Earned 1,000 XP', '💎', 'milestone', 'xp', 1000, 40),
('xp_5000', '5K XP', 'Earned 5,000 XP', '💎', 'milestone', 'xp', 5000, 41),
('xp_10000', '10K XP', 'Earned 10,000 XP', '💎', 'milestone', 'xp', 10000, 42),

-- Special badges
('beta_tester', 'Beta Tester', 'Helped test beta features', '🧪', 'special', 'manual', 0, 50),
('community_mentor', 'Community Mentor', 'Recognized as a community mentor', '🎓', 'special', 'level', 9, 51),
('guardian', 'Guardian', 'Reached Guardian status', '👑', 'special', 'level', 10, 52),

-- Story badges
('first_story', 'Storyteller', 'Published your first inspiring story', '📖', 'story', 'story', 1, 60),
('stories_3', 'Inspiring Storyteller', 'Published 3 inspiring stories', '🌟', 'story', 'story', 3, 61),
('story_100_hearts', 'Beloved Story', 'A story received 100 hearts', '💚', 'story', 'manual', 100, 62);
