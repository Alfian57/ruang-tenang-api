-- User feature unlocks table
CREATE TABLE user_feature_unlocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_id UUID NOT NULL REFERENCES feature_definitions(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, feature_id)
);

CREATE INDEX idx_user_feature_unlocks_user ON user_feature_unlocks(user_id);
CREATE INDEX idx_user_feature_unlocks_feature ON user_feature_unlocks(feature_id);
