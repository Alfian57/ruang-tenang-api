SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `wellness_plans` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `profile_id` BIGINT,
    `title` VARCHAR(180) NOT NULL,
    `summary` TEXT NOT NULL DEFAULT (''),
    `status` VARCHAR(30) NOT NULL DEFAULT 'active',
    `starts_on` DATE NOT NULL,
    `ends_on` DATE NOT NULL,
    `generated_from_mood` VARCHAR(50) NOT NULL DEFAULT '',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    `active_user_key` BIGINT GENERATED ALWAYS AS (CASE WHEN status = 'active' THEN user_id END) VIRTUAL,
    UNIQUE KEY `uq_wellness_plans_user_active` (`active_user_key`),
    KEY `idx_wellness_plans_user_dates` (`user_id`, `starts_on`, `ends_on`),
    CONSTRAINT `fk_wellness_plans_profile_id_fkey` FOREIGN KEY (`profile_id`) REFERENCES `user_wellness_profiles` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_wellness_plans_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
