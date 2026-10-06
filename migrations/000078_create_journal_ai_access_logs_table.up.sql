SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `journal_ai_access_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `journal_id` BIGINT NOT NULL,
    `chat_session_id` BIGINT,
    `accessed_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `context_type` VARCHAR(50),
    PRIMARY KEY (`id`),
    KEY `idx_journal_ai_access_logs_journal_id` (`journal_id`),
    KEY `idx_journal_ai_access_logs_user_id` (`user_id`),
    CONSTRAINT `fk_journal_ai_access_logs_chat_session_id_fkey` FOREIGN KEY (`chat_session_id`) REFERENCES `chat_sessions` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_journal_ai_access_logs_journal_id_fkey` FOREIGN KEY (`journal_id`) REFERENCES `journals` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_journal_ai_access_logs_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
