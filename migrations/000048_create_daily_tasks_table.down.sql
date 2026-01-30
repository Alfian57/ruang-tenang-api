DROP INDEX IF EXISTS idx_daily_tasks_claimed;
DROP INDEX IF EXISTS idx_daily_tasks_completed;
DROP INDEX IF EXISTS idx_daily_tasks_user_date;
DROP INDEX IF EXISTS idx_daily_tasks_task_date;
DROP INDEX IF EXISTS idx_daily_tasks_user_id;

DROP TABLE IF EXISTS daily_tasks;

ALTER TABLE users 
DROP COLUMN IF EXISTS last_login_date,
DROP COLUMN IF EXISTS login_streak;
