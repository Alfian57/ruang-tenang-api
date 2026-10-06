SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `push_subscriptions` (
    `id` CHAR(36) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `endpoint` VARCHAR(768) NOT NULL,
    `p256dh` TEXT NOT NULL,
    `auth` TEXT NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `push_subscriptions_endpoint_key` (`endpoint`),
    KEY `idx_push_sub_user` (`user_id`),
    CONSTRAINT `fk_push_subscriptions_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
