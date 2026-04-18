CREATE TABLE user_feature_usages (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_key VARCHAR(50) NOT NULL,
    usage_date DATE NOT NULL,
    used_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_feature_usages UNIQUE (user_id, feature_key, usage_date)
);

CREATE INDEX idx_user_feature_usages_feature_date ON user_feature_usages(feature_key, usage_date);
