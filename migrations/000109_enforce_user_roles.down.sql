ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

UPDATE users
SET role = 'user'
WHERE role IS NULL OR role = '';

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';
