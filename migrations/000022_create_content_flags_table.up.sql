SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `content_flags` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `content_type` VARCHAR(50) NOT NULL,
    `content_id` BIGINT NOT NULL,
    `flag_type` VARCHAR(50) NOT NULL,
    `flag_category` VARCHAR(100) NOT NULL,
    `severity` VARCHAR(20) DEFAULT 'medium',
    `ai_confidence` DECIMAL(5,2),
    `ai_reason` TEXT,
    `flagged_by_id` BIGINT,
    `is_resolved` TINYINT(1) DEFAULT 0,
    `resolved_by_id` BIGINT,
    `resolved_at` DATETIME,
    `resolution_notes` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_content_flags_flagged_by_id_fkey` FOREIGN KEY (`flagged_by_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_content_flags_resolved_by_id_fkey` FOREIGN KEY (`resolved_by_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
