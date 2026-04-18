CREATE TABLE b2b_pricing_quotes (
    id SERIAL PRIMARY KEY,
    quote_code VARCHAR(60) NOT NULL UNIQUE,
    organization_id BIGINT REFERENCES organizations(id) ON DELETE SET NULL,
    plan_id BIGINT REFERENCES b2b_plans(id) ON DELETE SET NULL,
    requested_seats INTEGER NOT NULL,
    billing_cycle VARCHAR(20) NOT NULL,
    selected_addons_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    base_price_per_seat BIGINT NOT NULL,
    gross_amount BIGINT NOT NULL,
    volume_discount_amount BIGINT NOT NULL DEFAULT 0,
    annual_discount_amount BIGINT NOT NULL DEFAULT 0,
    add_on_amount BIGINT NOT NULL DEFAULT 0,
    final_amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    valid_until TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_pricing_quotes_requested_seats CHECK (requested_seats > 0),
    CONSTRAINT chk_b2b_pricing_quotes_billing_cycle CHECK (billing_cycle IN ('monthly', 'yearly')),
    CONSTRAINT chk_b2b_pricing_quotes_status CHECK (status IN ('draft', 'accepted', 'expired')),
    CONSTRAINT chk_b2b_pricing_quotes_amounts CHECK (
        base_price_per_seat >= 0 AND gross_amount >= 0 AND volume_discount_amount >= 0 AND
        annual_discount_amount >= 0 AND add_on_amount >= 0 AND final_amount >= 0
    )
);

CREATE INDEX idx_b2b_pricing_quotes_org_status ON b2b_pricing_quotes(organization_id, status);
CREATE INDEX idx_b2b_pricing_quotes_valid_until ON b2b_pricing_quotes(valid_until);
