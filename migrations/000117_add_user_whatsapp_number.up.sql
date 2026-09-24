ALTER TABLE users ADD COLUMN whatsapp_number VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN whatsapp_verified_at TIMESTAMPTZ NULL;
CREATE UNIQUE INDEX uq_users_whatsapp_number_verified ON users(whatsapp_number) WHERE whatsapp_verified_at IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE user_phone_verifications (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_hash VARCHAR(64) NOT NULL UNIQUE,
    code_hash VARCHAR(64) NOT NULL DEFAULT '',
    attempts INTEGER NOT NULL DEFAULT 0,
    remember_me BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_phone_verifications_user UNIQUE (user_id)
);
