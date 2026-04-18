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
