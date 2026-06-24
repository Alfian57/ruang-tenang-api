package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/xpboost"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DailyTaskRepository interface {
	// Task CRUD
	CreateTask(ctx context.Context, task *model.DailyTask) error
	GetTaskByID(ctx context.Context, id uint) (*model.DailyTask, error)
	UpdateTask(ctx context.Context, task *model.DailyTask) error

	// Get tasks for a user
	GetUserTasksForDate(ctx context.Context, userID uint, date time.Time) ([]model.DailyTask, error)
	GetUserTask(ctx context.Context, userID uint, taskType model.DailyTaskType, date time.Time) (*model.DailyTask, error)

	// Batch create daily tasks
	CreateDailyTasksForUser(ctx context.Context, userID uint, date time.Time) error
	HasPremiumAccess(ctx context.Context, userID uint, date time.Time) bool

	// Task progress
	IncrementTaskProgress(ctx context.Context, userID uint, taskType model.DailyTaskType, date time.Time) error
	MarkTaskCompleted(ctx context.Context, userID uint, taskType model.DailyTaskType, date time.Time) error
	ClaimTaskReward(ctx context.Context, userID uint, taskID uint) (int, error)

	// Login tracking
	UpdateUserLoginStreak(ctx context.Context, userID uint, today time.Time) (int, bool, error) // returns streak, isNewDay, error

	// Summary
	GetDailyTaskSummary(ctx context.Context, userID uint, date time.Time) (*model.DailyTaskSummary, error)

	// History
	GetTaskHistory(ctx context.Context, userID uint, limit, offset int) ([]model.DailyTaskSummary, int64, error)
}

type dailyTaskRepository struct {
	db *gorm.DB
}

func NewDailyTaskRepository(db *gorm.DB) DailyTaskRepository {
	return &dailyTaskRepository{db: db}
}

func dayStart(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (r *dailyTaskRepository) HasPremiumAccess(ctx context.Context, userID uint, date time.Time) bool {
	var user model.User
	err := r.db.WithContext(ctx).
		Select("role", "is_premium", "premium_expires_at").
		First(&user, userID).Error
	if err != nil {
		return false
	}

	if user.Role == model.RoleMitra {
		return true
	}

	if user.IsPremium && user.PremiumExpiresAt == nil {
		return true
	}

	dateOnly := dayStart(date)
	if user.IsPremium && !user.PremiumExpiresAt.Before(dateOnly) {
		return true
	}

	return r.hasActiveB2BSeat(ctx, userID, entitlementCheckTime(date))
}

func entitlementCheckTime(date time.Time) time.Time {
	now := time.Now()
	if dayStart(date).Equal(dayStart(now)) {
		return now
	}
	return date
}

func (r *dailyTaskRepository) hasActiveB2BSeat(ctx context.Context, userID uint, at time.Time) bool {
	var count int64
	err := r.db.WithContext(ctx).
		Table("organization_members AS om").
		Joins("JOIN organizations AS o ON o.id = om.organization_id AND o.status = ?", model.OrganizationStatusActive).
		Joins("JOIN b2b_subscriptions AS bs ON bs.organization_id = om.organization_id").
		Joins("JOIN b2b_billing_histories AS bh ON bh.subscription_id = bs.id AND bh.status = ? AND bh.paid_at IS NOT NULL AND bh.billing_period_start <= ? AND bh.billing_period_end > ?", model.B2BBillingHistoryStatusPaid, at, at).
		Joins("JOIN b2b_seat_allocations AS sa ON sa.subscription_id = bs.id AND sa.organization_member_id = om.id AND sa.released_at IS NULL").
		Where("om.user_id = ? AND om.status = ?", userID, model.OrganizationMemberStatusActive).
		Where("bs.status = ? AND bs.starts_at <= ? AND bs.ends_at > ?", model.B2BSubscriptionStatusActive, at, at).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false
	}

	return count > 0
}

func (r *dailyTaskRepository) CreateTask(ctx context.Context, task *model.DailyTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *dailyTaskRepository) GetTaskByID(ctx context.Context, id uint) (*model.DailyTask, error) {
	var task model.DailyTask
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	task.PopulateTaskInfo()
	return &task, nil
}

func (r *dailyTaskRepository) UpdateTask(ctx context.Context, task *model.DailyTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *dailyTaskRepository) GetUserTasksForDate(ctx context.Context, userID uint, date time.Time) ([]model.DailyTask, error) {
	var tasks []model.DailyTask
	dateOnly := dayStart(date)
	err := r.db.WithContext(ctx).Where("user_id = ? AND task_date = ?", userID, dateOnly).
		Order("CASE task_type " +
			"WHEN 'daily_login' THEN 1 " +
			"WHEN 'record_mood' THEN 2 " +
			"WHEN 'chat_ai' THEN 3 " +
			"WHEN 'read_article' THEN 4 " +
			"WHEN 'listen_songs' THEN 5 " +
			"WHEN 'write_journal' THEN 6 " +
			"WHEN 'comment_forum' THEN 7 " +
			"WHEN 'premium_chat_deep_dive' THEN 8 " +
			"WHEN 'premium_breathing_pro' THEN 9 " +
			"ELSE 10 END").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	// Populate task info for each task
	for i := range tasks {
		tasks[i].PopulateTaskInfo()
	}
	return tasks, nil
}

