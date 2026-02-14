package repository

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type DailyTaskRepository interface {
	// Task CRUD
	CreateTask(task *model.DailyTask) error
	GetTaskByID(id uint) (*model.DailyTask, error)
	UpdateTask(task *model.DailyTask) error

	// Get tasks for a user
	GetUserTasksForDate(userID uint, date time.Time) ([]model.DailyTask, error)
	GetUserTask(userID uint, taskType model.DailyTaskType, date time.Time) (*model.DailyTask, error)

	// Batch create daily tasks
	CreateDailyTasksForUser(userID uint, date time.Time) error

	// Task progress
	IncrementTaskProgress(userID uint, taskType model.DailyTaskType, date time.Time) error
	MarkTaskCompleted(userID uint, taskType model.DailyTaskType, date time.Time) error
	ClaimTaskReward(userID uint, taskID uint) error

	// Login tracking
	UpdateUserLoginStreak(userID uint, today time.Time) (int, bool, error) // returns streak, isNewDay, error

	// Summary
	GetDailyTaskSummary(userID uint, date time.Time) (*model.DailyTaskSummary, error)

	// History
	GetTaskHistory(userID uint, limit, offset int) ([]model.DailyTaskSummary, int64, error)
}

type dailyTaskRepository struct {
	db *gorm.DB
}

func NewDailyTaskRepository(db *gorm.DB) DailyTaskRepository {
	return &dailyTaskRepository{db: db}
}

func (r *dailyTaskRepository) CreateTask(task *model.DailyTask) error {
	return r.db.Create(task).Error
}

func (r *dailyTaskRepository) GetTaskByID(id uint) (*model.DailyTask, error) {
	var task model.DailyTask
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	task.PopulateTaskInfo()
	return &task, nil
}

func (r *dailyTaskRepository) UpdateTask(task *model.DailyTask) error {
	return r.db.Save(task).Error
}

func (r *dailyTaskRepository) GetUserTasksForDate(userID uint, date time.Time) ([]model.DailyTask, error) {
	var tasks []model.DailyTask
	dateOnly := date.Truncate(24 * time.Hour)
	err := r.db.Where("user_id = ? AND task_date = ?", userID, dateOnly).
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

func (r *dailyTaskRepository) GetUserTask(userID uint, taskType model.DailyTaskType, date time.Time) (*model.DailyTask, error) {
	var task model.DailyTask
	dateOnly := date.Truncate(24 * time.Hour)
	err := r.db.Where("user_id = ? AND task_type = ? AND task_date = ?", userID, taskType, dateOnly).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	task.PopulateTaskInfo()
	return &task, nil
}

func (r *dailyTaskRepository) CreateDailyTasksForUser(userID uint, date time.Time) error {
	dateOnly := date.Truncate(24 * time.Hour)

	// Check if tasks already exist for today
	var count int64
	r.db.Model(&model.DailyTask{}).
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
		}
		if err := r.db.Create(&task).Error; err != nil {
			// Ignore duplicate errors
			continue
		}
	}
	return nil
}

func (r *dailyTaskRepository) IncrementTaskProgress(userID uint, taskType model.DailyTaskType, date time.Time) error {
	dateOnly := date.Truncate(24 * time.Hour)

	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *dailyTaskRepository) MarkTaskCompleted(userID uint, taskType model.DailyTaskType, date time.Time) error {
	dateOnly := date.Truncate(24 * time.Hour)

	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *dailyTaskRepository) ClaimTaskReward(userID uint, taskID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

		// Award XP to user
		return tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("exp", gorm.Expr("exp + ?", task.XPReward)).Error
	})
}

func (r *dailyTaskRepository) UpdateUserLoginStreak(userID uint, today time.Time) (int, bool, error) {
	todayDate := today.Truncate(24 * time.Hour)
	var user model.User
	var isNewDay bool
	var streak int

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			return err
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
			if lastLogin.Equal(yesterday) {
				// Streak continues
				user.LoginStreak++
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

func (r *dailyTaskRepository) GetDailyTaskSummary(userID uint, date time.Time) (*model.DailyTaskSummary, error) {
	tasks, err := r.GetUserTasksForDate(userID, date)
	if err != nil {
		return nil, err
	}

	// Get user's login streak
	var user model.User
	if err := r.db.Select("login_streak").First(&user, userID).Error; err != nil {
		return nil, err
	}

	summary := &model.DailyTaskSummary{
		Date:            date.Truncate(24 * time.Hour),
		TotalTasks:      len(tasks),
		TotalXPPossible: model.GetTotalPossibleXP(),
		Tasks:           tasks,
		LoginStreak:     user.LoginStreak,
	}

	for _, task := range tasks {
		if task.IsCompleted {
			summary.CompletedTasks++
		}
		if task.IsClaimed {
			summary.ClaimedTasks++
			summary.TotalXPEarned += task.XPReward
		}
	}

	return summary, nil
}

func (r *dailyTaskRepository) GetTaskHistory(userID uint, limit, offset int) ([]model.DailyTaskSummary, int64, error) {
	// Get distinct dates with tasks
	var dates []time.Time
	var total int64

	err := r.db.Model(&model.DailyTask{}).
		Select("DISTINCT task_date").
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Model(&model.DailyTask{}).
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
		summary, err := r.GetDailyTaskSummary(userID, date)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}

	return summaries, total, nil
}
