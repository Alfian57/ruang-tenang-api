SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_usage_daily_metrics` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `metric_date` DATE NOT NULL,
    `active_members` BIGINT NOT NULL DEFAULT 0,
    `invited_members` BIGINT NOT NULL DEFAULT 0,
    `pending_approvals` BIGINT NOT NULL DEFAULT 0,
    `contracted_seats` BIGINT NOT NULL DEFAULT 0,
    `used_seats` BIGINT NOT NULL DEFAULT 0,
    `messages_sent` BIGINT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_b2b_usage_daily_metrics_org_date` (`organization_id`, `metric_date`),
    CONSTRAINT `fk_b2b_usage_daily_metrics_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
