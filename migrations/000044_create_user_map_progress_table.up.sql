SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_map_progress` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `region_id` VARCHAR(36) NOT NULL,
    `is_unlocked` TINYINT(1) DEFAULT 0,
    `unlocked_at` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_map_progress_user_id_region_id_key` (`user_id`, `region_id`),
    KEY `idx_user_map_progress_user_id` (`user_id`),
    CONSTRAINT `fk_user_map_progress_region_id_fkey` FOREIGN KEY (`region_id`) REFERENCES `map_regions` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_map_progress_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
