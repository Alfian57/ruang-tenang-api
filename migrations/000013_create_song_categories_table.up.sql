SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `song_categories` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `slug` VARCHAR(300) NOT NULL,
    `thumbnail` VARCHAR(500),
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `song_categories_name_key` (`name`),
    KEY `idx_song_categories_deleted_at` (`deleted_at`),
    KEY `idx_song_categories_name` (`name`),
    UNIQUE KEY `idx_song_categories_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
