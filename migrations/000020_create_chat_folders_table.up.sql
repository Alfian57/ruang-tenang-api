SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `chat_folders` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `user_id` BIGINT NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `color` VARCHAR(7) DEFAULT '#6366f1',
    `icon` VARCHAR(50) DEFAULT 'folder',
    `position` BIGINT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_chat_folders_user` (`user_id`),
    UNIQUE KEY `idx_chat_folders_uuid` (`uuid`),
    CONSTRAINT `fk_chat_folders_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
