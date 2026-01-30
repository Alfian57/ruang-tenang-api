-- Breathing techniques table
CREATE TABLE breathing_techniques (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE,
    description TEXT,
    benefits TEXT,
    best_for TEXT,
    
    -- Timing configuration (in seconds)
    inhale_duration INT NOT NULL DEFAULT 4,
    inhale_hold_duration INT NOT NULL DEFAULT 0,
    exhale_duration INT NOT NULL DEFAULT 4,
    exhale_hold_duration INT NOT NULL DEFAULT 0,
    
    -- Visual & UI
    icon VARCHAR(50) DEFAULT '🌬️',
    color VARCHAR(20) DEFAULT '#6366F1',
    animation_type VARCHAR(50) DEFAULT 'circle',
    
    -- Metadata
    difficulty VARCHAR(20) DEFAULT 'easy',
    category VARCHAR(50) DEFAULT 'general',
    origin TEXT,
    
    -- System flags
    is_system BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- User ownership (NULL for system techniques)
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_breathing_techniques_user_id ON breathing_techniques(user_id);
CREATE INDEX idx_breathing_techniques_category ON breathing_techniques(category);
CREATE INDEX idx_breathing_techniques_is_system ON breathing_techniques(is_system);
CREATE INDEX idx_breathing_techniques_slug ON breathing_techniques(slug);
