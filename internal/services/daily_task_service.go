package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"gorm.io/gorm"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskNotCompleted   = errors.New("task not completed yet")
	ErrTaskAlreadyClaimed = errors.New("task already claimed")
)

type DailyTaskService interface {
	// Initialize daily tasks for user (called on login or first access)
	InitializeDailyTasks(userID uint) error

	// Process daily login - update streak and complete login task
	ProcessDailyLogin(userID uint) (*DailyLoginResult, error)

	// Update task progress (called by other services when user completes activities)
	UpdateTaskProgress(userID uint, taskType models.DailyTaskType) error

	// Get user's daily tasks for today
	GetTodayTasks(userID uint) (*models.DailyTaskSummary, error)

	// Claim reward for completed task
	ClaimTaskReward(userID uint, taskID uint) (*ClaimResult, error)

	// Claim all completed but unclaimed tasks
	ClaimAllRewards(userID uint) (*ClaimAllResult, error)

	// Get task history
	GetTaskHistory(userID uint, page, pageSize int) (*TaskHistoryResult, error)
}

type DailyLoginResult struct {
	IsNewDay         bool   `json:"is_new_day"`
	LoginStreak      int    `json:"login_streak"`
	TaskCompleted    bool   `json:"task_completed"`
	XPEarned         int    `json:"xp_earned,omitempty"`
	Message          string `json:"message"`
	StreakBonus      int    `json:"streak_bonus,omitempty"`
	TotalXPFromLogin int    `json:"total_xp_from_login"`
}

type ClaimResult struct {
	TaskID    uint      `json:"task_id"`
	TaskType  string    `json:"task_type"`
	TaskName  string    `json:"task_name"`
	XPEarned  int       `json:"xp_earned"`
	TotalXP   int64     `json:"total_xp"`
	ClaimedAt time.Time `json:"claimed_at"`
}

type ClaimAllResult struct {
	ClaimedTasks  []ClaimResult `json:"claimed_tasks"`
	TotalXPEarned int           `json:"total_xp_earned"`
	TotalClaimed  int           `json:"total_claimed"`
}

