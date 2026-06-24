package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/xpboost"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	gamificationpkg "github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GamificationService struct {
	db *gorm.DB
}

func NewGamificationService(db *gorm.DB) *GamificationService {
	return &GamificationService{db: db}
}

// AwardExp adds EXP to a user if the daily limit for the activity hasn't been reached.
func (s *GamificationService) AwardExp(ctx context.Context, userID uint, activityType gamificationpkg.ActivityType, points int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Atomic daily-limit check + increment.
		//
		// The previous implementation read Count then wrote separately (read-modify-write),
		// which let concurrent requests all pass the limit check and overshoot. It also used
		// time.Now().Truncate(24h), which snaps to a UTC midnight — not the app timezone
		// (Asia/Jakarta), so the daily window reset at the wrong local time.
		//
		// We now upsert atomically: ON CONFLICT increments count only when below the limit,
		// and RETURNING gives us the post-increment count in the same statement so the
		// check-and-increment is race-free under the (user_id, activity_type, date) unique index.
		if limit := getDailyLimit(activityType); limit > 0 {
			today := timeutil.Today()

			activity := model.UserActivity{
				UserID:       userID,
				ActivityType: string(activityType),
				Date:         today,
				Count:        1,
			}

			// Conditional increment prevents overshoot past the limit even on conflict.
			result := tx.Clauses(
				clause.OnConflict{
					Columns: []clause.Column{
						{Name: "user_id"},
						{Name: "activity_type"},
						{Name: "date"},
					},
					DoUpdates: clause.Assignments(map[string]interface{}{
						"count": gorm.Expr("CASE WHEN user_activities.count < ? THEN user_activities.count + 1 ELSE user_activities.count END", limit),
					}),
				},
				// Defensive explicit RETURNING to ensure activity.Count is populated
				// reliably across driver versions (Postgres supports this).
				clause.Returning{Columns: []clause.Column{{Name: "count"}}},
			).Create(&activity)

			if result.Error != nil {
				return result.Error
			}

			// activity.Count reflects the DB state after upsert (RETURNING populates it
			// for the inserted/updated row). If the limit was already reached, the count
			// stayed unchanged, so we award no EXP.
			if activity.Count > limit {
				return nil
			}
		}

		now := timeutil.Now()
		multiplier := xpboost.GetEffectiveMultiplier(ctx, tx, userID, now).Effective
		earnedPoints := xpboost.Apply(points, multiplier)

		// 3. Add EXP to User
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("exp", gorm.Expr("exp + ?", earnedPoints)).Error; err != nil {
			return err
		}

		// 4. Record EXP history
		expHistory := model.ExpHistory{
			UserID:       userID,
			ActivityType: string(activityType),
			Points:       int(earnedPoints),
			Description:  getActivityDescription(activityType),
		}
		if err := tx.Create(&expHistory).Error; err != nil {
			return err
		}

		return nil
	})
}

func getDailyLimit(activityType gamificationpkg.ActivityType) int {
	switch activityType {
	case gamificationpkg.ActivityChatAI:
		return gamificationpkg.LimitChatAI
	case gamificationpkg.ActivityForumComment:
		return gamificationpkg.LimitForumComment
	case gamificationpkg.ActivityPostUpvoteGiven:
		return gamificationpkg.LimitPostUpvoteGiven
	case gamificationpkg.ActivityStoryApproved:
		return gamificationpkg.LimitStoryApproved
	case gamificationpkg.ActivityHeartReceived:
		return gamificationpkg.LimitHeartReceived
	default:
		return 0 // No limit
	}
}

func getActivityDescription(activityType gamificationpkg.ActivityType) string {
	switch activityType {
	case gamificationpkg.ActivityChatAI:
		return "Melakukan chat dengan AI"
	case gamificationpkg.ActivityUploadArticle:
		return "Mengunggah artikel baru"
	case gamificationpkg.ActivityForumComment:
		return "Berkomentar di forum"
	case gamificationpkg.ActivityAcceptedAnswer:
		return "Jawaban diterima sebagai solusi terbaik"
	case gamificationpkg.ActivityPostUpvoteGiven:
		return "Menerima upvote pada jawaban forum"
	case gamificationpkg.ActivityPostUpvoteRemoved:
		return "Kehilangan upvote pada jawaban forum"
	case gamificationpkg.ActivityStoryApproved:
		return "Cerita inspiratif disetujui"
	case gamificationpkg.ActivityHeartReceived:
		return "Menerima heart pada cerita inspiratif"
	default:
		return "Aktivitas lainnya"
	}
}
