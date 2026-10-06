SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `forums` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `category_id` BIGINT,
    `title` VARCHAR(255) NOT NULL,
    `slug` VARCHAR(300) NOT NULL,
    `content` TEXT NOT NULL,
    `trigger_warnings` JSON,
    `is_flagged` TINYINT(1) NOT NULL DEFAULT 0,
    `flagged_reason` TEXT,
    `has_accepted_answer` TINYINT(1) DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    KEY `idx_forums_category_id` (`category_id`),
    KEY `idx_forums_deleted_at` (`deleted_at`),
    KEY `idx_forums_has_accepted` (`has_accepted_answer`),
    KEY `idx_forums_is_flagged` (`is_flagged`),
    UNIQUE KEY `idx_forums_slug` (`slug`),
    KEY `idx_forums_title` (`title`),
    KEY `idx_forums_user_id` (`user_id`),
    CONSTRAINT `fk_forums_category_id_fkey` FOREIGN KEY (`category_id`) REFERENCES `forum_categories` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_forums_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
