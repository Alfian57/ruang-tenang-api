SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_sso_configs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `provider` VARCHAR(30),
    `issuer_url` VARCHAR(500),
    `entrypoint_url` VARCHAR(500),
    `audience` VARCHAR(255),
    `certificate_pem` TEXT,
    `is_enabled` TINYINT(1) NOT NULL DEFAULT 0,
    `enforce_sso` TINYINT(1) NOT NULL DEFAULT 0,
    `metadata_json` JSON NOT NULL DEFAULT (JSON_OBJECT()),
    `created_by` BIGINT,
    `updated_by` BIGINT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `b2b_sso_configs_organization_id_key` (`organization_id`),
    KEY `idx_b2b_sso_configs_enabled` (`is_enabled`),
    CONSTRAINT `fk_b2b_sso_configs_created_by_fkey` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_b2b_sso_configs_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_b2b_sso_configs_updated_by_fkey` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
