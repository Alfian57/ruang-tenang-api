SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `exp_histories` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `activity_type` VARCHAR(50) NOT NULL,
    `points` BIGINT NOT NULL,
    `description` VARCHAR(255),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_exp_histories_user_id` (`user_id`),
    CONSTRAINT `fk_exp_histories_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
