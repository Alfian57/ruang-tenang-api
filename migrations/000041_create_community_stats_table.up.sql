-- Community stats table
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
