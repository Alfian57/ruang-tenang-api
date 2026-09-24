DROP TABLE IF EXISTS user_phone_verifications;
DROP INDEX IF EXISTS uq_users_whatsapp_number_verified;
ALTER TABLE users DROP COLUMN IF EXISTS whatsapp_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS whatsapp_number;
