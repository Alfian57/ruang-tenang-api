ALTER TABLE user_activities
ADD COLUMN IF NOT EXISTS date DATE NOT NULL DEFAULT CURRENT_DATE,
ADD COLUMN IF NOT EXISTS count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE user_activities
SET date = created_at::date,
    count = GREATEST(count, 1),
    updated_at = created_at
WHERE created_at IS NOT NULL;

WITH ranked_activities AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id, activity_type, date
            ORDER BY created_at ASC, id ASC
        ) AS activity_rank,
        SUM(GREATEST(count, 1)) OVER (
            PARTITION BY user_id, activity_type, date
        ) AS total_count,
        MAX(updated_at) OVER (
            PARTITION BY user_id, activity_type, date
        ) AS last_updated_at
    FROM user_activities
)
UPDATE user_activities
SET count = ranked_activities.total_count,
    updated_at = ranked_activities.last_updated_at
FROM ranked_activities
WHERE user_activities.id = ranked_activities.id
  AND ranked_activities.activity_rank = 1;

WITH ranked_activities AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id, activity_type, date
            ORDER BY created_at ASC, id ASC
        ) AS activity_rank
    FROM user_activities
)
DELETE FROM user_activities
USING ranked_activities
WHERE user_activities.id = ranked_activities.id
  AND ranked_activities.activity_rank > 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_activities_user_type_date
ON user_activities(user_id, activity_type, date);

CREATE INDEX IF NOT EXISTS idx_user_activities_date ON user_activities(date);
