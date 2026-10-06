SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `payment_webhook_events` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `provider` VARCHAR(20) NOT NULL,
    `order_id` VARCHAR(100) NOT NULL,
    `event_key` VARCHAR(255) NOT NULL,
    `payload` TEXT NOT NULL,
    `processed_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `payment_webhook_events_event_key_key` (`event_key`),
    KEY `idx_payment_webhook_events_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
