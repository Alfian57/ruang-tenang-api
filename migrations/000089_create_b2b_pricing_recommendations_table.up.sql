CREATE TABLE b2b_pricing_recommendations (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    generated_for_date DATE NOT NULL,
    recommended_plan_id BIGINT REFERENCES b2b_plans(id) ON DELETE SET NULL,
    recommended_billing_cycle VARCHAR(20) NOT NULL,
    recommended_seats INTEGER NOT NULL,
    estimated_monthly_cost BIGINT NOT NULL DEFAULT 0,
    estimated_yearly_saving BIGINT NOT NULL DEFAULT 0,
    confidence_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    reasons_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_pricing_recommendations_cycle CHECK (recommended_billing_cycle IN ('monthly', 'yearly')),
    CONSTRAINT chk_b2b_pricing_recommendations_seats CHECK (recommended_seats > 0),
    CONSTRAINT chk_b2b_pricing_recommendations_amounts CHECK (
        estimated_monthly_cost >= 0 AND estimated_yearly_saving >= 0
    ),
    CONSTRAINT chk_b2b_pricing_recommendations_confidence CHECK (confidence_score >= 0 AND confidence_score <= 100)
);

CREATE UNIQUE INDEX uq_b2b_pricing_recommendations_org_date ON b2b_pricing_recommendations(organization_id, generated_for_date);
CREATE INDEX idx_b2b_pricing_recommendations_org_created_at ON b2b_pricing_recommendations(organization_id, created_at DESC);
