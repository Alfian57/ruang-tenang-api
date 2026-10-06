SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `rewards` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `image` VARCHAR(500) DEFAULT '',
    `coin_cost` BIGINT NOT NULL,
    `reward_type` VARCHAR(50) NOT NULL DEFAULT 'general',
    `reward_value` VARCHAR(100) DEFAULT '',
    `stock` BIGINT NOT NULL DEFAULT '-1',
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_rewards_is_active` (`is_active`),
    KEY `idx_rewards_reward_type` (`reward_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
