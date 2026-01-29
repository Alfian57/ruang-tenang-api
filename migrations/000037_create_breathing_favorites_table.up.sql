-- Breathing Favorites Table
-- Stores user's favorite techniques for quick access
CREATE TABLE breathing_favorites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    technique_id UUID NOT NULL REFERENCES breathing_techniques(id) ON DELETE CASCADE,
    sort_order INT DEFAULT 0, -- For ordering favorites
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, technique_id)
);

-- Index for fast lookups
CREATE INDEX idx_breathing_favorites_user_id ON breathing_favorites(user_id);
