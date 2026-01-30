-- Forum Post Reports Table
-- Allows users to report inappropriate posts for moderation
CREATE TABLE IF NOT EXISTS forum_post_reports (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL,
    reporter_id INTEGER NOT NULL,
    reason VARCHAR(50) NOT NULL, -- 'spam', 'harassment', 'misinformation', 'self_harm', 'off_topic', 'other'
    description TEXT, -- Optional additional description
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'reviewed', 'dismissed', 'actioned'
    reviewed_by INTEGER, -- Moderator who reviewed
    reviewed_at TIMESTAMP,
    moderator_notes TEXT, -- Notes from moderator
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_post_reports_post FOREIGN KEY (post_id) REFERENCES forum_posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_reports_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_report_reason CHECK (reason IN ('spam', 'harassment', 'misinformation', 'self_harm', 'off_topic', 'rude', 'other')),
    CONSTRAINT chk_report_status CHECK (status IN ('pending', 'reviewed', 'dismissed', 'actioned'))
);

-- Prevent duplicate reports from same user for same post
CREATE UNIQUE INDEX idx_forum_post_reports_unique ON forum_post_reports(post_id, reporter_id) WHERE status = 'pending';
CREATE INDEX idx_forum_post_reports_post_id ON forum_post_reports(post_id);
CREATE INDEX idx_forum_post_reports_reporter_id ON forum_post_reports(reporter_id);
CREATE INDEX idx_forum_post_reports_status ON forum_post_reports(status);
CREATE INDEX idx_forum_post_reports_reason ON forum_post_reports(reason);
