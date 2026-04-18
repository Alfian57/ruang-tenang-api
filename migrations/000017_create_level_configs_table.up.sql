-- Level configs table with tier fields consolidated
CREATE TABLE level_configs (
    id SERIAL PRIMARY KEY,
    level INTEGER NOT NULL UNIQUE,
    min_exp INTEGER NOT NULL DEFAULT 0,
    badge_name VARCHAR(100) NOT NULL,
    badge_icon VARCHAR(255) NOT NULL,
    tier_name VARCHAR(50),
    tier_color VARCHAR(20),
    description TEXT,
    task_description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_level_configs_level ON level_configs(level);
CREATE INDEX idx_level_configs_min_exp ON level_configs(min_exp);
