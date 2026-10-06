SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_combos` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `combo_count` BIGINT NOT NULL DEFAULT 0,
    `multiplier` DECIMAL(3,1) NOT NULL DEFAULT 1.0,
    `last_activity_type` VARCHAR(50),
    `last_activity_at` DATETIME,
    `session_started_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_combos_user_id_key` (`user_id`),
    KEY `idx_user_combos_user` (`user_id`),
    CONSTRAINT `fk_user_combos_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
