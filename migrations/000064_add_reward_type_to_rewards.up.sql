-- Add reward_type and reward_value to rewards table for theme/xp_boost support
ALTER TABLE rewards ADD COLUMN reward_type VARCHAR(50) NOT NULL DEFAULT 'general';
ALTER TABLE rewards ADD COLUMN reward_value VARCHAR(100) DEFAULT '';

CREATE INDEX idx_rewards_reward_type ON rewards(reward_type);
