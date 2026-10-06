SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `broadcast_notifications` (
    `id` CHAR(36) NOT NULL,
    `title` VARCHAR(255) NOT NULL,
    `body` TEXT NOT NULL,
    `icon` VARCHAR(500) DEFAULT '',
    `url` VARCHAR(500) DEFAULT '',
    `status` VARCHAR(20) NOT NULL DEFAULT 'draft',
    `scheduled_at` DATETIME,
    `sent_at` DATETIME,
    `sent_count` BIGINT NOT NULL DEFAULT 0,
    `failed_count` BIGINT NOT NULL DEFAULT 0,
    `created_by` BIGINT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_broadcast_scheduled_at` (`scheduled_at`),
    KEY `idx_broadcast_status` (`status`),
    CONSTRAINT `fk_fk_broadcast_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
