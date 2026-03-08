-- League divisions (10 divisions: Bronze → Diamond)
CREATE TABLE IF NOT EXISTS league_divisions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    icon VARCHAR(50) NOT NULL,
    tier INTEGER NOT NULL UNIQUE,
    color VARCHAR(20) NOT NULL DEFAULT '#888888',
    min_rank INTEGER NOT NULL DEFAULT 0,
    promotion_slots INTEGER NOT NULL DEFAULT 10,
    demotion_slots INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- League seasons (weekly)
CREATE TABLE IF NOT EXISTS league_seasons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    week_number INTEGER NOT NULL,
    year INTEGER NOT NULL,
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    is_processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(week_number, year)
);

CREATE INDEX idx_league_seasons_active ON league_seasons(is_active);

-- League participants
CREATE TABLE IF NOT EXISTS league_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id UUID NOT NULL REFERENCES league_seasons(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    division_id INTEGER NOT NULL REFERENCES league_divisions(id) ON DELETE CASCADE,
    weekly_xp BIGINT NOT NULL DEFAULT 0,
    rank INTEGER NOT NULL DEFAULT 0,
    is_promoted BOOLEAN DEFAULT FALSE,
    is_demoted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(season_id, user_id)
);

CREATE INDEX idx_league_participants_season ON league_participants(season_id);
CREATE INDEX idx_league_participants_user ON league_participants(user_id);
CREATE INDEX idx_league_participants_division ON league_participants(division_id);
CREATE INDEX idx_league_participants_ranking ON league_participants(season_id, division_id, weekly_xp DESC);
