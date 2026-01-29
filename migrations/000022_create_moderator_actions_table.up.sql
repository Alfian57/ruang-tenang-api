-- Moderator actions audit log
CREATE TABLE moderator_actions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    moderator_id BIGINT UNSIGNED NOT NULL,
    action_type VARCHAR(100) NOT NULL, -- 'article_approved', 'article_rejected', 'content_removed', 'user_warned', 'user_suspended', 'user_banned', 'report_dismissed', etc.
    target_type VARCHAR(50) NOT NULL, -- 'article', 'forum', 'forum_post', 'user', 'report'
    target_id BIGINT UNSIGNED NOT NULL,
    previous_state TEXT, -- JSON of previous state
    new_state TEXT, -- JSON of new state
    reason TEXT,
    notes TEXT, -- internal notes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_moderator (moderator_id),
    INDEX idx_action_type (action_type),
    INDEX idx_target (target_type, target_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE CASCADE
);
