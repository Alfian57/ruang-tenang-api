CREATE TABLE b2b_plans (
    id SERIAL PRIMARY KEY,
    code VARCHAR(60) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    description TEXT,
    billing_cycle VARCHAR(20) NOT NULL,
    base_price_per_seat BIGINT NOT NULL,
    min_seats INTEGER NOT NULL DEFAULT 1,
    max_seats INTEGER NOT NULL DEFAULT 100000,
    features_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_b2b_plans_billing_cycle CHECK (billing_cycle IN ('monthly', 'yearly')),
    CONSTRAINT chk_b2b_plans_base_price CHECK (base_price_per_seat >= 0),
    CONSTRAINT chk_b2b_plans_min_max_seats CHECK (min_seats > 0 AND max_seats >= min_seats)
);

CREATE INDEX idx_b2b_plans_active_cycle ON b2b_plans(is_active, billing_cycle);

INSERT INTO b2b_plans (code, name, description, billing_cycle, base_price_per_seat, min_seats, max_seats, features_json, is_active)
VALUES
    (
        'b2b-campus-monthly',
        'Campus Monthly',
        'Plan bulanan untuk kampus/sekolah dengan akses premium tim.',
        'monthly',
        12000,
        25,
        20000,
        '{"premium_chat": true, "dashboard": true, "bulk_invite": true}'::jsonb,
        TRUE
    ),
    (
        'b2b-campus-yearly',
        'Campus Yearly',
        'Plan tahunan untuk kampus/sekolah dengan harga seat lebih hemat.',
        'yearly',
        10000,
        25,
        20000,
        '{"premium_chat": true, "dashboard": true, "bulk_invite": true, "annual_discount": true}'::jsonb,
        TRUE
    )
ON CONFLICT (code) DO NOTHING;
