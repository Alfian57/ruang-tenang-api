package presentation

import (
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedUserMoods seeds test user moods for development
func SeedUserMoods(db *gorm.DB) error {
	startOfToday, endOfToday := jakartaTodayBounds()

	var presentationUsers []model.User
	if err := db.Where("email IN ?", presentationAccountEmails).Find(&presentationUsers).Error; err != nil {
		return err
	}

	presentationUserIDs := make([]uint, 0, len(presentationUsers))
	for _, user := range presentationUsers {
		presentationUserIDs = append(presentationUserIDs, user.ID)
	}
	if len(presentationUserIDs) > 0 {
		if err := db.Where("user_id IN ? AND created_at >= ? AND created_at < ?", presentationUserIDs, startOfToday, endOfToday).
			Delete(&model.UserMood{}).Error; err != nil {
			return err
		}
	}

	// Get all seeded role-user accounts.
	var users []model.User
	if err := db.Where("role = ? AND email IN ?", model.RoleUser, presentationAccountEmails).Order("id ASC").Find(&users).Error; err != nil {
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
		// Check if user already has historical moods
		var count int64
		if err := db.Model(&model.UserMood{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		// Create moods for the last 14 completed days, excluding today.
		for i := 1; i <= 14; i++ {
			// 70% chance of having a mood entry each day
			if rand.Float32() > 0.7 {
				continue
			}

			baseDate := startOfToday.AddDate(0, 0, -i)
			randomHour := time.Duration(rand.Intn(16)+6) * time.Hour // 6am-10pm WIB
			moodTime := baseDate.Add(randomHour)

			mood := model.UserMood{
				UserID:    user.ID,
				Mood:      moods[rand.Intn(len(moods))],
				CreatedAt: moodTime,
			}

			if err := db.Create(&mood).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func jakartaTodayBounds() (time.Time, time.Time) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	now := time.Now().In(loc)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return startOfToday, startOfToday.Add(24 * time.Hour)
}
