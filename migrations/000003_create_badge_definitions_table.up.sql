SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `badge_definitions` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `badge_key` VARCHAR(100) NOT NULL,
    `badge_name` VARCHAR(200) NOT NULL,
    `description` TEXT,
    `icon` VARCHAR(50),
    `category` VARCHAR(50) DEFAULT 'general',
    `requirement_type` VARCHAR(50) NOT NULL,
    `requirement_value` BIGINT DEFAULT 0,
    `is_active` TINYINT(1) DEFAULT 1,
    `display_order` BIGINT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `badge_definitions_badge_key_key` (`badge_key`),
    KEY `idx_badge_definitions_category` (`category`),
    KEY `idx_badge_definitions_requirement_type` (`requirement_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
