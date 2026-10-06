SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `feature_definitions` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `feature_key` VARCHAR(100) NOT NULL,
    `feature_name` VARCHAR(200) NOT NULL,
    `description` TEXT,
    `icon` VARCHAR(50),
    `required_level` BIGINT NOT NULL DEFAULT 1,
    `category` VARCHAR(50) DEFAULT 'general',
    `is_active` TINYINT(1) DEFAULT 1,
    `display_order` BIGINT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `feature_definitions_feature_key_key` (`feature_key`),
    KEY `idx_feature_definitions_category` (`category`),
    KEY `idx_feature_definitions_level` (`required_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
