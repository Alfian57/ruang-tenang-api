package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/xpboost"
	"gorm.io/gorm"
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
	dateOnly := date.Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Where("user_id = ? AND task_date = ?", userID, dateOnly).
		Order("CASE task_type " +
			"WHEN 'daily_login' THEN 1 " +
			"WHEN 'record_mood' THEN 2 " +
			"WHEN 'chat_ai' THEN 3 " +
			"WHEN 'read_article' THEN 4 " +
			"WHEN 'listen_songs' THEN 5 " +
			"WHEN 'write_journal' THEN 6 " +
			"WHEN 'comment_forum' THEN 7 " +
			"ELSE 8 END").
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
	dateOnly := date.Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Where("user_id = ? AND task_type = ? AND task_date = ?", userID, taskType, dateOnly).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	task.PopulateTaskInfo()
	return &task, nil
}

func (r *dailyTaskRepository) CreateDailyTasksForUser(ctx context.Context, userID uint, date time.Time) error {
	dateOnly := date.Truncate(24 * time.Hour)

	// Check if tasks already exist for today
	var count int64
	r.db.WithContext(ctx).Model(&model.DailyTask{}).
		Where("user_id = ? AND task_date = ?", userID, dateOnly).
		Count(&count)

	if count > 0 {
		return nil // Tasks already exist
	}

	// Create all daily tasks
	configs := model.GetDailyTaskConfigs()
	for _, config := range configs {
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
	dateOnly := date.Truncate(24 * time.Hour)

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
	dateOnly := date.Truncate(24 * time.Hour)

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
		err := tx.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error
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
		now := time.Now()
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
	todayDate := today.Truncate(24 * time.Hour)
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
			daysSinceUsed := todayDate.Sub(user.StreakFreezeUsedAt.Truncate(24*time.Hour)).Hours() / 24
			if daysSinceUsed >= 7 {
				user.StreakFreezeAvailable = true
				user.StreakFreezeUsedAt = nil
			}
		}

		// Check if already logged in today
		if user.LastLoginDate != nil {
			lastLogin := user.LastLoginDate.Truncate(24 * time.Hour)
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
		Date:               date.Truncate(24 * time.Hour),
		TotalTasks:         len(tasks),
		TotalXPPossible:    model.GetTotalPossibleXP(),
		TotalCoinsPossible: model.GetTotalPossibleCoins(),
		Tasks:              tasks,
		LoginStreak:        user.LoginStreak,
	}

	for _, task := range tasks {
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