type TaskHistoryResult struct {
	History    []models.DailyTaskSummary `json:"history"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

type dailyTaskService struct {
	dailyTaskRepo repositories.DailyTaskRepository
	userRepo      *repositories.UserRepository
}

func NewDailyTaskService(
	dailyTaskRepo repositories.DailyTaskRepository,
	userRepo *repositories.UserRepository,
) DailyTaskService {
	return &dailyTaskService{
		dailyTaskRepo: dailyTaskRepo,
		userRepo:      userRepo,
	}
}

func (s *dailyTaskService) InitializeDailyTasks(userID uint) error {
	today := timeutil.Now()
	return s.dailyTaskRepo.CreateDailyTasksForUser(userID, today)
}

func (s *dailyTaskService) ProcessDailyLogin(userID uint) (*DailyLoginResult, error) {
	today := timeutil.Now()

	// First, ensure daily tasks exist
	if err := s.InitializeDailyTasks(userID); err != nil {
		return nil, err
	}

	// Update login streak
	streak, isNewDay, err := s.dailyTaskRepo.UpdateUserLoginStreak(userID, today)
	if err != nil {
		return nil, err
	}

	result := &DailyLoginResult{
		IsNewDay:    isNewDay,
		LoginStreak: streak,
		Message:     "Selamat datang kembali!",
	}

	if isNewDay {
		// Mark login task as completed
		if err := s.dailyTaskRepo.MarkTaskCompleted(userID, models.TaskTypeDailyLogin, today); err != nil {
			return nil, err
		}
		result.TaskCompleted = true

		// Calculate streak bonus (e.g., 5% bonus per day, max 50%)
		streakBonus := calculateStreakBonus(streak)
		result.StreakBonus = streakBonus

		// Get the task config for XP
		config := models.GetTaskConfig(models.TaskTypeDailyLogin)
		if config != nil {
			result.XPEarned = config.XPReward
			bonusXP := (config.XPReward * streakBonus) / 100
			result.TotalXPFromLogin = config.XPReward + bonusXP
		}

		// Update message based on streak
		if streak == 1 {
			result.Message = "Selamat datang! Ini login pertama kamu hari ini."
		} else if streak <= 7 {
			result.Message = "Luar biasa! Kamu sudah login " + formatStreak(streak) + " berturut-turut."
		} else if streak <= 30 {
			result.Message = "Hebat! Streak login kamu sudah " + formatStreak(streak) + "! Terus semangat!"
		} else {
			result.Message = "Wow! " + formatStreak(streak) + " login berturut-turut! Kamu sangat konsisten!"
		}
	} else {
		result.Message = "Kamu sudah klaim login hari ini. Tetap semangat!"
	}

	return result, nil
}

func (s *dailyTaskService) UpdateTaskProgress(userID uint, taskType models.DailyTaskType) error {
	today := timeutil.Now()

	// Ensure daily tasks exist
	if err := s.InitializeDailyTasks(userID); err != nil {
		return err
	}

	return s.dailyTaskRepo.IncrementTaskProgress(userID, taskType, today)
}

func (s *dailyTaskService) GetTodayTasks(userID uint) (*models.DailyTaskSummary, error) {
	today := timeutil.Now()

	// Ensure daily tasks exist
	if err := s.InitializeDailyTasks(userID); err != nil {
		return nil, err
	}

	return s.dailyTaskRepo.GetDailyTaskSummary(userID, today)
}

func (s *dailyTaskService) ClaimTaskReward(userID uint, taskID uint) (*ClaimResult, error) {
	// Get task
	task, err := s.dailyTaskRepo.GetTaskByID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, ErrTaskNotFound
	}

	// Check if completed
	if !task.IsCompleted {
		return nil, ErrTaskNotCompleted
	}

	// Check if already claimed
	if task.IsClaimed {
		return nil, ErrTaskAlreadyClaimed
	}

	// Claim reward
	if err := s.dailyTaskRepo.ClaimTaskReward(userID, taskID); err != nil {
		return nil, err
	}

	// Get updated user XP
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	task.PopulateTaskInfo()
	return &ClaimResult{
		TaskID:    task.ID,
		TaskType:  string(task.TaskType),
		TaskName:  task.TaskName,
		XPEarned:  task.XPReward,
		TotalXP:   user.Exp,
		ClaimedAt: timeutil.Now(),
	}, nil
}

func (s *dailyTaskService) ClaimAllRewards(userID uint) (*ClaimAllResult, error) {
	today := timeutil.Now()

	// Get all tasks for today
	tasks, err := s.dailyTaskRepo.GetUserTasksForDate(userID, today)
	if err != nil {
		return nil, err
	}

	result := &ClaimAllResult{
		ClaimedTasks: make([]ClaimResult, 0),
	}

	// Claim each completed but unclaimed task
	for _, task := range tasks {
		if task.IsCompleted && !task.IsClaimed {
			claimResult, err := s.ClaimTaskReward(userID, task.ID)
			if err != nil {
				continue // Skip errors and try next task
			}
			result.ClaimedTasks = append(result.ClaimedTasks, *claimResult)
			result.TotalXPEarned += claimResult.XPEarned
			result.TotalClaimed++
		}
	}

	return result, nil
}

func (s *dailyTaskService) GetTaskHistory(userID uint, page, pageSize int) (*TaskHistoryResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 7
	}
	if pageSize > 30 {
		pageSize = 30
	}

	offset := (page - 1) * pageSize
	history, total, err := s.dailyTaskRepo.GetTaskHistory(userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &TaskHistoryResult{
		History:    history,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Helper functions
func calculateStreakBonus(streak int) int {
	// 5% bonus per day, capped at 50% (10 days)
	bonus := (streak - 1) * 5
	if bonus > 50 {
		bonus = 50
	}
	if bonus < 0 {
		bonus = 0
	}
	return bonus
}

func formatStreak(streak int) string {
	return fmt.Sprintf("%d hari", streak)
}
