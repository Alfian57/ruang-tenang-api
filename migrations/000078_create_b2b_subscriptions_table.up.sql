CREATE TABLE b2b_subscriptions (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_id BIGINT NOT NULL REFERENCES b2b_plans(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    contracted_seats INTEGER NOT NULL,
    used_seats INTEGER NOT NULL DEFAULT 0,
    billing_cycle VARCHAR(20) NOT NULL,
    unit_price BIGINT NOT NULL,
    subtotal BIGINT NOT NULL,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,
    activated_at TIMESTAMP,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_subscriptions_status CHECK (status IN ('draft', 'active', 'suspended', 'expired', 'canceled')),
    CONSTRAINT chk_b2b_subscriptions_billing_cycle CHECK (billing_cycle IN ('monthly', 'yearly')),
    CONSTRAINT chk_b2b_subscriptions_contracted_seats CHECK (contracted_seats > 0),
    CONSTRAINT chk_b2b_subscriptions_used_seats CHECK (used_seats >= 0 AND used_seats <= contracted_seats),
    CONSTRAINT chk_b2b_subscriptions_amounts CHECK (unit_price >= 0 AND subtotal >= 0 AND discount_amount >= 0 AND total_amount >= 0)
);

CREATE INDEX idx_b2b_subscriptions_org_status ON b2b_subscriptions(organization_id, status);
CREATE INDEX idx_b2b_subscriptions_period ON b2b_subscriptions(starts_at, ends_at);
CREATE UNIQUE INDEX uq_b2b_subscriptions_org_active ON b2b_subscriptions(organization_id) WHERE status = 'active';
