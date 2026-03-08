-- Daily spin rewards definition
CREATE TABLE IF NOT EXISTS spin_rewards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50) NOT NULL,
    reward_type VARCHAR(50) NOT NULL,
    reward_value INTEGER NOT NULL DEFAULT 0,
    weight INTEGER NOT NULL DEFAULT 100,
    rarity VARCHAR(20) NOT NULL DEFAULT 'common',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User spin history
CREATE TABLE IF NOT EXISTS user_spins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_id INTEGER NOT NULL REFERENCES spin_rewards(id) ON DELETE CASCADE,
    spin_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, spin_date)
);

CREATE INDEX idx_user_spins_user ON user_spins(user_id);
CREATE INDEX idx_user_spins_date ON user_spins(spin_date);
