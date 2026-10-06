SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `weekly_insight_snapshots` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `week_start` DATE NOT NULL,
    `week_end` DATE NOT NULL,
    `mood_summary_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `activity_summary_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `insight_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `premium_sections_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `narrative` TEXT NOT NULL DEFAULT (''),
    `recommendations_json` JSON NOT NULL DEFAULT (JSON_ARRAY()),
    `is_ai_enhanced` TINYINT(1) NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `weekly_insight_snapshots_user_id_week_start_key` (`user_id`, `week_start`),
    CONSTRAINT `fk_weekly_insight_snapshots_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
