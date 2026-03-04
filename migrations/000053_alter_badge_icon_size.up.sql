-- Increase badge_icon column size to accommodate image URL paths
ALTER TABLE level_configs ALTER COLUMN badge_icon TYPE VARCHAR(255);
