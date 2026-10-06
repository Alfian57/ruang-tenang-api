SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `story_comments` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `story_id` VARCHAR(36) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `content` VARCHAR(500) NOT NULL,
    `heart_count` BIGINT DEFAULT 0,
    `is_hidden` TINYINT(1) DEFAULT 0,
    `hidden_reason` VARCHAR(255),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_story_comments_story` (`story_id`),
    KEY `idx_story_comments_user` (`user_id`),
    CONSTRAINT `fk_story_comments_story_id_fkey` FOREIGN KEY (`story_id`) REFERENCES `inspiring_stories` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_story_comments_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
