ALTER TABLE users ADD COLUMN streak_freeze_available BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN streak_freeze_used_at DATE;
