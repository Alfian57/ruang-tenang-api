SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `organizations` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `code` VARCHAR(60) NOT NULL,
    `name` VARCHAR(150) NOT NULL,
    `business_type` VARCHAR(50) NOT NULL DEFAULT 'general',
    `contact_email` VARCHAR(255) NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'active',
    `requires_member_approval` TINYINT(1) NOT NULL DEFAULT 1,
    `created_by` BIGINT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `organizations_code_key` (`code`),
    KEY `idx_organizations_status` (`status`),
    CONSTRAINT `fk_organizations_created_by_fkey` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
