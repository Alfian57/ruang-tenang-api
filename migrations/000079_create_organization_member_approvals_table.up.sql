SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `organization_member_approvals` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `organization_member_id` BIGINT NOT NULL,
    `requested_by` BIGINT,
    `approver_user_id` BIGINT,
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending',
    `note` TEXT,
    `decided_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_org_member_approvals_org_member` (`organization_id`, `organization_member_id`),
    KEY `idx_org_member_approvals_status` (`status`),
    CONSTRAINT `fk_organization_member_approvals_approver_user_id_fkey` FOREIGN KEY (`approver_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_organization_member_approvals_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_organization_member_approvals_organization_member_id_fkey` FOREIGN KEY (`organization_member_id`) REFERENCES `organization_members` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_organization_member_approvals_requested_by_fkey` FOREIGN KEY (`requested_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
