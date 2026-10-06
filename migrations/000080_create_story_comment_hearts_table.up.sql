SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `story_comment_hearts` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `comment_id` VARCHAR(36) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `story_comment_hearts_comment_id_user_id_key` (`comment_id`, `user_id`),
    KEY `idx_story_comment_hearts_comment` (`comment_id`),
    KEY `idx_story_comment_hearts_user` (`user_id`),
    CONSTRAINT `fk_story_comment_hearts_comment_id_fkey` FOREIGN KEY (`comment_id`) REFERENCES `story_comments` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_story_comment_hearts_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
