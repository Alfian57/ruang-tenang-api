-- Normalize legacy values and enforce strict single-role accounts.
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';

UPDATE users
SET role = LOWER(TRIM(role));

UPDATE users
SET role = 'admin'
WHERE role = 'moderator';

UPDATE users
SET role = 'user'
WHERE role = 'member' OR role IS NULL OR role = '';

UPDATE users
SET role = 'user'
WHERE role NOT IN ('admin', 'user', 'mitra');

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

ALTER TABLE users
ADD CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user', 'mitra'));
