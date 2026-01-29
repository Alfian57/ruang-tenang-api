-- Content flags for AI moderation and trigger warnings
CREATE TABLE content_flags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    content_type VARCHAR(50) NOT NULL, -- 'article', 'forum', 'forum_post'
    content_id BIGINT UNSIGNED NOT NULL,
    flag_type VARCHAR(50) NOT NULL, -- 'ai_moderation', 'trigger_warning', 'manual'
    flag_category VARCHAR(100) NOT NULL, -- 'misinformation', 'harmful_advice', 'self_harm', 'suicide', 'abuse', etc.
    severity VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'low', 'medium', 'high'
    ai_confidence DECIMAL(5,2), -- AI confidence score (0-100)
    ai_reason TEXT, -- AI explanation for flag
    flagged_by_id BIGINT UNSIGNED, -- NULL if AI, user_id if manual
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_by_id BIGINT UNSIGNED,
    resolved_at TIMESTAMP NULL,
    resolution_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_content (content_type, content_id),
    INDEX idx_flag_type (flag_type),
    INDEX idx_severity (severity),
    INDEX idx_is_resolved (is_resolved),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (flagged_by_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (resolved_by_id) REFERENCES users(id) ON DELETE SET NULL
);
