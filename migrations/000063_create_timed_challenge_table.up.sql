-- Timed challenge templates
CREATE TABLE IF NOT EXISTS timed_challenge_templates (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    challenge_type VARCHAR(50) NOT NULL,
    target_value INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 60,
    xp_reward INTEGER NOT NULL DEFAULT 50,
    coin_reward INTEGER NOT NULL DEFAULT 0,
    icon VARCHAR(50) NOT NULL DEFAULT '⚡',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User timed challenges (assigned instances)
CREATE TABLE IF NOT EXISTS user_timed_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id INTEGER NOT NULL REFERENCES timed_challenge_templates(id) ON DELETE CASCADE,
    current_value INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_timed_challenges_user ON user_timed_challenges(user_id);
CREATE INDEX idx_user_timed_challenges_status ON user_timed_challenges(user_id, status);
CREATE INDEX idx_user_timed_challenges_expires ON user_timed_challenges(expires_at);
