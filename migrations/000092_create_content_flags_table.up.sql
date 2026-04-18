CREATE TABLE IF NOT EXISTS content_flags (
    id SERIAL PRIMARY KEY,
    content_type VARCHAR(50) NOT NULL,
    content_id INT NOT NULL,
    flag_type VARCHAR(50) NOT NULL,
    flag_category VARCHAR(100) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium',
    ai_confidence DECIMAL(5,2),
    ai_reason TEXT,
    flagged_by_id INT REFERENCES users(id) ON DELETE SET NULL,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_by_id INT REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
