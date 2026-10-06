SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `organization_members` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `user_id` BIGINT,
    `email` VARCHAR(255) NOT NULL,
    `full_name` VARCHAR(150),
    `role` VARCHAR(20) NOT NULL DEFAULT 'member',
    `status` VARCHAR(20) NOT NULL DEFAULT 'invited',
    `invitation_token` VARCHAR(120),
    `invitation_expires_at` DATETIME,
    `invited_by` BIGINT,
    `invited_at` DATETIME,
    `joined_at` DATETIME,
    `removed_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_organization_members_org_email` (`organization_id`, `email`),
    UNIQUE KEY `uq_organization_members_org_user` (`organization_id`, `user_id`),
    KEY `idx_organization_members_org_status` (`organization_id`, `status`),
    KEY `idx_organization_members_user_status` (`user_id`, `status`),
    UNIQUE KEY `uq_organization_members_invitation_token` (`invitation_token`),
    CONSTRAINT `fk_organization_members_invited_by_fkey` FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_organization_members_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_organization_members_user_id_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
