SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `map_landmarks` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `region_id` VARCHAR(36) NOT NULL,
    `landmark_key` VARCHAR(100) NOT NULL,
    `name` VARCHAR(200) NOT NULL,
    `description` TEXT,
    `icon` VARCHAR(50),
    `unlock_type` VARCHAR(50) NOT NULL DEFAULT 'activity_count',
    `unlock_activity` VARCHAR(50),
    `unlock_value` BIGINT NOT NULL DEFAULT 1,
    `position_x` BIGINT NOT NULL DEFAULT 0,
    `position_y` BIGINT NOT NULL DEFAULT 0,
    `xp_reward` BIGINT NOT NULL DEFAULT 0,
    `coin_reward` BIGINT NOT NULL DEFAULT 0,
    `display_order` BIGINT NOT NULL DEFAULT 0,
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `map_landmarks_landmark_key_key` (`landmark_key`),
    KEY `idx_map_landmarks_region_id` (`region_id`),
    CONSTRAINT `fk_map_landmarks_region_id_fkey` FOREIGN KEY (`region_id`) REFERENCES `map_regions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
