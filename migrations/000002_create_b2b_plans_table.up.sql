SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_plans` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `code` VARCHAR(60) NOT NULL,
    `name` VARCHAR(120) NOT NULL,
    `description` TEXT,
    `billing_cycle` VARCHAR(20) NOT NULL,
    `base_price_per_seat` BIGINT NOT NULL,
    `min_seats` BIGINT NOT NULL DEFAULT 1,
    `max_seats` BIGINT NOT NULL DEFAULT 100000,
    `features_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `b2b_plans_code_key` (`code`),
    KEY `idx_b2b_plans_active_cycle` (`is_active`, `billing_cycle`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
