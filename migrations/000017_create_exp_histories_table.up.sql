CREATE TABLE exp_histories (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    points INTEGER NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exp_histories_user_created ON exp_histories(user_id, created_at DESC);
CREATE INDEX idx_exp_histories_activity ON exp_histories(activity_type);
