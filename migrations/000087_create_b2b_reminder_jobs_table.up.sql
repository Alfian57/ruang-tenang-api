CREATE TABLE b2b_reminder_jobs (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subscription_id BIGINT REFERENCES b2b_subscriptions(id) ON DELETE SET NULL,
    job_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    due_at TIMESTAMP NOT NULL,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_reminder_jobs_type CHECK (job_type IN ('seat_threshold', 'subscription_expiry', 'invoice_due')),
    CONSTRAINT chk_b2b_reminder_jobs_status CHECK (status IN ('pending', 'sent', 'failed')),
    CONSTRAINT chk_b2b_reminder_jobs_attempt_count CHECK (attempt_count >= 0)
);

CREATE UNIQUE INDEX uq_b2b_reminder_jobs_org_type_due ON b2b_reminder_jobs(organization_id, job_type, due_at);
CREATE INDEX idx_b2b_reminder_jobs_due_status ON b2b_reminder_jobs(organization_id, status, due_at);
