-- Remove tier fields from level_configs table
ALTER TABLE level_configs 
DROP COLUMN IF EXISTS tier_name,
DROP COLUMN IF EXISTS tier_color,
DROP COLUMN IF EXISTS description;
