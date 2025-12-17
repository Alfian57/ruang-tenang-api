package services

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"gorm.io/gorm"
)

type GamificationService struct {
	db *gorm.DB
}

func NewGamificationService(db *gorm.DB) *GamificationService {
	return &GamificationService{db: db}
}

// AwardExp adds EXP to a user if the daily limit for the activity hasn't been reached.
func (s *GamificationService) AwardExp(userID uint, activityType gamification.ActivityType, points int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Check daily limit if applicable
		if limit := getDailyLimit(activityType); limit > 0 {
			today := time.Now().Truncate(24 * time.Hour)
			var count int64
			err := tx.Model(&models.UserActivity{}).
				Where("user_id = ? AND activity_type = ? AND date = ?", userID, activityType, today).
				Count(&count).Error
			if err != nil {
				return err
			}

			if int(count) >= limit {
				return nil // Limit reached, no error, just no points
			}

			// 2. Record activity
			activity := models.UserActivity{
				UserID:       userID,
				ActivityType: string(activityType),
				Date:         today,
				Count:        1, // Initial count, though we might just insert rows.
				// Actually, my migration unique constraint is (user_id, activity_type, date).
				// So I should upsert (increment count) or just insert if I want one row per day per activity type tracking count.
			}

			// Let's use Upsert to increment count
			// On conflict (user_id, activity_type, date) do update count = count + 1
			err = tx.Where(models.UserActivity{
				UserID:       userID,
				ActivityType: string(activityType),
				Date:         today,
			}).Assign(models.UserActivity{
				Count: int(count) + 1,
			}).FirstOrCreate(&activity).Error

			if err != nil {
				// If FirstOrCreate fails, it might be race condition or something else, but let's try manual handling if needed.
				// However, GORM FirstOrCreate with Where matches existing.
				// Better approach:
				// If count == 0, create. If count > 0, update.
				if count == 0 {
					activity.Count = 1
					if err := tx.Create(&activity).Error; err != nil {
						return err
					}
				} else {
					if err := tx.Model(&models.UserActivity{}).
						Where("user_id = ? AND activity_type = ? AND date = ?", userID, activityType, today).
						Update("count", gorm.Expr("count + ?", 1)).Error; err != nil {
						return err
					}
				}
			}
		}

		// 3. Add EXP to User
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("exp", gorm.Expr("exp + ?", points)).Error; err != nil {
			return err
		}

		// 4. Record EXP history
		expHistory := models.ExpHistory{
			UserID:       userID,
			ActivityType: string(activityType),
			Points:       int(points),
			Description:  getActivityDescription(activityType),
		}
		if err := tx.Create(&expHistory).Error; err != nil {
			return err
		}

		return nil
	})
}

func getDailyLimit(activityType gamification.ActivityType) int {
	switch activityType {
	case gamification.ActivityChatAI:
		return gamification.LimitChatAI
	case gamification.ActivityForumComment:
		return gamification.LimitForumComment
	default:
		return 0 // No limit
	}
}

func getActivityDescription(activityType gamification.ActivityType) string {
	switch activityType {
	case gamification.ActivityChatAI:
		return "Melakukan chat dengan AI"
	case gamification.ActivityUploadArticle:
		return "Mengunggah artikel baru"
	case gamification.ActivityForumComment:
		return "Berkomentar di forum"
	default:
		return "Aktivitas lainnya"
	}
}
