SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `crisis_keywords` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `keyword` VARCHAR(255) NOT NULL,
    `category` VARCHAR(100) NOT NULL,
    `severity` VARCHAR(20) NOT NULL DEFAULT 'high',
    `language` VARCHAR(10) NOT NULL DEFAULT 'id',
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    `notes` TEXT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `crisis_keywords_keyword_language_key` (`keyword`, `language`),
    KEY `idx_crisis_keywords_category` (`category`),
    KEY `idx_crisis_keywords_is_active` (`is_active`),
    KEY `idx_crisis_keywords_language` (`language`),
    KEY `idx_crisis_keywords_severity` (`severity`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
