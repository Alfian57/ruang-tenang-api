SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `wellness_plan_items` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `plan_id` VARCHAR(36) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `day_number` BIGINT NOT NULL,
    `item_date` DATE NOT NULL,
    `title` VARCHAR(180) NOT NULL,
    `description` TEXT NOT NULL DEFAULT (''),
    `action_type` VARCHAR(50) NOT NULL,
    `route` VARCHAR(255) NOT NULL DEFAULT '',
    `status` VARCHAR(30) NOT NULL DEFAULT 'pending',
    `completed_at` DATETIME,
    `metadata_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_wellness_plan_items_plan_day` (`plan_id`, `day_number`),
    KEY `idx_wellness_plan_items_user_date` (`user_id`, `item_date`),
    CONSTRAINT `fk_wellness_plan_items_plan_id_fkey` FOREIGN KEY (`plan_id`) REFERENCES `wellness_plans` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_wellness_plan_items_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
