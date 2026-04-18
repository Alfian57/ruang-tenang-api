CREATE TABLE topup_packages (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    coins BIGINT NOT NULL,
    bonus_coins BIGINT NOT NULL DEFAULT 0,
    price INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO topup_packages (code, name, coins, bonus_coins, price, is_active)
VALUES
    ('topup_100', 'Topup 100 Koin', 100, 0, 15000, TRUE),
    ('topup_250', 'Topup 250 Koin', 250, 25, 35000, TRUE),
    ('topup_500', 'Topup 500 Koin', 500, 75, 65000, TRUE)
ON CONFLICT (code) DO NOTHING;
