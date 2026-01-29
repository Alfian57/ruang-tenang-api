-- User strikes for tracking violations
CREATE TABLE user_strikes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    report_id BIGINT UNSIGNED, -- related report if any
    reason VARCHAR(255) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning', -- 'warning', 'minor', 'major'
    issued_by_id BIGINT UNSIGNED NOT NULL, -- moderator who issued
    expires_at TIMESTAMP NULL, -- NULL means permanent
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user (user_id),
    INDEX idx_is_active (is_active),
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (report_id) REFERENCES user_reports(id) ON DELETE SET NULL,
    FOREIGN KEY (issued_by_id) REFERENCES users(id) ON DELETE CASCADE
);
