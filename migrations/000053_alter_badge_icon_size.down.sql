-- Revert badge_icon column size
ALTER TABLE level_configs ALTER COLUMN badge_icon TYPE VARCHAR(50);
