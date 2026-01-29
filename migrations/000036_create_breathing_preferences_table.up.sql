-- Breathing Preferences Table
-- Stores user preferences for breathing exercises
CREATE TABLE breathing_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Default session settings
    default_duration_seconds INT DEFAULT 300, -- 5 minutes
    default_technique_id UUID REFERENCES breathing_techniques(id) ON DELETE SET NULL,
    
    -- Guidance preferences
    voice_guidance VARCHAR(20) DEFAULT 'ask', -- always_on, always_off, ask
    background_sound VARCHAR(20) DEFAULT 'ask', -- always_on, always_off, ask
    default_background_sound VARCHAR(50) DEFAULT 'none', -- ocean, rain, forest, piano, white_noise, none
    haptic_feedback BOOLEAN DEFAULT TRUE,
    
    -- Visual preferences
    animation_speed VARCHAR(20) DEFAULT 'normal', -- slow, normal, fast
    theme VARCHAR(50) DEFAULT 'default', -- default, calm_blue, warm_orange, forest_green
    
    -- Reminder settings
    reminder_enabled BOOLEAN DEFAULT FALSE,
    reminder_time TIME, -- e.g., '08:00:00'
    reminder_days VARCHAR(20) DEFAULT '1234567', -- 1=Mon, 2=Tue, etc.
    
    -- Tutorial status
    tutorial_completed BOOLEAN DEFAULT FALSE,
    
    -- Streak tracking
    current_streak INT DEFAULT 0,
    longest_streak INT DEFAULT 0,
    last_practice_date DATE,
    streak_freeze_available BOOLEAN DEFAULT TRUE, -- Reset weekly
    streak_freeze_used_at DATE,
    
    -- Daily XP tracking (for cap)
    daily_xp_earned INT DEFAULT 0,
    daily_xp_date DATE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast user lookup
CREATE INDEX idx_breathing_preferences_user_id ON breathing_preferences(user_id);
