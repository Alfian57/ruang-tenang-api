SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_landmark_progress` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `landmark_id` VARCHAR(36) NOT NULL,
    `is_unlocked` TINYINT(1) DEFAULT 0,
    `current_value` BIGINT NOT NULL DEFAULT 0,
    `unlocked_at` DATETIME,
    `reward_claimed` TINYINT(1) DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_landmark_progress_user_id_landmark_id_key` (`user_id`, `landmark_id`),
    KEY `idx_user_landmark_progress_user_id` (`user_id`),
    CONSTRAINT `fk_user_landmark_progress_landmark_id_fkey` FOREIGN KEY (`landmark_id`) REFERENCES `map_landmarks` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_landmark_progress_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
