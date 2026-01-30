package development

import (
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

// SeedUserMoods seeds test user moods for development
func SeedUserMoods(db *gorm.DB) error {
	// Get test users
	var users []models.User
	if err := db.Where("role = ?", models.RoleMember).Limit(5).Find(&users).Error; err != nil {
		return err
	}

	// Use actual MoodType values from the model
	moods := []models.MoodType{
		models.MoodHappy,
		models.MoodNeutral,
		models.MoodAngry,
		models.MoodDisappointed,
		models.MoodSad,
		models.MoodCrying,
	}

	for _, user := range users {
		// Check if user already has moods
		var count int64
		db.Model(&models.UserMood{}).Where("user_id = ?", user.ID).Count(&count)
		if count > 0 {
			continue
		}

		// Create moods for last 14 days
		for i := 0; i < 14; i++ {
			// 70% chance of having a mood entry each day
			if rand.Float32() > 0.7 {
				continue
			}

			mood := models.UserMood{
				UserID:    user.ID,
				Mood:      moods[rand.Intn(len(moods))],
				CreatedAt: time.Now().AddDate(0, 0, -i).Add(time.Duration(rand.Intn(12)+8) * time.Hour), // Random time between 8am-8pm
			}

			db.Create(&mood)
		}
	}

	return nil
}
