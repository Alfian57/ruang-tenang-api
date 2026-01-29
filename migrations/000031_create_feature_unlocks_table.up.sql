-- Feature unlock definitions
CREATE TABLE feature_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_key VARCHAR(100) NOT NULL UNIQUE,
    feature_name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    required_level INTEGER NOT NULL DEFAULT 1,
    category VARCHAR(50) DEFAULT 'general',
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Track which features users have unlocked (automatically tracks when user levels up)
CREATE TABLE user_feature_unlocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_id UUID NOT NULL REFERENCES feature_definitions(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, feature_id)
);

CREATE INDEX idx_feature_definitions_level ON feature_definitions(required_level);
CREATE INDEX idx_user_feature_unlocks_user ON user_feature_unlocks(user_id);

-- Insert default feature definitions based on requirements
INSERT INTO feature_definitions (feature_key, feature_name, description, icon, required_level, category, display_order) VALUES
-- Level 1-2 (Free for all)
('basic_access', 'Basic Access', 'Access to all core features', '✅', 1, 'core', 1),
('basic_breathing', '4-7-8 Breathing', 'Basic breathing technique', '🌬️', 1, 'breathing', 2),

-- Level 3
('custom_playlists', 'Custom Playlists', 'Create your own music playlists', '🎵', 3, 'music', 10),
('advanced_mood_tracking', 'Advanced Mood Charts', 'View detailed mood analytics and charts', '📊', 3, 'mood', 11),
('breathing_timer_custom', 'Custom Breathing Timer', 'Customize breathing exercise duration', '⏱️', 3, 'breathing', 12),

-- Level 4
('all_breathing_techniques', 'All Breathing Techniques', 'Access to 5+ breathing techniques', '🧘', 4, 'breathing', 20),
('chat_folders', 'AI Chat Folders', 'Organize your conversations in folders', '💬', 4, 'chat', 21),
('profile_themes', 'Profile Themes', 'Customize your profile appearance', '🎨', 4, 'profile', 22),

-- Level 5
('private_journal', 'Private Journal', 'Write personal journal entries', '📝', 5, 'journal', 30),
('pin_messages', 'Pin Messages', 'Pin important chat messages', '📌', 5, 'chat', 31),
('publish_articles', 'Publish Articles', 'Write and publish user-generated articles', '📖', 5, 'content', 32),
('submit_stories', 'Submit Inspiring Stories', 'Share your recovery journey', '✨', 5, 'content', 33),

-- Level 6
('chat_export', 'Export Conversations', 'Export chat history to PDF/TXT', '📥', 6, 'chat', 40),
('voice_input', 'Voice Input', 'Use voice input for chat messages', '🎤', 6, 'chat', 41),
('journal_tags', 'Custom Journal Tags', 'Create custom tags for journal entries', '🏷️', 6, 'journal', 42),
('create_polls', 'Create Forum Polls', 'Create polls in forum discussions', '🗳️', 6, 'forum', 43),

-- Level 7
('chat_summaries', 'AI Chat Summaries', 'Generate AI summaries of conversations', '📋', 7, 'chat', 50),
('report_content', 'Report Content', 'Report inappropriate content', '🚨', 7, 'moderation', 51),
('vote_best_answer', 'Vote Best Answer', 'Vote for best answers in forum', '🎯', 7, 'forum', 52),
('custom_notifications', 'Custom Notifications', 'Advanced notification preferences', '🔔', 7, 'profile', 53),

-- Level 8
('premium_themes', 'Premium Themes', 'Access to premium themes and customizations', '🌈', 8, 'profile', 60),
('exclusive_content', 'Exclusive Content', 'Access exclusive mental health resources', '📚', 8, 'content', 61),
('nominate_stories', 'Nominate Stories', 'Nominate inspiring stories for featured section', '🏅', 8, 'content', 62),
('advanced_search', 'Advanced Search', 'Advanced search filters', '⚙️', 8, 'general', 63),

-- Level 9
('moderator_application', 'Moderator Application', 'Apply to become a community moderator', '👤', 9, 'moderation', 70),
('mentor_badge', 'Community Mentor Badge', 'Special Community Mentor badge', '✨', 9, 'badge', 71),
('beta_features', 'Beta Features', 'Early access to new features', '🎁', 9, 'general', 72),
('featured_shoutout', 'Featured Shoutout', 'Opportunity for featured member shoutout', '💌', 9, 'profile', 73),

-- Level 10
('all_features', 'All Features', 'Lifetime access to all features', '👑', 10, 'general', 80),
('guardian_badge', 'Guardian Badge', 'Exclusive Guardian badge', '🎖️', 10, 'badge', 81),
('community_voting', 'Community Voting', 'Vote on community feature decisions', '📢', 10, 'moderation', 82),
('guardian_profile', 'Guardian Profile', 'Profile featured in Community Guardians', '🌟', 10, 'profile', 83);
