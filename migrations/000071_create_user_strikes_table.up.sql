SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_strikes` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `report_id` BIGINT,
    `reason` VARCHAR(255) NOT NULL,
    `severity` VARCHAR(20) DEFAULT 'warning',
    `issued_by_id` BIGINT NOT NULL,
    `expires_at` DATETIME,
    `is_active` TINYINT(1) DEFAULT 1,
    `notes` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_user_strikes_issued_by_id_fkey` FOREIGN KEY (`issued_by_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_strikes_report_id_fkey` FOREIGN KEY (`report_id`) REFERENCES `user_reports` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_user_strikes_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
