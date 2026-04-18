CREATE TABLE IF NOT EXISTS guild_challenge_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES guild_challenges(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value INTEGER NOT NULL DEFAULT 0,
    contributed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_guild_challenge_contributions_challenge_id ON guild_challenge_contributions(challenge_id);
CREATE INDEX idx_guild_challenge_contributions_user_id ON guild_challenge_contributions(user_id);
