DROP INDEX IF EXISTS uq_b2b_subscriptions_org_active;

CREATE INDEX IF NOT EXISTS idx_b2b_subscriptions_org_active_period
ON b2b_subscriptions(organization_id, status, starts_at, ends_at);
