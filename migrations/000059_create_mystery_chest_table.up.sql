-- Mystery chests
CREATE TABLE IF NOT EXISTS user_chests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rarity VARCHAR(20) NOT NULL DEFAULT 'common',
    is_opened BOOLEAN DEFAULT FALSE,
    reward_type VARCHAR(50),
    reward_value INTEGER DEFAULT 0,
    reward_label VARCHAR(200),
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'milestone',
    trigger_description VARCHAR(200),
    opened_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_chests_user ON user_chests(user_id);
CREATE INDEX idx_user_chests_unopened ON user_chests(user_id, is_opened);
