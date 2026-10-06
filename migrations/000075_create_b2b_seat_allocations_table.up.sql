SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_seat_allocations` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `subscription_id` BIGINT NOT NULL,
    `organization_member_id` BIGINT NOT NULL,
    `allocated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `released_at` DATETIME,
    `release_reason` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    `active_allocation_key` VARCHAR(64) GENERATED ALWAYS AS (CASE WHEN released_at IS NULL THEN CONCAT(subscription_id, '-', organization_member_id) END) VIRTUAL,
    UNIQUE KEY `uq_b2b_seat_allocations_subscription_member_active` (`active_allocation_key`),
    KEY `idx_b2b_seat_allocations_member` (`organization_member_id`),
    KEY `idx_b2b_seat_allocations_subscription_active` (`subscription_id`, `released_at`),
    CONSTRAINT `fk_b2b_seat_allocations_organization_member_id_fkey` FOREIGN KEY (`organization_member_id`) REFERENCES `organization_members` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_b2b_seat_allocations_subscription_id_fkey` FOREIGN KEY (`subscription_id`) REFERENCES `b2b_subscriptions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
