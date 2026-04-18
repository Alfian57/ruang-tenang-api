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
