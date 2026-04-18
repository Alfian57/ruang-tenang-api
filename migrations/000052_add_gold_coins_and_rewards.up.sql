CREATE TABLE rewards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image VARCHAR(500) DEFAULT '',
    coin_cost INTEGER NOT NULL,
    reward_type VARCHAR(50) NOT NULL DEFAULT 'general',
    reward_value VARCHAR(100) DEFAULT '',
    stock INTEGER NOT NULL DEFAULT -1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rewards_is_active ON rewards(is_active);
CREATE INDEX idx_rewards_reward_type ON rewards(reward_type);
