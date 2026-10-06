SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `community_stats` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `month` BIGINT NOT NULL,
    `year` BIGINT NOT NULL,
    `total_xp_earned` BIGINT DEFAULT 0,
    `active_members` BIGINT DEFAULT 0,
    `total_achievements` BIGINT DEFAULT 0,
    `new_members` BIGINT DEFAULT 0,
    `total_stories_published` BIGINT DEFAULT 0,
    `total_articles_published` BIGINT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `community_stats_month_year_key` (`month`, `year`),
    KEY `idx_community_stats_period` (`year`, `month`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
