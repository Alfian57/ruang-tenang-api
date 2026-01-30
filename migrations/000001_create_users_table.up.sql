-- Users table with all fields consolidated
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member', -- 'admin', 'moderator', 'member'
    exp BIGINT NOT NULL DEFAULT 0,
    avatar VARCHAR(255) DEFAULT '',
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Reset token fields
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    
    -- Moderation fields
    suspension_end TIMESTAMP NULL,
    suspension_reason TEXT,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    ban_reason TEXT,
    has_accepted_ai_disclaimer BOOLEAN NOT NULL DEFAULT FALSE,
    content_warning_preference VARCHAR(20) NOT NULL DEFAULT 'show', -- 'show', 'hide_all', 'ask_each_time'
    
    -- Profile customization
    profile_theme VARCHAR(50) DEFAULT 'default',
    profile_banner VARCHAR(500),
    avatar_border_color VARCHAR(20),
    tagline VARCHAR(200),
    bio TEXT,
    
    -- Gamification stats
    current_streak INTEGER DEFAULT 0,
    longest_streak INTEGER DEFAULT 0,
    last_activity_date DATE,
    total_activities INTEGER DEFAULT 0,
    last_login_date DATE,
    login_streak INTEGER DEFAULT 0,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_blocked ON users(is_blocked);
CREATE INDEX idx_users_is_banned ON users(is_banned);
CREATE INDEX idx_users_suspension_end ON users(suspension_end);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
