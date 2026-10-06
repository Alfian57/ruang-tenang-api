SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_phone_verifications` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `challenge_hash` VARCHAR(64) NOT NULL,
    `code_hash` VARCHAR(64) NOT NULL DEFAULT '',
    `attempts` BIGINT NOT NULL DEFAULT 0,
    `remember_me` TINYINT(1) NOT NULL DEFAULT 0,
    `expires_at` DATETIME NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_user_phone_verifications_user` (`user_id`),
    UNIQUE KEY `user_phone_verifications_challenge_hash_key` (`challenge_hash`),
    CONSTRAINT `fk_user_phone_verifications_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
