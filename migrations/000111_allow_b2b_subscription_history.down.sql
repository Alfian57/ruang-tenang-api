DROP INDEX IF EXISTS idx_b2b_subscriptions_org_active_period;

CREATE UNIQUE INDEX uq_b2b_subscriptions_org_active
ON b2b_subscriptions(organization_id)
WHERE status = 'active';
