package presentation

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedChatContextPreferences applies varied chat context settings for demo/dev users.
func SeedChatContextPreferences(db *gorm.DB) error {
	targetEmails := []string{
		"gading@gmail.com",
		"dery@gmail.com",
		"andhika@gmail.com",
	}

	type contextPreset struct {
		intent            model.ChatSessionIntent
		enableMood        bool
		enableJournal     bool
		enableDailyTask   bool
		enableXPLevel     bool
		enableBreathing   bool
		enablePlaylist    bool
		enableRewards     bool
		enableProgressMap bool
		enableSocial      bool
	}

	presets := []contextPreset{
		{
			intent:            model.ChatSessionIntentReflection,
			enableMood:        true,
			enableJournal:     true,
			enableDailyTask:   true,
			enableXPLevel:     true,
			enableBreathing:   true,
			enablePlaylist:    false,
			enableRewards:     false,
			enableProgressMap: false,
			enableSocial:      false,
		},
		{
			intent:            model.ChatSessionIntentPlanning,
			enableMood:        true,
			enableJournal:     false,
			enableDailyTask:   true,
			enableXPLevel:     true,
			enableBreathing:   false,
			enablePlaylist:    false,
			enableRewards:     true,
			enableProgressMap: true,
			enableSocial:      false,
		},
		{
			intent:            model.ChatSessionIntentGrounding,
			enableMood:        true,
			enableJournal:     false,
			enableDailyTask:   false,
			enableXPLevel:     false,
			enableBreathing:   true,
			enablePlaylist:    true,
			enableRewards:     false,
			enableProgressMap: false,
			enableSocial:      false,
		},
		{
			intent:            model.ChatSessionIntentCoping,
			enableMood:        true,
			enableJournal:     true,
			enableDailyTask:   true,
			enableXPLevel:     false,
			enableBreathing:   true,
			enablePlaylist:    true,
			enableRewards:     true,
			enableProgressMap: false,
			enableSocial:      true,
		},
	}

	for _, email := range targetEmails {
		var user model.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			continue
		}

		var sessions []model.ChatSession
		if err := db.Where("user_id = ?", user.ID).Order("created_at ASC").Find(&sessions).Error; err != nil {
			return err
		}

		if len(sessions) == 0 {
			session := model.ChatSession{
				UserID:                   user.ID,
				Title:                    "Sesi refleksi otomatis",
				EnableMoodContext:        true,
				EnableJournalContext:     true,
				EnableDailyTaskContext:   true,
				EnableXPLevelContext:     true,
				EnableBreathingContext:   true,
				EnablePlaylistContext:    false,
				EnableRewardsContext:     false,
				EnableProgressMapContext: false,
				EnableSocialContext:      false,
				SessionIntent:            model.ChatSessionIntentReflection,
			}
			if err := db.Create(&session).Error; err != nil {
				return err
			}
			sessions = append(sessions, session)
		}

		now := time.Now().UTC()
		for idx, session := range sessions {
			preset := presets[idx%len(presets)]
			if err := db.Model(&model.ChatSession{}).Where("id = ?", session.ID).Updates(map[string]interface{}{
				"enable_mood_context":         preset.enableMood,
				"enable_journal_context":      preset.enableJournal,
				"enable_daily_task_context":   preset.enableDailyTask,
				"enable_xp_level_context":     preset.enableXPLevel,
				"enable_breathing_context":    preset.enableBreathing,
				"enable_playlist_context":     preset.enablePlaylist,
				"enable_rewards_context":      preset.enableRewards,
				"enable_progress_map_context": preset.enableProgressMap,
				"enable_social_context":       preset.enableSocial,
				"session_intent":              preset.intent,
				"updated_at":                  now,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
