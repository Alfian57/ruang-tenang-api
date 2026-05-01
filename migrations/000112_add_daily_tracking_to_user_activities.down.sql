DROP INDEX IF EXISTS idx_user_activities_date;
DROP INDEX IF EXISTS uq_user_activities_user_type_date;

ALTER TABLE user_activities
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS count,
DROP COLUMN IF EXISTS date;
