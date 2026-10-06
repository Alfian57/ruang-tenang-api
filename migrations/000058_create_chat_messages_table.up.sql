SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `chat_messages` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `chat_session_id` BIGINT NOT NULL,
    `role` VARCHAR(20) NOT NULL,
    `content` TEXT NOT NULL,
    `is_pinned` TINYINT(1) DEFAULT 0,
    `type` VARCHAR(20) DEFAULT 'text',
    `is_liked` TINYINT(1) DEFAULT 0,
    `is_disliked` TINYINT(1) DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_chat_messages_chat_session_id` (`chat_session_id`),
    KEY `idx_chat_messages_pinned` (`chat_session_id`, `is_pinned`),
    KEY `idx_chat_messages_role` (`role`),
    UNIQUE KEY `idx_chat_messages_uuid` (`uuid`),
    CONSTRAINT `fk_chat_messages_chat_session_id_fkey` FOREIGN KEY (`chat_session_id`) REFERENCES `chat_sessions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
