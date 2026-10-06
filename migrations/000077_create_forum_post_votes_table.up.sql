SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `forum_post_votes` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `post_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `vote_type` VARCHAR(10) NOT NULL DEFAULT 'upvote',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_post_user_vote` (`post_id`, `user_id`),
    KEY `idx_forum_post_votes_post_id` (`post_id`),
    KEY `idx_forum_post_votes_type` (`vote_type`),
    KEY `idx_forum_post_votes_user_id` (`user_id`),
    CONSTRAINT `fk_forum_post_votes_post_id_fkey` FOREIGN KEY (`post_id`) REFERENCES `forum_posts` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_forum_post_votes_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
