SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_subscriptions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `plan_id` BIGINT NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'draft',
    `contracted_seats` BIGINT NOT NULL,
    `used_seats` BIGINT NOT NULL DEFAULT 0,
    `billing_cycle` VARCHAR(20) NOT NULL,
    `unit_price` BIGINT NOT NULL,
    `subtotal` BIGINT NOT NULL,
    `discount_amount` BIGINT NOT NULL DEFAULT 0,
    `total_amount` BIGINT NOT NULL,
    `starts_at` DATETIME NOT NULL,
    `ends_at` DATETIME NOT NULL,
    `activated_at` DATETIME,
    `metadata_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_b2b_subscriptions_org_active_period` (`organization_id`, `status`, `starts_at`, `ends_at`),
    KEY `idx_b2b_subscriptions_org_status` (`organization_id`, `status`),
    KEY `idx_b2b_subscriptions_period` (`starts_at`, `ends_at`),
    CONSTRAINT `fk_b2b_subscriptions_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_b2b_subscriptions_plan_id_fkey` FOREIGN KEY (`plan_id`) REFERENCES `b2b_plans` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
