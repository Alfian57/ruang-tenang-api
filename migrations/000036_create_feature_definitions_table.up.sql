-- Feature definitions table
CREATE TABLE feature_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_key VARCHAR(100) NOT NULL UNIQUE,
    feature_name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    required_level INTEGER NOT NULL DEFAULT 1,
    category VARCHAR(50) DEFAULT 'general',
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_feature_definitions_level ON feature_definitions(required_level);
CREATE INDEX idx_feature_definitions_category ON feature_definitions(category);
