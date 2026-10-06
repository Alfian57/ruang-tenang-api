SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_reminder_jobs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `subscription_id` BIGINT,
    `job_type` VARCHAR(50) NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending',
    `due_at` DATETIME NOT NULL,
    `payload_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `attempt_count` BIGINT NOT NULL DEFAULT 0,
    `last_error` TEXT,
    `sent_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_b2b_reminder_jobs_due_status` (`organization_id`, `status`, `due_at`),
    UNIQUE KEY `uq_b2b_reminder_jobs_org_type_due` (`organization_id`, `job_type`, `due_at`),
    CONSTRAINT `fk_b2b_reminder_jobs_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_b2b_reminder_jobs_subscription_id_fkey` FOREIGN KEY (`subscription_id`) REFERENCES `b2b_subscriptions` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
