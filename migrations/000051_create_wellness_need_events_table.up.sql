SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `wellness_need_events` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `condition` VARCHAR(50) NOT NULL,
    `recommendations_json` JSON NOT NULL DEFAULT (JSON_ARRAY()),
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_wellness_need_events_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
