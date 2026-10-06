SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_subscriptions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `plan_id` BIGINT NOT NULL,
    `source_order_id` VARCHAR(100),
    `status` VARCHAR(20) NOT NULL DEFAULT 'active',
    `starts_at` DATETIME NOT NULL,
    `ends_at` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_subscriptions_ends_at` (`ends_at`),
    KEY `idx_user_subscriptions_user_status` (`user_id`, `status`),
    CONSTRAINT `fk_user_subscriptions_plan_id_fkey` FOREIGN KEY (`plan_id`) REFERENCES `premium_plans` (`id`),
    CONSTRAINT `fk_user_subscriptions_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
