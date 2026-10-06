SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_moods` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `mood` VARCHAR(50) NOT NULL,
    `note` TEXT,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    KEY `idx_user_moods_created_at` (`created_at`),
    KEY `idx_user_moods_deleted_at` (`deleted_at`),
    KEY `idx_user_moods_mood` (`mood`),
    KEY `idx_user_moods_user_id` (`user_id`),
    CONSTRAINT `fk_user_moods_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
