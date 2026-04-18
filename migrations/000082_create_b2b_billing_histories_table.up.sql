CREATE TABLE b2b_billing_histories (
    id SERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES b2b_subscriptions(id) ON DELETE CASCADE,
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    billing_period_start TIMESTAMP NOT NULL,
    billing_period_end TIMESTAMP NOT NULL,
    seats_billed INTEGER NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_billing_histories_status CHECK (status IN ('pending', 'paid', 'failed', 'void')),
    CONSTRAINT chk_b2b_billing_histories_seats CHECK (seats_billed > 0),
    CONSTRAINT chk_b2b_billing_histories_amount CHECK (amount >= 0)
);

CREATE INDEX idx_b2b_billing_histories_subscription_status ON b2b_billing_histories(subscription_id, status);
