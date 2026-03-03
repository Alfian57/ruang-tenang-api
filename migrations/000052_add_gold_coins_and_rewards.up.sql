-- Add gold_coins to users
ALTER TABLE users ADD COLUMN gold_coins INTEGER NOT NULL DEFAULT 0;

-- Add coin_reward to daily_tasks
ALTER TABLE daily_tasks ADD COLUMN coin_reward INTEGER NOT NULL DEFAULT 0;

-- Rewards table (managed by admin)
CREATE TABLE rewards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image VARCHAR(500) DEFAULT '',
    coin_cost INTEGER NOT NULL,
    stock INTEGER NOT NULL DEFAULT -1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Reward claims table
CREATE TABLE reward_claims (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_id INTEGER NOT NULL REFERENCES rewards(id) ON DELETE CASCADE,
    coin_spent INTEGER NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_reward_claims_user_id ON reward_claims(user_id);
CREATE INDEX idx_reward_claims_reward_id ON reward_claims(reward_id);
CREATE INDEX idx_rewards_is_active ON rewards(is_active);
