SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_feature_unlocks` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `feature_id` VARCHAR(36) NOT NULL,
    `unlocked_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_feature_unlocks_user_id_feature_id_key` (`user_id`, `feature_id`),
    KEY `idx_user_feature_unlocks_feature` (`feature_id`),
    KEY `idx_user_feature_unlocks_user` (`user_id`),
    CONSTRAINT `fk_user_feature_unlocks_feature_id_fkey` FOREIGN KEY (`feature_id`) REFERENCES `feature_definitions` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_feature_unlocks_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
