-- Remove moderation, profile, and gamification fields from users table
ALTER TABLE users 
DROP COLUMN IF EXISTS suspension_end,
DROP COLUMN IF EXISTS suspension_reason,
DROP COLUMN IF EXISTS is_banned,
DROP COLUMN IF EXISTS ban_reason,
DROP COLUMN IF EXISTS has_accepted_a_idisclaimer,
DROP COLUMN IF EXISTS content_warning_preference,
DROP COLUMN IF EXISTS profile_theme,
DROP COLUMN IF EXISTS profile_banner,
DROP COLUMN IF EXISTS avatar_border_color,
DROP COLUMN IF EXISTS tagline,
DROP COLUMN IF EXISTS bio,
DROP COLUMN IF EXISTS current_streak,
DROP COLUMN IF EXISTS longest_streak,
DROP COLUMN IF EXISTS last_activity_date,
DROP COLUMN IF EXISTS total_activities,
DROP COLUMN IF EXISTS last_login_date,
DROP COLUMN IF EXISTS login_streak;
