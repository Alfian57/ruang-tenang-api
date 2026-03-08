-- Map regions (islands/areas on the progress map)
CREATE TABLE IF NOT EXISTS map_regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_key VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    image VARCHAR(500),
    unlock_type VARCHAR(50) NOT NULL DEFAULT 'level',
    unlock_value INTEGER NOT NULL DEFAULT 1,
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0,
    parent_region_id UUID REFERENCES map_regions(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Map landmarks (points of interest within a region)
CREATE TABLE IF NOT EXISTS map_landmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id UUID NOT NULL REFERENCES map_regions(id) ON DELETE CASCADE,
    landmark_key VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    unlock_type VARCHAR(50) NOT NULL DEFAULT 'activity_count',
    unlock_activity VARCHAR(50),
    unlock_value INTEGER NOT NULL DEFAULT 1,
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    xp_reward INTEGER NOT NULL DEFAULT 0,
    coin_reward INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_map_landmarks_region_id ON map_landmarks(region_id);

-- User map progress
CREATE TABLE IF NOT EXISTS user_map_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    region_id UUID NOT NULL REFERENCES map_regions(id) ON DELETE CASCADE,
    is_unlocked BOOLEAN DEFAULT FALSE,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, region_id)
);

CREATE INDEX idx_user_map_progress_user_id ON user_map_progress(user_id);

-- User landmark progress
CREATE TABLE IF NOT EXISTS user_landmark_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    landmark_id UUID NOT NULL REFERENCES map_landmarks(id) ON DELETE CASCADE,
    is_unlocked BOOLEAN DEFAULT FALSE,
    current_value INTEGER NOT NULL DEFAULT 0,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    reward_claimed BOOLEAN DEFAULT FALSE,
    UNIQUE(user_id, landmark_id)
);

CREATE INDEX idx_user_landmark_progress_user_id ON user_landmark_progress(user_id);
