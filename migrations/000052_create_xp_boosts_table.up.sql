SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `xp_boosts` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `multiplier` DECIMAL(3,1) NOT NULL DEFAULT 2.0,
    `trigger_type` VARCHAR(50) NOT NULL DEFAULT 'activity_chain',
    `started_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expires_at` DATETIME NOT NULL,
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_xp_boosts_expires` (`expires_at`),
    KEY `idx_xp_boosts_user_active` (`user_id`, `is_active`),
    CONSTRAINT `fk_xp_boosts_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
