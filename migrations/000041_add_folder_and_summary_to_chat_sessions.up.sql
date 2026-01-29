-- Add folder_id to chat_sessions for folder organization
ALTER TABLE chat_sessions 
ADD COLUMN folder_id INTEGER REFERENCES chat_folders(id) ON DELETE SET NULL;

-- Add summary field for AI-generated session summaries
ALTER TABLE chat_sessions 
ADD COLUMN summary TEXT;

-- Add summary_generated_at to track when summary was last generated
ALTER TABLE chat_sessions 
ADD COLUMN summary_generated_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_chat_sessions_folder ON chat_sessions(folder_id);
