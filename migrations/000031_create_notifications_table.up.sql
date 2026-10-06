SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `notifications` (
    `id` CHAR(36) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `type` VARCHAR(50) NOT NULL,
    `title` VARCHAR(255) NOT NULL,
    `message` VARCHAR(500) NOT NULL,
    `is_read` TINYINT(1) DEFAULT 0,
    `data` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_notification_user_read` (`user_id`, `is_read`),
    CONSTRAINT `fk_notifications_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
