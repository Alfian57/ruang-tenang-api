SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `forum_likes` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `forum_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `forum_likes_forum_id_user_id_key` (`forum_id`, `user_id`),
    KEY `idx_forum_likes_forum_id` (`forum_id`),
    KEY `idx_forum_likes_user_id` (`user_id`),
    CONSTRAINT `fk_forum_likes_forum_id_fkey` FOREIGN KEY (`forum_id`) REFERENCES `forums` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_forum_likes_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
