SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_activities` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `activity_type` VARCHAR(50) NOT NULL,
    `description` TEXT,
    `exp_gained` BIGINT DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `date` DATE NOT NULL DEFAULT (CURRENT_DATE),
    `count` BIGINT NOT NULL DEFAULT 0,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_activities_activity_type` (`activity_type`),
    KEY `idx_user_activities_created_at` (`created_at`),
    KEY `idx_user_activities_date` (`date`),
    KEY `idx_user_activities_user_id` (`user_id`),
    UNIQUE KEY `uq_user_activities_user_type_date` (`user_id`, `activity_type`, `date`),
    CONSTRAINT `fk_user_activities_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
