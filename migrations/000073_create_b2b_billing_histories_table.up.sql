SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_billing_histories` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `subscription_id` BIGINT NOT NULL,
    `invoice_number` VARCHAR(100) NOT NULL,
    `billing_period_start` DATETIME NOT NULL,
    `billing_period_end` DATETIME NOT NULL,
    `seats_billed` BIGINT NOT NULL,
    `amount` BIGINT NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending',
    `paid_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `b2b_billing_histories_invoice_number_key` (`invoice_number`),
    KEY `idx_b2b_billing_histories_subscription_status` (`subscription_id`, `status`),
    CONSTRAINT `fk_b2b_billing_histories_subscription_id_fkey` FOREIGN KEY (`subscription_id`) REFERENCES `b2b_subscriptions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
