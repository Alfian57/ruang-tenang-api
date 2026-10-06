SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `story_tags` (
    `id` VARCHAR(36) NOT NULL DEFAULT (UUID()),
    `story_id` VARCHAR(36) NOT NULL,
    `tag` VARCHAR(50) NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_story_tags_story` (`story_id`),
    KEY `idx_story_tags_tag` (`tag`),
    CONSTRAINT `fk_story_tags_story_id_fkey` FOREIGN KEY (`story_id`) REFERENCES `inspiring_stories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
