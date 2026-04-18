CREATE TABLE payment_transactions (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(100) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_type VARCHAR(20) NOT NULL,
    item_id INTEGER NOT NULL,
    item_name VARCHAR(150) NOT NULL,
    amount INTEGER NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_provider VARCHAR(20) NOT NULL DEFAULT 'midtrans',
    provider_transaction_id VARCHAR(120),
    provider_payment_type VARCHAR(50),
    snap_token TEXT,
    snap_redirect_url TEXT,
    callback_payload TEXT,
    failure_reason TEXT,
    paid_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_payment_transactions_item_type CHECK (item_type IN ('subscription', 'topup')),
    CONSTRAINT chk_payment_transactions_status CHECK (status IN ('pending', 'paid', 'failed', 'expired', 'canceled', 'refunded'))
);

CREATE INDEX idx_payment_transactions_user_id ON payment_transactions(user_id);
CREATE INDEX idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX idx_payment_transactions_created_at ON payment_transactions(created_at);
