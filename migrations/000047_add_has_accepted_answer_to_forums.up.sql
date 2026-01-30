-- Add has_accepted_answer flag to forums table for quick filtering
-- This helps identify threads that have been "solved"

ALTER TABLE forums
ADD COLUMN has_accepted_answer BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_forums_has_accepted ON forums(has_accepted_answer);
