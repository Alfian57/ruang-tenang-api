SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_feature_usages` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `feature_key` VARCHAR(50) NOT NULL,
    `usage_date` DATE NOT NULL,
    `used_count` BIGINT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `usage_window_start` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_user_feature_usages_window` (`user_id`, `feature_key`, `usage_window_start`),
    KEY `idx_user_feature_usages_feature_date` (`feature_key`, `usage_date`),
    KEY `idx_user_feature_usages_feature_window` (`feature_key`, `usage_window_start`),
    CONSTRAINT `fk_user_feature_usages_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
