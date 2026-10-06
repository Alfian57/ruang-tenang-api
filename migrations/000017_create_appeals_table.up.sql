SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `appeals` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `reason` TEXT NOT NULL,
    `evidence` TEXT,
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending',
    `reviewer_notes` TEXT,
    `reviewer_id` BIGINT,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_appeals_user_status` (`user_id`, `status`),
    CONSTRAINT `fk_appeals_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_appeals_reviewer` FOREIGN KEY (`reviewer_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
