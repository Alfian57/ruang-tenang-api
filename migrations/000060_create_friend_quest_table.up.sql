-- Friend quests (collaborative 1-on-1 missions)
CREATE TABLE IF NOT EXISTS friend_quests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    partner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    quest_type VARCHAR(50) NOT NULL DEFAULT 'total_xp',
    target_value INTEGER NOT NULL,
    requester_progress INTEGER NOT NULL DEFAULT 0,
    partner_progress INTEGER NOT NULL DEFAULT 0,
    xp_reward INTEGER NOT NULL DEFAULT 0,
    coin_reward INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    starts_at TIMESTAMP WITH TIME ZONE,
    ends_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_friend_quests_requester ON friend_quests(requester_id);
CREATE INDEX idx_friend_quests_partner ON friend_quests(partner_id);
CREATE INDEX idx_friend_quests_status ON friend_quests(status);
