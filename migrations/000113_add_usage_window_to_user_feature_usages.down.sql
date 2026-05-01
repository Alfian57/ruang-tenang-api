DROP INDEX IF EXISTS idx_user_feature_usages_feature_window;

ALTER TABLE user_feature_usages
    DROP CONSTRAINT IF EXISTS uq_user_feature_usages_window;

DELETE FROM user_feature_usages older
USING user_feature_usages newer
WHERE older.user_id = newer.user_id
  AND older.feature_key = newer.feature_key
  AND older.usage_date = newer.usage_date
  AND older.id < newer.id;

ALTER TABLE user_feature_usages
    ADD CONSTRAINT uq_user_feature_usages UNIQUE (user_id, feature_key, usage_date);

ALTER TABLE user_feature_usages
    DROP COLUMN IF EXISTS usage_window_start;
