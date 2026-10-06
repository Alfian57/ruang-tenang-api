SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `playlists` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `thumbnail` VARCHAR(500),
    `is_public` TINYINT(1) DEFAULT 0,
    `is_admin_playlist` TINYINT(1) DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    KEY `idx_playlists_deleted_at` (`deleted_at`),
    KEY `idx_playlists_user_id` (`user_id`),
    UNIQUE KEY `idx_playlists_uuid` (`uuid`),
    CONSTRAINT `fk_playlists_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
