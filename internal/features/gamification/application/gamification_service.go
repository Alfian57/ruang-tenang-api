package application

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/xpboost"
	gamificationpkg "github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// levelResolver maps an exp amount to its level config (derive-on-read).
type levelResolver interface {
	GetLevelByExp(ctx context.Context, exp int64) (*model.LevelConfig, error)
}

// featureUnlocker unlocks the features tied to a specific level when reached.
type featureUnlocker interface {
	UnlockFeaturesOnLevelUp(ctx context.Context, userID uint, newLevel int) ([]dto.FeatureUnlockResponse, error)
}

// levelUpNotifier sends an in-app/push notification when a user levels up.
type levelUpNotifier interface {
	CreateCustomNotification(ctx context.Context, userID uint, notificationType, title, message string, data map[string]string)
}

type GamificationService struct {
	db *gorm.DB

	// Optional level-up side-effect handlers (injected after construction to
	// avoid import cycles). When nil, AwardExp still updates exp safely.
	levelResolver   levelResolver
	featureUnlocker featureUnlocker
	notifier        levelUpNotifier
}

func NewGamificationService(db *gorm.DB) *GamificationService {
	return &GamificationService{db: db}
}

// SetLevelUpHandlers wires the side-effect collaborators used when a user
// crosses a level boundary. Safe to call once during dependency setup.
func (s *GamificationService) SetLevelUpHandlers(resolver levelResolver, unlocker featureUnlocker, notifier levelUpNotifier) {
	s.levelResolver = resolver
	s.featureUnlocker = unlocker
	s.notifier = notifier
}

// AwardExp adds EXP to a user if the daily limit for the activity hasn't been reached.
//
// After the exp transaction commits, it detects whether the user crossed one or
// more level boundaries (level is derived from exp) and triggers the level-up
// side effects: unlocking the newly available features and sending a level-up
// notification. Exp is clamped at zero so negative awards can never push a user
// below the lowest level.
func (s *GamificationService) AwardExp(ctx context.Context, userID uint, activityType gamificationpkg.ActivityType, points int64) error {
	var (
		oldExp       int64
		newExp       int64
		earnedPoints int64
		expChanged   bool
	)

	err := s.db.Transaction(func(tx *gorm.DB) error {
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
		earnedPoints = xpboost.Apply(points, multiplier)
		if earnedPoints == 0 {
			return nil
		}

		// 2. Capture exp before the change so we can detect a level-up afterwards.
		var before model.User
		if err := tx.Select("id", "exp").Where("id = ?", userID).First(&before).Error; err != nil {
			return err
		}

		// 3. Add EXP to the user, clamped at zero so negative awards (e.g. losing
		//    an upvote) can never drive exp — and therefore level — below zero.
		var after model.User
		if err := tx.Model(&after).
			Clauses(clause.Returning{Columns: []clause.Column{{Name: "exp"}}}).
			Where("id = ?", userID).
			Update("exp", gorm.Expr("GREATEST(exp + ?, 0)", earnedPoints)).Error; err != nil {
			return err
		}

		oldExp = before.Exp
		newExp = after.Exp
		expChanged = true

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
	if err != nil {
		return err
	}

	// Side effects run after the transaction commits (they use their own
	// services/connections). Only a positive exp gain can cross a level up.
	if expChanged && earnedPoints > 0 && newExp > oldExp {
		s.handleLevelUp(ctx, userID, oldExp, newExp)
	}

	return nil
}

// handleLevelUp unlocks features and notifies the user when their exp crossed
// one or more level boundaries. Best-effort: failures are logged, not returned,
// so they never roll back or block the exp award that already succeeded.
func (s *GamificationService) handleLevelUp(ctx context.Context, userID uint, oldExp, newExp int64) {
	if s.levelResolver == nil {
		return
	}

	oldLevelCfg, err := s.levelResolver.GetLevelByExp(ctx, oldExp)
	if err != nil {
		logger.Warn("gamification: failed to resolve old level for level-up check",
			zap.Uint("user_id", userID), zap.Error(err))
		return
	}
	newLevelCfg, err := s.levelResolver.GetLevelByExp(ctx, newExp)
	if err != nil {
		logger.Warn("gamification: failed to resolve new level for level-up check",
			zap.Uint("user_id", userID), zap.Error(err))
		return
	}

	if newLevelCfg.Level <= oldLevelCfg.Level {
		return // no level change
	}

	// Unlock features for every level boundary crossed (a single large award can
	// jump more than one level).
	if s.featureUnlocker != nil {
		for lvl := oldLevelCfg.Level + 1; lvl <= newLevelCfg.Level; lvl++ {
			if _, err := s.featureUnlocker.UnlockFeaturesOnLevelUp(ctx, userID, lvl); err != nil {
				logger.Warn("gamification: failed to unlock features on level up",
					zap.Uint("user_id", userID), zap.Int("level", lvl), zap.Error(err))
			}
		}
	}

	// Send a single notification for the highest level reached.
	if s.notifier != nil {
		message := fmt.Sprintf("Selamat! Kamu naik ke Level %d. Fitur baru telah terbuka untukmu.", newLevelCfg.Level)
		s.notifier.CreateCustomNotification(
			ctx,
			userID,
			string(model.NotificationTypeLevelUp),
			"Naik Level!",
			message,
			map[string]string{"level": strconv.Itoa(newLevelCfg.Level)},
		)
	}
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
