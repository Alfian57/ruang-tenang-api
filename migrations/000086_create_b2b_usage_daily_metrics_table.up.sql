CREATE TABLE b2b_usage_daily_metrics (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    active_members INTEGER NOT NULL DEFAULT 0,
    invited_members INTEGER NOT NULL DEFAULT 0,
    pending_approvals INTEGER NOT NULL DEFAULT 0,
    contracted_seats INTEGER NOT NULL DEFAULT 0,
    used_seats INTEGER NOT NULL DEFAULT 0,
    messages_sent INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_usage_daily_metrics_non_negative CHECK (
        active_members >= 0 AND invited_members >= 0 AND pending_approvals >= 0 AND
        contracted_seats >= 0 AND used_seats >= 0 AND messages_sent >= 0
    )
);

CREATE UNIQUE INDEX uq_b2b_usage_daily_metrics_org_date ON b2b_usage_daily_metrics(organization_id, metric_date);
CREATE INDEX idx_b2b_usage_daily_metrics_org_date ON b2b_usage_daily_metrics(organization_id, metric_date DESC);
