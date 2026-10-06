SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `organization_onboarding_templates` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `role` VARCHAR(20) NOT NULL,
    `title` VARCHAR(150) NOT NULL,
    `welcome_message` TEXT,
    `checklist_json` JSON NOT NULL DEFAULT (JSON_ARRAY()),
    `is_default` TINYINT(1) NOT NULL DEFAULT 0,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    `created_by` BIGINT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_org_onboarding_templates_active` (`organization_id`, `is_active`),
    UNIQUE KEY `uq_org_onboarding_templates_org_role` (`organization_id`, `role`),
    CONSTRAINT `fk_organization_onboarding_templates_created_by_fkey` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_organization_onboarding_templates_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
