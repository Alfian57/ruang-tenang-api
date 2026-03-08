-- XP Boosts (double XP timer)
CREATE TABLE IF NOT EXISTS xp_boosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    multiplier DECIMAL(3,1) NOT NULL DEFAULT 2.0,
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'activity_chain',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_xp_boosts_user_active ON xp_boosts(user_id, is_active);
CREATE INDEX idx_xp_boosts_expires ON xp_boosts(expires_at);

-- User combo state (chain multiplier)
CREATE TABLE IF NOT EXISTS user_combos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    combo_count INTEGER NOT NULL DEFAULT 0,
    multiplier DECIMAL(3,1) NOT NULL DEFAULT 1.0,
    last_activity_type VARCHAR(50),
    last_activity_at TIMESTAMP WITH TIME ZONE,
    session_started_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_combos_user ON user_combos(user_id);
