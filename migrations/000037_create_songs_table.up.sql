SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `songs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `title` VARCHAR(255) NOT NULL,
    `slug` VARCHAR(300) NOT NULL,
    `file_path` VARCHAR(500) NOT NULL,
    `thumbnail` VARCHAR(500),
    `song_category_id` BIGINT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    `attribution` VARCHAR(255) NOT NULL DEFAULT '',
    `source_url` VARCHAR(500) NOT NULL DEFAULT '',
    `license_url` VARCHAR(500) NOT NULL DEFAULT '',
    PRIMARY KEY (`id`),
    KEY `idx_songs_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_songs_slug` (`slug`),
    KEY `idx_songs_song_category_id` (`song_category_id`),
    KEY `idx_songs_title` (`title`),
    CONSTRAINT `fk_songs_song_category_id_fkey` FOREIGN KEY (`song_category_id`) REFERENCES `song_categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
