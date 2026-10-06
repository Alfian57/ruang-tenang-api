SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `daily_tasks` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `task_type` VARCHAR(50) NOT NULL,
    `task_date` DATE NOT NULL,
    `target_count` BIGINT NOT NULL DEFAULT 1,
    `current_count` BIGINT NOT NULL DEFAULT 0,
    `is_completed` TINYINT(1) DEFAULT 0,
    `is_claimed` TINYINT(1) DEFAULT 0,
    `xp_reward` BIGINT NOT NULL DEFAULT 0,
    `coin_reward` BIGINT NOT NULL DEFAULT 0,
    `completed_at` DATETIME,
    `claimed_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_daily_task_user_type_date` (`user_id`, `task_type`, `task_date`),
    KEY `idx_daily_tasks_claimed` (`is_claimed`),
    KEY `idx_daily_tasks_completed` (`is_completed`),
    KEY `idx_daily_tasks_task_date` (`task_date`),
    KEY `idx_daily_tasks_user_date` (`user_id`, `task_date`),
    KEY `idx_daily_tasks_user_id` (`user_id`),
    CONSTRAINT `fk_daily_tasks_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
