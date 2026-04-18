CREATE TABLE b2b_sso_configs (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    provider VARCHAR(30),
    issuer_url VARCHAR(500),
    entrypoint_url VARCHAR(500),
    audience VARCHAR(255),
    certificate_pem TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    enforce_sso BOOLEAN NOT NULL DEFAULT FALSE,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_sso_configs_provider CHECK (
        provider IS NULL OR provider IN ('saml', 'oidc', 'google_workspace', 'azure_ad', 'okta')
    )
);

CREATE INDEX idx_b2b_sso_configs_enabled ON b2b_sso_configs(is_enabled);
