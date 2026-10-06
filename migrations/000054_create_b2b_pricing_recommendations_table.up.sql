SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `b2b_pricing_recommendations` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `organization_id` BIGINT NOT NULL,
    `generated_for_date` DATE NOT NULL,
    `recommended_plan_id` BIGINT,
    `recommended_billing_cycle` VARCHAR(20) NOT NULL,
    `recommended_seats` BIGINT NOT NULL,
    `estimated_monthly_cost` BIGINT NOT NULL DEFAULT 0,
    `estimated_yearly_saving` BIGINT NOT NULL DEFAULT 0,
    `confidence_score` DECIMAL(5,2) NOT NULL DEFAULT 0,
    `reasons_json` JSON NOT NULL DEFAULT (JSON_ARRAY()),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_b2b_pricing_recommendations_org_date` (`organization_id`, `generated_for_date`),
    CONSTRAINT `fk_b2b_pricing_recommendations_organization_id_fkey` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_b2b_pricing_recommendations_recommended_plan_id_fkey` FOREIGN KEY (`recommended_plan_id`) REFERENCES `b2b_plans` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
