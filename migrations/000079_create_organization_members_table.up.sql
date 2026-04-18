CREATE TABLE organization_members (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(150),
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    status VARCHAR(20) NOT NULL DEFAULT 'invited',
    invitation_token VARCHAR(120),
    invitation_expires_at TIMESTAMP,
    invited_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    invited_at TIMESTAMP,
    joined_at TIMESTAMP,
    removed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_organization_members_role CHECK (role IN ('owner', 'admin', 'member')),
    CONSTRAINT chk_organization_members_status CHECK (status IN ('invited', 'pending_approval', 'active', 'removed')),
    CONSTRAINT uq_organization_members_org_email UNIQUE (organization_id, email),
    CONSTRAINT uq_organization_members_org_user UNIQUE (organization_id, user_id)
);

CREATE INDEX idx_organization_members_org_status ON organization_members(organization_id, status);
CREATE INDEX idx_organization_members_user_status ON organization_members(user_id, status);
CREATE UNIQUE INDEX uq_organization_members_invitation_token ON organization_members(invitation_token) WHERE invitation_token IS NOT NULL;
