SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `story_category_relations` (
    `story_id` VARCHAR(36) NOT NULL,
    `category_id` VARCHAR(36) NOT NULL,
    PRIMARY KEY (`story_id`, `category_id`),
    KEY `idx_story_category_relations_category` (`category_id`),
    KEY `idx_story_category_relations_story` (`story_id`),
    CONSTRAINT `fk_story_category_relations_category_id_fkey` FOREIGN KEY (`category_id`) REFERENCES `story_categories` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_story_category_relations_story_id_fkey` FOREIGN KEY (`story_id`) REFERENCES `inspiring_stories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
