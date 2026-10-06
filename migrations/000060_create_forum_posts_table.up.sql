SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `forum_posts` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `forum_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `content` TEXT NOT NULL,
    `is_flagged` TINYINT(1) NOT NULL DEFAULT 0,
    `flagged_reason` TEXT,
    `is_accepted_answer` TINYINT(1) DEFAULT 0,
    `is_community_favorite` TINYINT(1) DEFAULT 0,
    `upvotes_count` BIGINT DEFAULT 0,
    `downvotes_count` BIGINT DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    KEY `idx_forum_posts_accepted` (`forum_id`, `is_accepted_answer`),
    KEY `idx_forum_posts_community_fav` (`forum_id`, `is_community_favorite`),
    KEY `idx_forum_posts_deleted_at` (`deleted_at`),
    KEY `idx_forum_posts_forum_id` (`forum_id`),
    KEY `idx_forum_posts_is_flagged` (`is_flagged`),
    KEY `idx_forum_posts_user_id` (`user_id`),
    CONSTRAINT `fk_forum_posts_forum_id_fkey` FOREIGN KEY (`forum_id`) REFERENCES `forums` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_forum_posts_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
