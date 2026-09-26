ALTER TABLE payment_transactions
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_status;

UPDATE payment_transactions
SET status = 'canceled'
WHERE status = 'refunded';

ALTER TABLE payment_transactions
    ADD CONSTRAINT chk_payment_transactions_status
        CHECK (status IN ('pending', 'paid', 'failed', 'expired', 'canceled')),
    ALTER COLUMN payment_provider SET DEFAULT 'duitku',
    ADD COLUMN provider_reference VARCHAR(120),
    ADD COLUMN payment_url TEXT;

DROP TABLE IF EXISTS payment_refund_reconciliation_events;
DROP TABLE IF EXISTS payment_refunds;

ALTER TABLE payment_transactions
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_refund_coins,
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_requested_refund_amount,
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_reported_refund_amount,
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_refund_amount,
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_refund_reconciliation_status,
    DROP CONSTRAINT IF EXISTS chk_payment_transactions_refund_status,
    DROP COLUMN IF EXISTS coins_written_off,
    DROP COLUMN IF EXISTS coins_reversed,
    DROP COLUMN IF EXISTS refund_reconciliation_reason,
    DROP COLUMN IF EXISTS refund_reconciliation_status,
    DROP COLUMN IF EXISTS refund_status,
    DROP COLUMN IF EXISTS provider_refund_amount_reported,
    DROP COLUMN IF EXISTS refund_requested_amount,
    DROP COLUMN IF EXISTS refunded_amount,
    DROP COLUMN IF EXISTS snap_token,
    DROP COLUMN IF EXISTS snap_redirect_url;
