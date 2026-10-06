SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `journals` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `title` VARCHAR(255),
    `content` TEXT NOT NULL,
    `summary` TEXT,
    `mood_id` BIGINT,
    `tags` JSON,
    `is_private` TINYINT(1) NOT NULL DEFAULT 1,
    `share_with_ai` TINYINT(1) NOT NULL DEFAULT 0,
    `ai_accessed_at` DATETIME,
    `word_count` BIGINT DEFAULT 0,
    `sentiment_score` DECIMAL(3,2),
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_journals_user_id` (`user_id`),
    KEY `idx_journals_user_id_share_with_ai` (`user_id`, `share_with_ai`),
    UNIQUE KEY `idx_journals_uuid` (`uuid`),
    CONSTRAINT `fk_journals_mood_id_fkey` FOREIGN KEY (`mood_id`) REFERENCES `user_moods` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_journals_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
