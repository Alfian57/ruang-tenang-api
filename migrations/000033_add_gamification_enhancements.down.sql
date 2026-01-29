DROP TABLE IF EXISTS community_stats;
DROP TABLE IF EXISTS monthly_hall_of_fame;

ALTER TABLE users DROP COLUMN IF EXISTS profile_theme;
ALTER TABLE users DROP COLUMN IF EXISTS profile_banner;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_border_color;
ALTER TABLE users DROP COLUMN IF EXISTS tagline;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
ALTER TABLE users DROP COLUMN IF EXISTS current_streak;
ALTER TABLE users DROP COLUMN IF EXISTS longest_streak;
ALTER TABLE users DROP COLUMN IF EXISTS last_activity_date;
ALTER TABLE users DROP COLUMN IF EXISTS total_activities;

ALTER TABLE level_configs DROP COLUMN IF EXISTS tier_name;
ALTER TABLE level_configs DROP COLUMN IF EXISTS tier_color;
ALTER TABLE level_configs DROP COLUMN IF EXISTS description;
