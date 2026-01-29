-- Update level_configs with additional fields for rewards system
ALTER TABLE level_configs ADD COLUMN IF NOT EXISTS tier_name VARCHAR(50);
ALTER TABLE level_configs ADD COLUMN IF NOT EXISTS tier_color VARCHAR(20);
ALTER TABLE level_configs ADD COLUMN IF NOT EXISTS description TEXT;

-- Update existing level configs with tiers
UPDATE level_configs SET tier_name = 'Newcomer', tier_color = 'gray', description = 'Welcome to Ruang Tenang! Start your journey.' WHERE level IN (1, 2);
UPDATE level_configs SET tier_name = 'Explorer', tier_color = 'blue', description = 'You are exploring and growing.' WHERE level IN (3, 4, 5);
UPDATE level_configs SET tier_name = 'Trusted', tier_color = 'green', description = 'A trusted member of the community.' WHERE level IN (6, 7);
UPDATE level_configs SET tier_name = 'Veteran', tier_color = 'purple', description = 'A veteran with valuable experience.' WHERE level IN (8, 9);
UPDATE level_configs SET tier_name = 'Guardian', tier_color = 'gold', description = 'A guardian of the community.' WHERE level = 10;

-- Add user profile customization fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_theme VARCHAR(50) DEFAULT 'default';
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_banner VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_border_color VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS tagline VARCHAR(200);
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS current_streak INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS longest_streak INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_activity_date DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_activities INTEGER DEFAULT 0;

-- Create monthly hall of fame tracking
CREATE TABLE monthly_hall_of_fame (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    level INTEGER NOT NULL,
    month INTEGER NOT NULL, -- 1-12
    year INTEGER NOT NULL,
    rank INTEGER NOT NULL, -- 1, 2, or 3
    monthly_xp INTEGER NOT NULL,
    message VARCHAR(300), -- Optional inspiring message
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(level, month, year, rank)
);

CREATE INDEX idx_hall_of_fame_level_month ON monthly_hall_of_fame(level, month, year);
CREATE INDEX idx_hall_of_fame_user ON monthly_hall_of_fame(user_id);

-- Community stats tracking (monthly aggregates)
CREATE TABLE community_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    month INTEGER NOT NULL,
    year INTEGER NOT NULL,
    total_xp_earned BIGINT DEFAULT 0,
    active_members INTEGER DEFAULT 0,
    total_achievements INTEGER DEFAULT 0,
    new_members INTEGER DEFAULT 0,
    total_stories_published INTEGER DEFAULT 0,
    total_articles_published INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(month, year)
);

CREATE INDEX idx_community_stats_period ON community_stats(year, month);
