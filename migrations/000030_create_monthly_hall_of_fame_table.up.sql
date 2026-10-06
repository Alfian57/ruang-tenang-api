SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `monthly_hall_of_fame` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `level` BIGINT NOT NULL,
    `month` BIGINT NOT NULL,
    `year` BIGINT NOT NULL,
    `rank` BIGINT NOT NULL,
    `monthly_xp` BIGINT NOT NULL,
    `message` VARCHAR(300),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `monthly_hall_of_fame_level_month_year_rank_key` (`level`, `month`, `year`, `rank`),
    KEY `idx_hall_of_fame_level_month` (`level`, `month`, `year`),
    KEY `idx_hall_of_fame_user` (`user_id`),
    CONSTRAINT `fk_monthly_hall_of_fame_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
