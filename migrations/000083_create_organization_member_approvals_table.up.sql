CREATE TABLE organization_member_approvals (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    organization_member_id BIGINT NOT NULL REFERENCES organization_members(id) ON DELETE CASCADE,
    requested_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    approver_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    note TEXT,
    decided_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_org_member_approvals_status CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_org_member_approvals_org_member ON organization_member_approvals(organization_id, organization_member_id);
CREATE INDEX idx_org_member_approvals_status ON organization_member_approvals(status);
