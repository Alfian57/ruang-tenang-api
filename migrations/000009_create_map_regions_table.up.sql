SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `map_regions` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `region_key` VARCHAR(100) NOT NULL,
    `name` VARCHAR(200) NOT NULL,
    `description` TEXT,
    `icon` VARCHAR(50),
    `image` VARCHAR(500),
    `unlock_type` VARCHAR(50) NOT NULL DEFAULT 'level',
    `unlock_value` BIGINT NOT NULL DEFAULT 1,
    `position_x` BIGINT NOT NULL DEFAULT 0,
    `position_y` BIGINT NOT NULL DEFAULT 0,
    `display_order` BIGINT NOT NULL DEFAULT 0,
    `parent_region_id` VARCHAR(36),
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `map_regions_region_key_key` (`region_key`),
    CONSTRAINT `fk_map_regions_parent_region_id_fkey` FOREIGN KEY (`parent_region_id`) REFERENCES `map_regions` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
