CREATE TABLE organization_onboarding_templates (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    title VARCHAR(150) NOT NULL,
    welcome_message TEXT,
    checklist_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_org_onboarding_templates_role CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE UNIQUE INDEX uq_org_onboarding_templates_org_role ON organization_onboarding_templates(organization_id, role);
CREATE INDEX idx_org_onboarding_templates_active ON organization_onboarding_templates(organization_id, is_active);
