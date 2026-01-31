-- Breathing sessions table
CREATE TABLE breathing_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    technique_id UUID NOT NULL REFERENCES breathing_techniques(id) ON DELETE CASCADE,
    
    -- Session details
    duration_seconds INT NOT NULL,
    target_duration_seconds INT NOT NULL,
    cycles_completed INT NOT NULL DEFAULT 0,
    
    -- Session settings used
    voice_guidance_enabled BOOLEAN DEFAULT FALSE,
    background_sound VARCHAR(50),
    haptic_feedback_enabled BOOLEAN DEFAULT FALSE,
    
    -- Completion status
    completed BOOLEAN DEFAULT FALSE,
    completed_percentage INT DEFAULT 0,
    
    -- Timestamps
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    
    -- XP awarded
    xp_earned INT DEFAULT 0,
    
    -- Optional: mood before/after
    mood_before VARCHAR(20),
    mood_after VARCHAR(20),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_breathing_sessions_user_id ON breathing_sessions(user_id);
CREATE INDEX idx_breathing_sessions_technique_id ON breathing_sessions(technique_id);
CREATE INDEX idx_breathing_sessions_started_at ON breathing_sessions(started_at);
-- Composite user_date index not needed - use user_id + started_at indexes instead
