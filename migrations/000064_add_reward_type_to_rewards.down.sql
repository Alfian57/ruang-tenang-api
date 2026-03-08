DROP INDEX IF EXISTS idx_rewards_reward_type;
ALTER TABLE rewards DROP COLUMN IF EXISTS reward_value;
ALTER TABLE rewards DROP COLUMN IF EXISTS reward_type;
