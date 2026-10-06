SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `story_categories` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `name` VARCHAR(100) NOT NULL,
    `slug` VARCHAR(100) NOT NULL,
    `description` TEXT,
    `icon` VARCHAR(50) DEFAULT '📖',
    `display_order` BIGINT DEFAULT 0,
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `story_categories_name_key` (`name`),
    UNIQUE KEY `story_categories_slug_key` (`slug`),
    KEY `idx_story_categories_active` (`is_active`),
    KEY `idx_story_categories_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
