CREATE TABLE IF NOT EXISTS broadcast_notifications (
    id CHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    icon VARCHAR(500) DEFAULT '',
    url VARCHAR(500) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    scheduled_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    sent_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    created_by INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_broadcast_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_broadcast_status CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'cancelled'))
);

CREATE INDEX idx_broadcast_status ON broadcast_notifications(status);
CREATE INDEX idx_broadcast_scheduled_at ON broadcast_notifications(scheduled_at);
