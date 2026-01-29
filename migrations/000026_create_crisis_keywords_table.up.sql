-- Crisis keywords for detection system
CREATE TABLE crisis_keywords (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL, -- 'self_harm', 'suicide', 'severe_depression', 'emergency'
    severity VARCHAR(20) NOT NULL DEFAULT 'high', -- 'medium', 'high', 'critical'
    language VARCHAR(10) NOT NULL DEFAULT 'id', -- 'id' for Indonesian, 'en' for English
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE INDEX idx_keyword_lang (keyword, language),
    INDEX idx_category (category),
    INDEX idx_is_active (is_active)
);

-- Insert initial Indonesian crisis keywords
INSERT INTO crisis_keywords (keyword, category, severity, language, is_active) VALUES
-- Self-harm keywords
('menyakiti diri sendiri', 'self_harm', 'critical', 'id', TRUE),
('melukai diri', 'self_harm', 'critical', 'id', TRUE),
('menyayat', 'self_harm', 'critical', 'id', TRUE),
('self harm', 'self_harm', 'critical', 'id', TRUE),
('cutting', 'self_harm', 'high', 'id', TRUE),
('menyilet', 'self_harm', 'critical', 'id', TRUE),

-- Suicide keywords
('bunuh diri', 'suicide', 'critical', 'id', TRUE),
('ingin mati', 'suicide', 'critical', 'id', TRUE),
('mau mati', 'suicide', 'critical', 'id', TRUE),
('mengakhiri hidup', 'suicide', 'critical', 'id', TRUE),
('mengakhiri semuanya', 'suicide', 'high', 'id', TRUE),
('tidak ingin hidup', 'suicide', 'critical', 'id', TRUE),
('nggak mau hidup', 'suicide', 'critical', 'id', TRUE),
('gak mau hidup lagi', 'suicide', 'critical', 'id', TRUE),
('lebih baik mati', 'suicide', 'critical', 'id', TRUE),
('suicide', 'suicide', 'critical', 'id', TRUE),
('gantung diri', 'suicide', 'critical', 'id', TRUE),
('overdose', 'suicide', 'critical', 'id', TRUE),
('minum obat banyak', 'suicide', 'high', 'id', TRUE),

-- Severe depression indicators
('tidak ada harapan', 'severe_depression', 'high', 'id', TRUE),
('nggak ada harapan', 'severe_depression', 'high', 'id', TRUE),
('hidup tidak berarti', 'severe_depression', 'high', 'id', TRUE),
('hidup nggak ada artinya', 'severe_depression', 'high', 'id', TRUE),
('beban bagi semua orang', 'severe_depression', 'high', 'id', TRUE),
('semua akan lebih baik tanpa aku', 'severe_depression', 'critical', 'id', TRUE),
('tidak ada yang peduli', 'severe_depression', 'medium', 'id', TRUE),

-- Emergency indicators
('darurat', 'emergency', 'high', 'id', TRUE),
('tolong', 'emergency', 'medium', 'id', TRUE),
('butuh bantuan sekarang', 'emergency', 'high', 'id', TRUE),
('tidak bisa menahan', 'emergency', 'high', 'id', TRUE);
