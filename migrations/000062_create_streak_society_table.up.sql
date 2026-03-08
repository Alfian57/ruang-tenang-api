-- Streak society tiers
CREATE TABLE IF NOT EXISTS streak_societies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    icon VARCHAR(50) NOT NULL,
    min_streak INTEGER NOT NULL UNIQUE,
    border_color VARCHAR(20) NOT NULL DEFAULT '#888888',
    badge_glow BOOLEAN DEFAULT FALSE,
    exclusive_chat BOOLEAN DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User society memberships
CREATE TABLE IF NOT EXISTS user_society_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    society_id INTEGER NOT NULL REFERENCES streak_societies(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(user_id, society_id)
);

CREATE INDEX idx_user_society_user ON user_society_memberships(user_id);
CREATE INDEX idx_user_society_active ON user_society_memberships(user_id, is_active);
