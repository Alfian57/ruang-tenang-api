SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `story_hearts` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `story_id` VARCHAR(36) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `story_hearts_story_id_user_id_key` (`story_id`, `user_id`),
    KEY `idx_story_hearts_story` (`story_id`),
    KEY `idx_story_hearts_user` (`user_id`),
    CONSTRAINT `fk_story_hearts_story_id_fkey` FOREIGN KEY (`story_id`) REFERENCES `inspiring_stories` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_story_hearts_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
