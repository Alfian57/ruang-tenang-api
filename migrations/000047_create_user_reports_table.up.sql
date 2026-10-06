SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_reports` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `reporter_id` BIGINT NOT NULL,
    `report_type` VARCHAR(50) NOT NULL,
    `reported_content_id` BIGINT,
    `reported_user_id` BIGINT,
    `reason` VARCHAR(100) NOT NULL,
    `description` TEXT,
    `status` VARCHAR(50) DEFAULT 'pending',
    `handled_by_id` BIGINT,
    `handled_at` DATETIME,
    `action_taken` VARCHAR(100),
    `moderator_notes` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_user_reports_handled_by_id_fkey` FOREIGN KEY (`handled_by_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_user_reports_reported_user_id_fkey` FOREIGN KEY (`reported_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_user_reports_reporter_id_fkey` FOREIGN KEY (`reporter_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
