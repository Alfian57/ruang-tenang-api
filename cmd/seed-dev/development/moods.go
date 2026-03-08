package development

import (
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedUserMoods seeds test user moods for development
func SeedUserMoods(db *gorm.DB) error {
	// Get test users
	var users []model.User
	if err := db.Where("role = ?", model.RoleMember).Limit(5).Find(&users).Error; err != nil {
		return err
	}

	// Use actual MoodType values from the model
	moods := []model.MoodType{
		model.MoodHappy,
		model.MoodNeutral,
		model.MoodAngry,
		model.MoodDisappointed,
		model.MoodSad,
		model.MoodCrying,
	}

	for _, user := range users {
		// Check if user already has moods
		var count int64
		db.Model(&model.UserMood{}).Where("user_id = ?", user.ID).Count(&count)
		if count > 0 {
			continue
		}

		// Create moods for last 14 days
		for i := 0; i < 14; i++ {
			// 70% chance of having a mood entry each day
			if rand.Float32() > 0.7 {
				continue
			}

			// Use UTC time and add random hours within the day (0-23)
			// This ensures the timestamp stays within the correct calendar day
			baseDate := time.Now().UTC().AddDate(0, 0, -i).Truncate(24 * time.Hour)
			randomHour := time.Duration(rand.Intn(16)+6) * time.Hour // 6am-10pm UTC
			moodTime := baseDate.Add(randomHour)

			mood := model.UserMood{
				UserID:    user.ID,
				Mood:      moods[rand.Intn(len(moods))],
				CreatedAt: moodTime,
			}

			db.Create(&mood)
		}
	}

	return nil
}
