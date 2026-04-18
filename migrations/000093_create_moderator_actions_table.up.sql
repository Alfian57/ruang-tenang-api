CREATE TABLE IF NOT EXISTS moderator_actions (
    id SERIAL PRIMARY KEY,
    moderator_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id INT NOT NULL,
    previous_state TEXT,
    new_state TEXT,
    reason TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
