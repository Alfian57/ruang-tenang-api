SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `playlist_items` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `playlist_id` BIGINT NOT NULL,
    `song_id` BIGINT NOT NULL,
    `position` BIGINT NOT NULL DEFAULT 0,
    `added_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_playlist_song` (`playlist_id`, `song_id`),
    KEY `idx_playlist_items_deleted_at` (`deleted_at`),
    KEY `idx_playlist_items_playlist_id` (`playlist_id`),
    KEY `idx_playlist_items_song_id` (`song_id`),
    UNIQUE KEY `idx_playlist_items_uuid` (`uuid`),
    CONSTRAINT `fk_playlist_items_playlist_id_fkey` FOREIGN KEY (`playlist_id`) REFERENCES `playlists` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_playlist_items_song_id_fkey` FOREIGN KEY (`song_id`) REFERENCES `songs` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
