SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `moderator_actions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `moderator_id` BIGINT NOT NULL,
    `action_type` VARCHAR(100) NOT NULL,
    `target_type` VARCHAR(50) NOT NULL,
    `target_id` BIGINT NOT NULL,
    `previous_state` TEXT,
    `new_state` TEXT,
    `reason` TEXT,
    `notes` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_moderator_actions_moderator_id_fkey` FOREIGN KEY (`moderator_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
