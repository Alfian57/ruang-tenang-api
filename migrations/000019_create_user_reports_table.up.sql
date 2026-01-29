-- User reports for content and users
CREATE TABLE user_reports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reporter_id BIGINT UNSIGNED NOT NULL,
    report_type VARCHAR(50) NOT NULL, -- 'article', 'forum', 'forum_post', 'user'
    reported_content_id BIGINT UNSIGNED, -- content_id or NULL if user report
    reported_user_id BIGINT UNSIGNED, -- reported user id
    reason VARCHAR(100) NOT NULL, -- 'misinformation', 'harmful', 'harassment', 'spam', 'impersonation', 'other'
    description TEXT, -- optional description from reporter
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'reviewing', 'resolved', 'dismissed'
    handled_by_id BIGINT UNSIGNED, -- moderator who handled
    handled_at TIMESTAMP NULL,
    action_taken VARCHAR(100), -- 'none', 'content_removed', 'user_warned', 'user_suspended', 'user_banned'
    moderator_notes TEXT, -- internal notes for moderators
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_reporter (reporter_id),
    INDEX idx_report_type (report_type),
    INDEX idx_reported_user (reported_user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reported_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (handled_by_id) REFERENCES users(id) ON DELETE SET NULL
);
