ALTER TABLE payment_transactions
    ADD COLUMN refunded_amount BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN refund_requested_amount BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN provider_refund_amount_reported BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN refund_status VARCHAR(32) NOT NULL DEFAULT 'none',
    ADD COLUMN refund_reconciliation_status VARCHAR(16) NOT NULL DEFAULT 'not_required',
    ADD COLUMN refund_reconciliation_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN coins_reversed BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN coins_written_off BIGINT NOT NULL DEFAULT 0;

ALTER TABLE payment_transactions
    ADD CONSTRAINT chk_payment_transactions_refund_status
        CHECK (refund_status IN ('none', 'pending_confirmation', 'partially_refunded', 'refunded')),
    ADD CONSTRAINT chk_payment_transactions_refund_reconciliation_status
        CHECK (refund_reconciliation_status IN ('not_required', 'pending', 'resolved')),
    ADD CONSTRAINT chk_payment_transactions_refund_amount
        CHECK (refunded_amount >= 0 AND refunded_amount <= amount),
    ADD CONSTRAINT chk_payment_transactions_requested_refund_amount
        CHECK (refund_requested_amount >= 0 AND refund_requested_amount <= amount),
    ADD CONSTRAINT chk_payment_transactions_reported_refund_amount
        CHECK (provider_refund_amount_reported >= 0 AND provider_refund_amount_reported <= amount),
    ADD CONSTRAINT chk_payment_transactions_refund_coins
        CHECK (coins_reversed >= 0 AND coins_written_off >= 0);

UPDATE payment_transactions
SET refunded_amount = amount,
    refund_requested_amount = amount,
    provider_refund_amount_reported = amount,
    refund_status = 'refunded'
WHERE status = 'refunded';

UPDATE payment_transactions
SET refund_reconciliation_status = 'pending',
    refund_reconciliation_reason = 'legacy_topup_refund_requires_coin_review'
WHERE status = 'refunded' AND item_type = 'topup';

UPDATE payment_transactions
SET refund_reconciliation_status = 'pending',
    refund_reconciliation_reason = 'legacy_subscription_refund_requires_entitlement_review'
WHERE status = 'refunded' AND item_type = 'subscription';

CREATE TABLE payment_refunds (
    id BIGSERIAL PRIMARY KEY,
    payment_transaction_id INTEGER NOT NULL REFERENCES payment_transactions(id) ON DELETE CASCADE,
    refund_key VARCHAR(120) NOT NULL,
    provider_refund_id VARCHAR(120) NOT NULL DEFAULT '',
    amount BIGINT NOT NULL CHECK (amount > 0),
    reason VARCHAR(255) NOT NULL DEFAULT '',
    refund_method VARCHAR(32) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'requested',
    requested_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    provider_created_at TIMESTAMPTZ NULL,
    bank_confirmed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payment_refunds_refund_key UNIQUE (payment_transaction_id, refund_key),
    CONSTRAINT uq_payment_refunds_provider_id UNIQUE (payment_transaction_id, provider_refund_id),
    CONSTRAINT chk_payment_refunds_status CHECK (status IN ('requested', 'pending_confirmation', 'confirmed', 'rejected'))
);

CREATE INDEX idx_payment_refunds_transaction_id ON payment_refunds(payment_transaction_id);
CREATE INDEX idx_payment_refunds_status ON payment_refunds(status);

CREATE TABLE payment_refund_reconciliation_events (
    id BIGSERIAL PRIMARY KEY,
    payment_transaction_id INTEGER NOT NULL REFERENCES payment_transactions(id) ON DELETE CASCADE,
    actor_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(40) NOT NULL,
    note TEXT NOT NULL,
    refunded_amount BIGINT NOT NULL DEFAULT 0,
    coins_reversed BIGINT NOT NULL DEFAULT 0,
    coins_written_off BIGINT NOT NULL DEFAULT 0,
    premium_days_reduced INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_refund_reconciliation_transaction_id
    ON payment_refund_reconciliation_events(payment_transaction_id);
