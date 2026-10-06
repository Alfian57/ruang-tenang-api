SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_badges` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `badge_id` VARCHAR(36) NOT NULL,
    `earned_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `is_showcased` TINYINT(1) DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_badges_user_id_badge_id_key` (`user_id`, `badge_id`),
    KEY `idx_user_badges_badge` (`badge_id`),
    KEY `idx_user_badges_showcased` (`user_id`, `is_showcased`),
    KEY `idx_user_badges_user` (`user_id`),
    CONSTRAINT `fk_user_badges_badge_id_fkey` FOREIGN KEY (`badge_id`) REFERENCES `badge_definitions` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_badges_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
