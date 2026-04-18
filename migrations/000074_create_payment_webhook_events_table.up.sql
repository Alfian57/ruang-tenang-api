CREATE TABLE payment_webhook_events (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(20) NOT NULL,
    order_id VARCHAR(100) NOT NULL,
    event_key VARCHAR(255) NOT NULL UNIQUE,
    payload TEXT NOT NULL,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_webhook_events_order_id ON payment_webhook_events(order_id);
