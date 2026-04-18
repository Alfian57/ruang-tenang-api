CREATE TABLE IF NOT EXISTS user_landmark_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    landmark_id UUID NOT NULL REFERENCES map_landmarks(id) ON DELETE CASCADE,
    is_unlocked BOOLEAN DEFAULT FALSE,
    current_value INTEGER NOT NULL DEFAULT 0,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    reward_claimed BOOLEAN DEFAULT FALSE,
    UNIQUE(user_id, landmark_id)
);

CREATE INDEX idx_user_landmark_progress_user_id ON user_landmark_progress(user_id);
