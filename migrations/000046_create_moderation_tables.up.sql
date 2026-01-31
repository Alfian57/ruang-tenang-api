-- Create user_blocks table
CREATE TABLE IF NOT EXISTS user_blocks (
    id SERIAL PRIMARY KEY,
    blocker_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(blocker_id, blocked_id)
);

-- Create user_reports table
CREATE TABLE IF NOT EXISTS user_reports (
    id SERIAL PRIMARY KEY,
    reporter_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    report_type VARCHAR(50) NOT NULL,
    reported_content_id INT,
    reported_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    handled_by_id INT REFERENCES users(id) ON DELETE SET NULL,
    handled_at TIMESTAMP WITH TIME ZONE,
    action_taken VARCHAR(100),
    moderator_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create user_strikes table
CREATE TABLE IF NOT EXISTS user_strikes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    report_id INT REFERENCES user_reports(id) ON DELETE SET NULL,
    reason VARCHAR(255) NOT NULL,
    severity VARCHAR(20) DEFAULT 'warning',
    issued_by_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create content_flags table
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

-- Create moderator_actions table
CREATE TABLE IF NOT EXISTS moderator_actions (
    id SERIAL PRIMARY KEY,
    moderator_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id INT NOT NULL,
    previous_state TEXT,
    new_state TEXT,
    reason TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
