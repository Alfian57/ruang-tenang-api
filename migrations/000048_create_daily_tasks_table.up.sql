-- Daily Tasks Table
-- Tracks daily tasks and their completion status for each user
CREATE TABLE IF NOT EXISTS daily_tasks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    task_type VARCHAR(50) NOT NULL, -- 'daily_login', 'record_mood', 'chat_ai', 'read_article', 'listen_songs', 'write_journal', 'comment_forum'
    task_date DATE NOT NULL,
    target_count INTEGER NOT NULL DEFAULT 1, -- e.g., 3 for chat messages, 3 for songs
    current_count INTEGER NOT NULL DEFAULT 0,
    is_completed BOOLEAN DEFAULT FALSE,
    is_claimed BOOLEAN DEFAULT FALSE,
    xp_reward INTEGER NOT NULL DEFAULT 0,
    completed_at TIMESTAMP,
    claimed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_daily_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_daily_task_user_type_date UNIQUE (user_id, task_type, task_date)
);

CREATE INDEX idx_daily_tasks_user_id ON daily_tasks(user_id);
CREATE INDEX idx_daily_tasks_task_date ON daily_tasks(task_date);
CREATE INDEX idx_daily_tasks_user_date ON daily_tasks(user_id, task_date);
CREATE INDEX idx_daily_tasks_completed ON daily_tasks(is_completed);
CREATE INDEX idx_daily_tasks_claimed ON daily_tasks(is_claimed);

-- Add last_login_date to users table for tracking daily login
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS last_login_date DATE,
ADD COLUMN IF NOT EXISTS login_streak INTEGER DEFAULT 0;
