-- Crisis keywords table
CREATE TABLE crisis_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL, -- 'self_harm', 'suicide', 'severe_depression', 'emergency'
    severity VARCHAR(20) NOT NULL DEFAULT 'high', -- 'medium', 'high', 'critical'
    language VARCHAR(10) NOT NULL DEFAULT 'id', -- 'id' for Indonesian, 'en' for English
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(keyword, language)
);

CREATE INDEX idx_crisis_keywords_category ON crisis_keywords(category);
CREATE INDEX idx_crisis_keywords_severity ON crisis_keywords(severity);
CREATE INDEX idx_crisis_keywords_is_active ON crisis_keywords(is_active);
CREATE INDEX idx_crisis_keywords_language ON crisis_keywords(language);
