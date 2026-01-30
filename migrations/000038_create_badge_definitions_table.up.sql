-- Badge definitions table
CREATE TABLE badge_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    badge_key VARCHAR(100) NOT NULL UNIQUE,
    badge_name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    category VARCHAR(50) DEFAULT 'general',
    requirement_type VARCHAR(50) NOT NULL,
    requirement_value INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_badge_definitions_category ON badge_definitions(category);
CREATE INDEX idx_badge_definitions_requirement_type ON badge_definitions(requirement_type);
