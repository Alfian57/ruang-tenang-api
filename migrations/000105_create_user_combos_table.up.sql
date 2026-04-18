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