func (r *dailyTaskRepository) GetUserTask(ctx context.Context, userID uint, taskType model.DailyTaskType, date time.Time) (*model.DailyTask, error) {
	var task model.DailyTask
	dateOnly := dayStart(date)
	err := r.db.WithContext(ctx).Where("user_id = ? AND task_type = ? AND task_date = ?", userID, taskType, dateOnly).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	task.PopulateTaskInfo()
	return &task, nil
}

func (r *dailyTaskRepository) CreateDailyTasksForUser(ctx context.Context, userID uint, date time.Time) error {
	dateOnly := dayStart(date)

	var existingTasks []model.DailyTask
	if err := r.db.WithContext(ctx).
		Select("task_type").
		Where("user_id = ? AND task_date = ?", userID, dateOnly).
		Find(&existingTasks).Error; err != nil {
		return err
	}

	existingTaskTypes := make(map[model.DailyTaskType]struct{}, len(existingTasks))
	for _, task := range existingTasks {
		existingTaskTypes[task.TaskType] = struct{}{}
	}

	includePremium := r.HasPremiumAccess(ctx, userID, date)
	configs := model.GetDailyTaskConfigsByTier(includePremium)
	for _, config := range configs {
		if _, exists := existingTaskTypes[config.Type]; exists {
			continue
		}

		task := model.DailyTask{
			UserID:      userID,
			TaskType:    config.Type,
			TaskDate:    dateOnly,
			TargetCount: config.TargetCount,
			XPReward:    config.XPReward,
			CoinReward:  config.CoinReward,
		}
		if err := r.db.WithContext(ctx).Create(&task).Error; err != nil {
			// Ignore duplicate errors
			continue
		}
	}
	return nil
}

func (r *dailyTaskRepository) IncrementTaskProgress(ctx context.Context, userID uint, taskType model.DailyTaskType, date time.Time) error {
	dateOnly := dayStart(date)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task model.DailyTask
		err := tx.Where("user_id = ? AND task_type = ? AND task_date = ?", userID, taskType, dateOnly).
			First(&task).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create the task if it doesn't exist
				config := model.GetTaskConfig(taskType)
				if config == nil {
					return nil
				}
				task = model.DailyTask{
					UserID:       userID,
					TaskType:     taskType,
					TaskDate:     dateOnly,
					TargetCount:  config.TargetCount,
					CurrentCount: 1,
					XPReward:     config.XPReward,
					CoinReward:   config.CoinReward,
				}
				// Check if completed
				if task.CurrentCount >= task.TargetCount {
					task.IsCompleted = true
					now := time.Now()
					task.CompletedAt = &now
				}
				return tx.Create(&task).Error
			}
			return err
		}

		// Don't increment if already completed
		if task.IsCompleted {
			return nil
		}

		// Increment count
		task.CurrentCount++
		if task.CurrentCount >= task.TargetCount {
			task.IsCompleted = true
			now := time.Now()
			task.CompletedAt = &now
		}

		return tx.Save(&task).Error
	})
}

func (r *dailyTaskRepository) MarkTaskCompleted(ctx context.Context, userID uint, taskType model.DailyTaskType, date time.Time) error {
	dateOnly := dayStart(date)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task model.DailyTask
		err := tx.Where("user_id = ? AND task_type = ? AND task_date = ?", userID, taskType, dateOnly).
			First(&task).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create and complete the task
				config := model.GetTaskConfig(taskType)
				if config == nil {
					return nil
				}
				now := time.Now()
				task = model.DailyTask{
					UserID:       userID,
					TaskType:     taskType,
					TaskDate:     dateOnly,
					TargetCount:  config.TargetCount,
					CurrentCount: config.TargetCount,
					IsCompleted:  true,
					CompletedAt:  &now,
					XPReward:     config.XPReward,
					CoinReward:   config.CoinReward,
				}
				return tx.Create(&task).Error
			}
			return err
		}

		// Skip if already completed
		if task.IsCompleted {
			return nil
		}

		// Mark as completed
		task.CurrentCount = task.TargetCount
		task.IsCompleted = true
		now := time.Now()
		task.CompletedAt = &now

		return tx.Save(&task).Error
	})
}

