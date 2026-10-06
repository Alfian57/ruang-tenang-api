SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `organization_audit_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `actor_user_id` BIGINT,
    `action` VARCHAR(80) NOT NULL,
    `entity_type` VARCHAR(80) NOT NULL,
    `entity_id` VARCHAR(120),
    `metadata_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_organization_audit_logs_action` (`action`),
    CONSTRAINT `fk_organization_audit_logs_actor_user_id_fkey` FOREIGN KEY (`actor_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_organization_audit_logs_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
