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
