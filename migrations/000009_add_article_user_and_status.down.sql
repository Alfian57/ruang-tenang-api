-- Remove user_id and status columns from articles table
DROP INDEX IF EXISTS idx_articles_status;
DROP INDEX IF EXISTS idx_articles_user_id;
ALTER TABLE articles DROP COLUMN IF EXISTS status;
ALTER TABLE articles DROP COLUMN IF EXISTS user_id;
