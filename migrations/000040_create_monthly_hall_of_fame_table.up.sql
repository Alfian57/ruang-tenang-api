-- Monthly hall of fame table
CREATE TABLE monthly_hall_of_fame (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    level INTEGER NOT NULL,
    month INTEGER NOT NULL,
    year INTEGER NOT NULL,
    rank INTEGER NOT NULL,
    monthly_xp INTEGER NOT NULL,
    message VARCHAR(300),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(level, month, year, rank)
);

CREATE INDEX idx_hall_of_fame_level_month ON monthly_hall_of_fame(level, month, year);
CREATE INDEX idx_hall_of_fame_user ON monthly_hall_of_fame(user_id);
