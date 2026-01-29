-- Update role enum to include moderator
ALTER TABLE users MODIFY COLUMN role VARCHAR(20) DEFAULT 'member';

-- Add user suspension/ban fields
ALTER TABLE users
ADD COLUMN suspension_end TIMESTAMP NULL AFTER is_blocked,
ADD COLUMN suspension_reason TEXT AFTER suspension_end,
ADD COLUMN is_banned BOOLEAN NOT NULL DEFAULT FALSE AFTER suspension_reason,
ADD COLUMN ban_reason TEXT AFTER is_banned,
ADD COLUMN has_accepted_ai_disclaimer BOOLEAN NOT NULL DEFAULT FALSE AFTER ban_reason,
ADD COLUMN content_warning_preference VARCHAR(20) NOT NULL DEFAULT 'show' AFTER has_accepted_ai_disclaimer;

-- Note: 'moderator' role will be stored as varchar, we'll handle it in code
-- content_warning_preference: 'show', 'hide_all', 'ask_each_time'

ALTER TABLE users ADD INDEX idx_is_banned (is_banned);
ALTER TABLE users ADD INDEX idx_suspension_end (suspension_end);
