SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `reward_claims` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `reward_id` BIGINT NOT NULL,
    `coin_spent` BIGINT NOT NULL,
    `claimed_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_reward_claims_reward_id` (`reward_id`),
    KEY `idx_reward_claims_user_id` (`user_id`),
    CONSTRAINT `fk_reward_claims_reward_id_fkey` FOREIGN KEY (`reward_id`) REFERENCES `rewards` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_reward_claims_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