func (r *dailyTaskRepository) ClaimTaskReward(ctx context.Context, userID uint, taskID uint) (int, error) {
	awardedXP := 0

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task model.DailyTask
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error
		if err != nil {
			return err
		}

		// Check if already claimed
		if task.IsClaimed {
			return gorm.ErrRecordNotFound // Or custom error
		}

		// Check if completed
		if !task.IsCompleted {
			return gorm.ErrInvalidData // Or custom error
		}

		// Mark as claimed
		now := timeutil.Now()
		task.IsClaimed = true
		task.ClaimedAt = &now

		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		multiplier := xpboost.GetEffectiveMultiplier(ctx, tx, userID, now).Effective
		awardedXP = int(xpboost.Apply(int64(task.XPReward), multiplier))

		// Award XP to user
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("exp", gorm.Expr("exp + ?", awardedXP)).Error; err != nil {
			return err
		}

		// Award gold coins to user
		if task.CoinReward > 0 {
			return tx.Model(&model.User{}).
				Where("id = ?", userID).
				Update("gold_coins", gorm.Expr("gold_coins + ?", task.CoinReward)).Error
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return awardedXP, nil
}

func (r *dailyTaskRepository) UpdateUserLoginStreak(ctx context.Context, userID uint, today time.Time) (int, bool, error) {
	todayDate := dayStart(today)
	var user model.User
	var isNewDay bool
	var streak int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		// Weekly reset of streak freeze (every Monday, if freeze was used > 7 days ago)
		weekday := todayDate.Weekday()
		if weekday == time.Monday && user.StreakFreezeUsedAt != nil {
			daysSinceUsed := todayDate.Sub(dayStart(*user.StreakFreezeUsedAt)).Hours() / 24
			if daysSinceUsed >= 7 {
				user.StreakFreezeAvailable = true
				user.StreakFreezeUsedAt = nil
			}
		}

		// Check if already logged in today
		if user.LastLoginDate != nil {
			lastLogin := dayStart(*user.LastLoginDate)
			if lastLogin.Equal(todayDate) {
				// Already logged in today
				isNewDay = false
				streak = user.LoginStreak
				return nil
			}

			// Check if logged in yesterday (streak continues)
			yesterday := todayDate.AddDate(0, 0, -1)
			twoDaysAgo := todayDate.AddDate(0, 0, -2)
			if lastLogin.Equal(yesterday) {
				// Streak continues
				user.LoginStreak++
			} else if lastLogin.Equal(twoDaysAgo) && user.StreakFreezeAvailable && user.LoginStreak > 0 {
				// Missed exactly 1 day - use streak freeze
				user.LoginStreak++
				user.StreakFreezeAvailable = false
				freezeDate := todayDate
				user.StreakFreezeUsedAt = &freezeDate
			} else {
				// Streak broken, reset to 1
				user.LoginStreak = 1
			}
		} else {
			// First login ever
			user.LoginStreak = 1
		}

		isNewDay = true
		streak = user.LoginStreak
		user.LastLoginDate = &todayDate

		// Update longest streak if needed
		if user.LoginStreak > user.LongestStreak {
			user.LongestStreak = user.LoginStreak
		}

		return tx.Save(&user).Error
	})

	return streak, isNewDay, err
}

func (r *dailyTaskRepository) GetDailyTaskSummary(ctx context.Context, userID uint, date time.Time) (*model.DailyTaskSummary, error) {
	tasks, err := r.GetUserTasksForDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	// Get user's login streak
	var user model.User
	if err := r.db.WithContext(ctx).Select("login_streak").First(&user, userID).Error; err != nil {
		return nil, err
	}

	summary := &model.DailyTaskSummary{
		Date:        dayStart(date),
		TotalTasks:  len(tasks),
		Tasks:       tasks,
		LoginStreak: user.LoginStreak,
	}

	for _, task := range tasks {
		summary.TotalXPPossible += task.XPReward
		summary.TotalCoinsPossible += task.CoinReward

		if task.IsCompleted {
			summary.CompletedTasks++
		}
		if task.IsClaimed {
			summary.ClaimedTasks++
			summary.TotalXPEarned += task.XPReward
			summary.TotalCoinsEarned += task.CoinReward
		}
	}

	return summary, nil
}

func (r *dailyTaskRepository) GetTaskHistory(ctx context.Context, userID uint, limit, offset int) ([]model.DailyTaskSummary, int64, error) {
	// Get distinct dates with tasks
	var dates []time.Time
	var total int64

	err := r.db.WithContext(ctx).Model(&model.DailyTask{}).
		Select("DISTINCT task_date").
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Model(&model.DailyTask{}).
		Select("DISTINCT task_date").
		Where("user_id = ?", userID).
		Order("task_date DESC").
		Limit(limit).
		Offset(offset).
		Pluck("task_date", &dates).Error
	if err != nil {
		return nil, 0, err
	}

	var summaries []model.DailyTaskSummary
	for _, date := range dates {
		summary, err := r.GetDailyTaskSummary(ctx, userID, date)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}

	return summaries, total, nil
}
