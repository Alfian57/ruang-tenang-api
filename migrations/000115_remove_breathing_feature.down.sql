-- The feature is intentionally not restored with data on rollback. This down
-- migration restores the schema shape only; deleted feature data is not
-- recoverable from the database after the up migration has run.

ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS enable_breathing_context BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE IF NOT EXISTS breathing_techniques (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE,
    description TEXT,
    benefits TEXT,
    best_for TEXT,
    inhale_duration INT NOT NULL DEFAULT 4,
    inhale_hold_duration INT NOT NULL DEFAULT 0,
    exhale_duration INT NOT NULL DEFAULT 4,
    exhale_hold_duration INT NOT NULL DEFAULT 0,
    icon VARCHAR(50) DEFAULT '🌬️',
    color VARCHAR(20) DEFAULT '#6366F1',
    animation_type VARCHAR(50) DEFAULT 'circle',
    difficulty VARCHAR(20) DEFAULT 'easy',
    category VARCHAR(50) DEFAULT 'general',
    origin TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_breathing_techniques_user_id ON breathing_techniques(user_id);
CREATE INDEX IF NOT EXISTS idx_breathing_techniques_category ON breathing_techniques(category);
CREATE INDEX IF NOT EXISTS idx_breathing_techniques_is_system ON breathing_techniques(is_system);
CREATE INDEX IF NOT EXISTS idx_breathing_techniques_slug ON breathing_techniques(slug);

CREATE TABLE IF NOT EXISTS breathing_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    technique_id UUID NOT NULL REFERENCES breathing_techniques(id) ON DELETE CASCADE,
    duration_seconds INT NOT NULL,
    target_duration_seconds INT NOT NULL,
    cycles_completed INT NOT NULL DEFAULT 0,
    voice_guidance_enabled BOOLEAN DEFAULT FALSE,
    background_sound VARCHAR(50),
    haptic_feedback_enabled BOOLEAN DEFAULT FALSE,
    completed BOOLEAN DEFAULT FALSE,
    completed_percentage INT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    xp_earned INT DEFAULT 0,
    mood_before VARCHAR(20),
    mood_after VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_breathing_sessions_user_id ON breathing_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_breathing_sessions_technique_id ON breathing_sessions(technique_id);
CREATE INDEX IF NOT EXISTS idx_breathing_sessions_started_at ON breathing_sessions(started_at);

CREATE TABLE IF NOT EXISTS breathing_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    default_duration_seconds INT DEFAULT 300,
    default_technique_id UUID REFERENCES breathing_techniques(id) ON DELETE SET NULL,
    voice_guidance VARCHAR(20) DEFAULT 'ask',
    background_sound VARCHAR(20) DEFAULT 'ask',
    default_background_sound VARCHAR(50) DEFAULT 'none',
    haptic_feedback BOOLEAN DEFAULT TRUE,
    animation_speed VARCHAR(20) DEFAULT 'normal',
    theme VARCHAR(50) DEFAULT 'default',
    reminder_enabled BOOLEAN DEFAULT FALSE,
    reminder_time TIME,
    reminder_days VARCHAR(20) DEFAULT '1234567',
    tutorial_completed BOOLEAN DEFAULT FALSE,
    current_streak INT DEFAULT 0,
    longest_streak INT DEFAULT 0,
    last_practice_date DATE,
    streak_freeze_available BOOLEAN DEFAULT TRUE,
    streak_freeze_used_at DATE,
    daily_xp_earned INT DEFAULT 0,
    daily_xp_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_breathing_preferences_user_id ON breathing_preferences(user_id);

CREATE TABLE IF NOT EXISTS breathing_favorites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    technique_id UUID NOT NULL REFERENCES breathing_techniques(id) ON DELETE CASCADE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, technique_id)
);

CREATE INDEX IF NOT EXISTS idx_breathing_favorites_user_id ON breathing_favorites(user_id);
