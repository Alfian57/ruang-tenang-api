CREATE TABLE level_configs (
    id SERIAL PRIMARY KEY,
    level INTEGER NOT NULL UNIQUE,
    min_exp INTEGER NOT NULL DEFAULT 0,
    badge_name VARCHAR(100) NOT NULL,
    badge_icon VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_level_configs_min_exp ON level_configs(min_exp);
