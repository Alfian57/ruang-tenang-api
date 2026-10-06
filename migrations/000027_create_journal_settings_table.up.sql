SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `journal_settings` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `allow_ai_access` TINYINT(1) NOT NULL DEFAULT 0,
    `is_blocked` TINYINT(1) DEFAULT 0,
    `ai_context_days` BIGINT DEFAULT 7,
    `ai_context_max_entries` BIGINT DEFAULT 5,
    `default_share_with_ai` TINYINT(1) DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `journal_settings_user_id_key` (`user_id`),
    KEY `idx_journal_settings_user_id` (`user_id`),
    CONSTRAINT `fk_journal_settings_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
