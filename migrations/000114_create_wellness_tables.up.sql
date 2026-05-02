CREATE TABLE user_wellness_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    initial_mood VARCHAR(50) NOT NULL DEFAULT '',
    goals_json JSONB NOT NULL DEFAULT '[]',
    habits_json JSONB NOT NULL DEFAULT '[]',
    tour_completed_at TIMESTAMPTZ,
    onboarding_completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wellness_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id BIGINT REFERENCES user_wellness_profiles(id) ON DELETE SET NULL,
    title VARCHAR(180) NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    starts_on DATE NOT NULL,
    ends_on DATE NOT NULL,
    generated_from_mood VARCHAR(50) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wellness_plans_status CHECK (status IN ('active', 'completed', 'archived'))
);

CREATE UNIQUE INDEX uq_wellness_plans_user_active
ON wellness_plans(user_id)
WHERE status = 'active';

CREATE INDEX idx_wellness_plans_user_dates ON wellness_plans(user_id, starts_on, ends_on);

CREATE TABLE wellness_plan_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES wellness_plans(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_number INTEGER NOT NULL,
    item_date DATE NOT NULL,
    title VARCHAR(180) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    action_type VARCHAR(50) NOT NULL,
    route VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    completed_at TIMESTAMPTZ,
    metadata_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wellness_plan_items_day CHECK (day_number BETWEEN 1 AND 7),
    CONSTRAINT chk_wellness_plan_items_status CHECK (status IN ('pending', 'completed', 'skipped'))
);

CREATE INDEX idx_wellness_plan_items_plan_day ON wellness_plan_items(plan_id, day_number);
CREATE INDEX idx_wellness_plan_items_user_date ON wellness_plan_items(user_id, item_date);

CREATE TABLE weekly_insight_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    mood_summary_json JSONB NOT NULL DEFAULT '{}',
    activity_summary_json JSONB NOT NULL DEFAULT '{}',
    insight_json JSONB NOT NULL DEFAULT '{}',
    premium_sections_json JSONB NOT NULL DEFAULT '{}',
    narrative TEXT NOT NULL DEFAULT '',
    recommendations_json JSONB NOT NULL DEFAULT '[]',
    is_ai_enhanced BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, week_start)
);

CREATE INDEX idx_weekly_insight_snapshots_user_week
ON weekly_insight_snapshots(user_id, week_start DESC);

CREATE TABLE wellness_need_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    condition VARCHAR(50) NOT NULL,
    recommendations_json JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wellness_need_events_user_created
ON wellness_need_events(user_id, created_at DESC);
