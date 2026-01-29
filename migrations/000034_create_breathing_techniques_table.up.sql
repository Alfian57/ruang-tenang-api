-- Breathing Techniques Table
-- Stores both predefined and custom user techniques
CREATE TABLE breathing_techniques (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE,
    description TEXT,
    benefits TEXT,
    best_for TEXT,
    
    -- Timing configuration (in seconds)
    inhale_duration INT NOT NULL DEFAULT 4,
    inhale_hold_duration INT NOT NULL DEFAULT 0,
    exhale_duration INT NOT NULL DEFAULT 4,
    exhale_hold_duration INT NOT NULL DEFAULT 0,
    
    -- Visual & UI
    icon VARCHAR(50) DEFAULT '🌬️',
    color VARCHAR(20) DEFAULT '#6366F1',
    animation_type VARCHAR(50) DEFAULT 'circle', -- circle, square, wave, minimal
    
    -- Metadata
    difficulty VARCHAR(20) DEFAULT 'easy', -- easy, intermediate, advanced
    category VARCHAR(50) DEFAULT 'general', -- general, sleep, stress, energy, focus
    origin TEXT, -- e.g., "Developed by Dr. Andrew Weil"
    
    -- System flags
    is_system BOOLEAN DEFAULT FALSE, -- true for predefined techniques
    is_active BOOLEAN DEFAULT TRUE,
    
    -- User ownership (NULL for system techniques)
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookups
CREATE INDEX idx_breathing_techniques_user_id ON breathing_techniques(user_id);
CREATE INDEX idx_breathing_techniques_category ON breathing_techniques(category);
CREATE INDEX idx_breathing_techniques_is_system ON breathing_techniques(is_system);

-- Insert predefined system techniques
INSERT INTO breathing_techniques (name, slug, description, benefits, best_for, inhale_duration, inhale_hold_duration, exhale_duration, exhale_hold_duration, icon, color, animation_type, difficulty, category, origin, is_system) VALUES
(
    'Box Breathing',
    'box-breathing',
    'Teknik pernapasan kotak yang digunakan oleh Navy SEALs untuk tetap tenang dan fokus dalam situasi stres tinggi.',
    'Mengurangi stres, menenangkan sistem saraf, meningkatkan fokus dan konsentrasi, membantu mengendalikan respons fight-or-flight.',
    'Saat merasa cemas sebelum presentasi, menghadapi deadline, atau saat serangan panik ringan.',
    4, 4, 4, 4,
    '⬜', '#3B82F6', 'square', 'easy', 'stress',
    'Digunakan oleh Navy SEALs dan first responders',
    TRUE
),
(
    '4-7-8 Breathing',
    '4-7-8-breathing',
    'Teknik relaksasi yang dikembangkan Dr. Andrew Weil, sangat efektif untuk membantu tidur dan meredakan kecemasan.',
    'Membantu tidur lebih cepat, meredakan kecemasan, memperlambat detak jantung, mengaktifkan sistem saraf parasimpatis.',
    'Sebelum tidur, saat insomnia, pikiran yang terus berputar, atau butuh relaksasi mendalam.',
    4, 7, 8, 0,
    '🌙', '#8B5CF6', 'circle', 'intermediate', 'sleep',
    'Dikembangkan oleh Dr. Andrew Weil berdasarkan pranayama yoga',
    TRUE
),
(
    'Coherent Breathing',
    'coherent-breathing',
    'Pernapasan seimbang dengan ritme yang konsisten untuk mencapai keseimbangan sistem saraf dan ketenangan pikiran.',
    'Menyeimbangkan sistem saraf, meningkatkan heart rate variability (HRV), meningkatkan konsentrasi dan clarity mental.',
    'Meditasi, belajar, mental reset, atau saat butuh ketenangan tanpa mengantuk.',
    5, 0, 5, 0,
    '♾️', '#10B981', 'wave', 'easy', 'focus',
    'Berdasarkan penelitian heart rate variability',
    TRUE
),
(
    'Energizing Breath',
    'energizing-breath',
    'Teknik pernapasan cepat yang meningkatkan energi dan kesadaran tanpa kafein.',
    'Meningkatkan energi dan alertness, melawan kelelahan, meningkatkan sirkulasi darah, membangunkan tubuh dan pikiran.',
    'Pagi hari setelah bangun, afternoon slump, sebelum olahraga, atau saat merasa lesu.',
    2, 0, 4, 0,
    '⚡', '#F59E0B', 'circle', 'easy', 'energy',
    'Adaptasi dari teknik Kapalabhati pranayama',
    TRUE
),
(
    'Deep Calm Breathing',
    'deep-calm-breathing',
    'Pernapasan dalam dengan exhale panjang untuk menenangkan emosi intens dan overwhelm.',
    'Regulasi emosi yang kuat, menenangkan setelah konflik, mengurangi overwhelm, menurunkan tekanan darah.',
    'Setelah pertengkaran, saat merasa overwhelmed, emosi intens, atau butuh grounding cepat.',
    6, 2, 8, 0,
    '🌊', '#06B6D4', 'wave', 'intermediate', 'stress',
    'Berdasarkan teknik pernapasan diafragma klinis',
    TRUE
);
