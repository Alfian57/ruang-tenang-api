SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `forum_post_reports` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `post_id` BIGINT NOT NULL,
    `reporter_id` BIGINT NOT NULL,
    `reason` VARCHAR(50) NOT NULL,
    `description` TEXT,
    `status` VARCHAR(20) DEFAULT 'pending',
    `reviewed_by` BIGINT,
    `reviewed_at` DATETIME,
    `moderator_notes` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    `pending_report_key` VARCHAR(64) GENERATED ALWAYS AS (CASE WHEN status = 'pending' THEN CONCAT(post_id, '-', reporter_id) END) VIRTUAL,
    UNIQUE KEY `idx_forum_post_reports_unique` (`pending_report_key`),
    KEY `idx_forum_post_reports_post_id` (`post_id`),
    KEY `idx_forum_post_reports_reason` (`reason`),
    KEY `idx_forum_post_reports_reporter_id` (`reporter_id`),
    KEY `idx_forum_post_reports_status` (`status`),
    CONSTRAINT `fk_forum_post_reports_post_id_fkey` FOREIGN KEY (`post_id`) REFERENCES `forum_posts` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_forum_post_reports_reporter_id_fkey` FOREIGN KEY (`reporter_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_forum_post_reports_reviewed_by_fkey` FOREIGN KEY (`reviewed_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
