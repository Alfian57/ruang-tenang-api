ALTER TABLE user_feature_usages
    ADD COLUMN usage_window_start TIMESTAMP;

UPDATE user_feature_usages
SET usage_window_start = usage_date::timestamp
WHERE usage_window_start IS NULL;

ALTER TABLE user_feature_usages
    ALTER COLUMN usage_window_start SET NOT NULL;

ALTER TABLE user_feature_usages
    DROP CONSTRAINT IF EXISTS uq_user_feature_usages;

ALTER TABLE user_feature_usages
    ADD CONSTRAINT uq_user_feature_usages_window UNIQUE (user_id, feature_key, usage_window_start);

CREATE INDEX idx_user_feature_usages_feature_window
    ON user_feature_usages(feature_key, usage_window_start);
