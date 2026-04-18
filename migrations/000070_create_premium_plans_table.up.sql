CREATE TABLE premium_plans (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL,
    duration_days INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO premium_plans (code, name, description, price, duration_days, is_active)
VALUES
    ('premium_monthly', 'Premium Bulanan', 'Akses premium selama 30 hari', 29900, 30, TRUE),
    ('premium_yearly', 'Premium Tahunan', 'Akses premium selama 365 hari', 299000, 365, TRUE)
ON CONFLICT (code) DO NOTHING;
