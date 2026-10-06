SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_wellness_profiles` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `initial_mood` VARCHAR(50) NOT NULL DEFAULT '',
    `goals_json` JSON NOT NULL DEFAULT (JSON_ARRAY()),
    `habits_json` JSON NOT NULL DEFAULT (JSON_ARRAY()),
    `tour_completed_at` DATETIME,
    `onboarding_completed_at` DATETIME,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_wellness_profiles_user_id_key` (`user_id`),
    CONSTRAINT `fk_user_wellness_profiles_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
