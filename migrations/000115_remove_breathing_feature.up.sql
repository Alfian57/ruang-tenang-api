-- Remove persisted data and schema owned by the breathing feature.
-- Historical migrations remain unchanged; this migration is the forward cleanup
-- for databases that already contain the feature.

DELETE FROM user_timed_challenges
WHERE template_id IN (
    SELECT id FROM timed_challenge_templates WHERE challenge_type = 'breathing'
);

DELETE FROM timed_challenge_templates
WHERE challenge_type = 'breathing';

DELETE FROM user_feature_usages
WHERE feature_key IN ('basic_breathing', 'breathing_timer_custom', 'all_breathing_techniques');

DELETE FROM user_feature_unlocks
WHERE feature_id IN (
    SELECT id
    FROM feature_definitions
    WHERE feature_key IN ('basic_breathing', 'breathing_timer_custom', 'all_breathing_techniques')
);

DELETE FROM feature_definitions
WHERE feature_key IN ('basic_breathing', 'breathing_timer_custom', 'all_breathing_techniques');

DELETE FROM user_landmark_progress
WHERE landmark_id IN (
    SELECT id FROM map_landmarks WHERE unlock_activity = 'breathing'
);

DELETE FROM map_landmarks
WHERE unlock_activity = 'breathing';

DELETE FROM daily_tasks
WHERE task_type IN ('breathing_exercise', 'premium_breathing_pro');

DELETE FROM user_activities
WHERE activity_type = 'breathing';

DELETE FROM exp_histories
WHERE activity_type IN ('breathing', 'breathing_session');

DELETE FROM friend_quests
WHERE quest_type = 'breathing';

DELETE FROM user_combos
WHERE last_activity_type = 'breathing';

DROP TABLE IF EXISTS breathing_favorites CASCADE;
DROP TABLE IF EXISTS breathing_sessions CASCADE;
DROP TABLE IF EXISTS breathing_preferences CASCADE;
DROP TABLE IF EXISTS breathing_techniques CASCADE;

ALTER TABLE chat_sessions
    DROP COLUMN IF EXISTS enable_breathing_context;
