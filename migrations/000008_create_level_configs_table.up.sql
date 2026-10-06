SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `level_configs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `level` BIGINT NOT NULL,
    `min_exp` BIGINT NOT NULL DEFAULT 0,
    `badge_name` VARCHAR(100) NOT NULL,
    `badge_icon` VARCHAR(255) NOT NULL,
    `tier_name` VARCHAR(50),
    `tier_color` VARCHAR(20),
    `description` TEXT,
    `task_description` TEXT,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `level_configs_level_key` (`level`),
    KEY `idx_level_configs_level` (`level`),
    KEY `idx_level_configs_min_exp` (`min_exp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
