-- Breathing preferences table
CREATE TABLE breathing_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Default session settings
    default_duration_seconds INT DEFAULT 300,
    default_technique_id UUID REFERENCES breathing_techniques(id) ON DELETE SET NULL,
    
    -- Guidance preferences
    voice_guidance VARCHAR(20) DEFAULT 'ask',
    background_sound VARCHAR(20) DEFAULT 'ask',
    default_background_sound VARCHAR(50) DEFAULT 'none',
    haptic_feedback BOOLEAN DEFAULT TRUE,
    
    -- Visual preferences
    animation_speed VARCHAR(20) DEFAULT 'normal',
    theme VARCHAR(50) DEFAULT 'default',
    
    -- Reminder settings
    reminder_enabled BOOLEAN DEFAULT FALSE,
    reminder_time TIME,
    reminder_days VARCHAR(20) DEFAULT '1234567',
    
    -- Tutorial status
    tutorial_completed BOOLEAN DEFAULT FALSE,
    
    -- Streak tracking
    current_streak INT DEFAULT 0,
    longest_streak INT DEFAULT 0,
    last_practice_date DATE,
    streak_freeze_available BOOLEAN DEFAULT TRUE,
    streak_freeze_used_at DATE,
    
    -- Daily XP tracking
    daily_xp_earned INT DEFAULT 0,
    daily_xp_date DATE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_breathing_preferences_user_id ON breathing_preferences(user_id);
