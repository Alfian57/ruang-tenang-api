-- Forum post reports table
CREATE TABLE forum_post_reports (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    reporter_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    moderator_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_report_reason CHECK (reason IN ('spam', 'harassment', 'misinformation', 'self_harm', 'off_topic', 'rude', 'other')),
    CONSTRAINT chk_report_status CHECK (status IN ('pending', 'reviewed', 'dismissed', 'actioned'))
);

CREATE UNIQUE INDEX idx_forum_post_reports_unique ON forum_post_reports(post_id, reporter_id) WHERE status = 'pending';
CREATE INDEX idx_forum_post_reports_post_id ON forum_post_reports(post_id);
CREATE INDEX idx_forum_post_reports_reporter_id ON forum_post_reports(reporter_id);
CREATE INDEX idx_forum_post_reports_status ON forum_post_reports(status);
CREATE INDEX idx_forum_post_reports_reason ON forum_post_reports(reason);
