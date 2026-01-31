-- Add missing tier fields to level_configs table
ALTER TABLE level_configs 
ADD COLUMN IF NOT EXISTS tier_name VARCHAR(50),
ADD COLUMN IF NOT EXISTS tier_color VARCHAR(20),
ADD COLUMN IF NOT EXISTS description TEXT;
