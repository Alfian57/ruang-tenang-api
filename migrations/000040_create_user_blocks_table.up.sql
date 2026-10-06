SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_blocks` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `blocker_id` BIGINT NOT NULL,
    `blocked_id` BIGINT NOT NULL,
    `reason` VARCHAR(255),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_blocks_blocker_id_blocked_id_key` (`blocker_id`, `blocked_id`),
    CONSTRAINT `fk_user_blocks_blocked_id_fkey` FOREIGN KEY (`blocked_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_blocks_blocker_id_fkey` FOREIGN KEY (`blocker_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
