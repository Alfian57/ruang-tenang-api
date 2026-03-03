package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
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
	InitializeDailyTasks(ctx context.Context, userID uint) error

	// Process daily login - update streak and complete login task
	ProcessDailyLogin(ctx context.Context, userID uint) (*DailyLoginResult, error)

	// Update task progress (called by other services when user completes activities)
	UpdateTaskProgress(ctx context.Context, userID uint, taskType model.DailyTaskType) error

	// Get user's daily tasks for today
	GetTodayTasks(ctx context.Context, userID uint) (*model.DailyTaskSummary, error)

	// Claim reward for completed task
	ClaimTaskReward(ctx context.Context, userID uint, taskID uint) (*ClaimResult, error)

	// Claim all completed but unclaimed tasks
	ClaimAllRewards(ctx context.Context, userID uint) (*ClaimAllResult, error)

	// Get task history
	GetTaskHistory(ctx context.Context, userID uint, page, pageSize int) (*TaskHistoryResult, error)
}

type DailyLoginResult struct {
	IsNewDay              bool   `json:"is_new_day"`
	LoginStreak           int    `json:"login_streak"`
	TaskCompleted         bool   `json:"task_completed"`
	XPEarned              int    `json:"xp_earned,omitempty"`
	Message               string `json:"message"`
	StreakBonus           int    `json:"streak_bonus,omitempty"`
	TotalXPFromLogin      int    `json:"total_xp_from_login"`
	StreakFreezeUsed      bool   `json:"streak_freeze_used"`
	StreakFreezeAvailable bool   `json:"streak_freeze_available"`
}

type ClaimResult struct {
	TaskID     uint      `json:"task_id"`
	TaskType   string    `json:"task_type"`
	TaskName   string    `json:"task_name"`
	XPEarned   int       `json:"xp_earned"`
	CoinEarned int       `json:"coin_earned"`
	TotalXP    int64     `json:"total_xp"`
	ClaimedAt  time.Time `json:"claimed_at"`
}

type ClaimAllResult struct {
	ClaimedTasks     []ClaimResult `json:"claimed_tasks"`
	TotalXPEarned    int           `json:"total_xp_earned"`
	TotalCoinsEarned int           `json:"total_coins_earned"`
	TotalClaimed     int           `json:"total_claimed"`
}

type TaskHistoryResult struct {
	History    []model.DailyTaskSummary `json:"history"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

type dailyTaskService struct {
	dailyTaskRepo repository.DailyTaskRepository
	userRepo      *repository.UserRepository
}

func NewDailyTaskService(
	dailyTaskRepo repository.DailyTaskRepository,
	userRepo *repository.UserRepository,
) DailyTaskService {
	return &dailyTaskService{
		dailyTaskRepo: dailyTaskRepo,
		userRepo:      userRepo,
	}
}

func (s *dailyTaskService) InitializeDailyTasks(ctx context.Context, userID uint) error {
	today := timeutil.Now()
	return s.dailyTaskRepo.CreateDailyTasksForUser(ctx, userID, today)
}

func (s *dailyTaskService) ProcessDailyLogin(ctx context.Context, userID uint) (*DailyLoginResult, error) {
	today := timeutil.Now()

	// First, ensure daily tasks exist
	if err := s.InitializeDailyTasks(ctx, userID); err != nil {
		return nil, err
	}

	// Update login streak
	streak, isNewDay, err := s.dailyTaskRepo.UpdateUserLoginStreak(ctx, userID, today)
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
		if err := s.dailyTaskRepo.MarkTaskCompleted(ctx, userID, model.TaskTypeDailyLogin, today); err != nil {
			return nil, err
		}
		result.TaskCompleted = true

		// Calculate streak bonus (e.g., 5% bonus per day, max 50%)
		streakBonus := calculateStreakBonus(streak)
		result.StreakBonus = streakBonus

		// Get the task config for XP
		config := model.GetTaskConfig(model.TaskTypeDailyLogin)
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

	// Fetch updated user to get streak freeze status
	user, err := s.userRepo.FindByID(ctx, userID)
	if err == nil {
		result.StreakFreezeAvailable = user.StreakFreezeAvailable
		result.StreakFreezeUsed = user.StreakFreezeUsedAt != nil && !user.StreakFreezeAvailable
	}

	return result, nil
}

func (s *dailyTaskService) UpdateTaskProgress(ctx context.Context, userID uint, taskType model.DailyTaskType) error {
	today := timeutil.Now()

	// Ensure daily tasks exist
	if err := s.InitializeDailyTasks(ctx, userID); err != nil {
		return err
	}

	return s.dailyTaskRepo.IncrementTaskProgress(ctx, userID, taskType, today)
}

func (s *dailyTaskService) GetTodayTasks(ctx context.Context, userID uint) (*model.DailyTaskSummary, error) {
	today := timeutil.Now()

	// Ensure daily tasks exist
	if err := s.InitializeDailyTasks(ctx, userID); err != nil {
		return nil, err
	}

	return s.dailyTaskRepo.GetDailyTaskSummary(ctx, userID, today)
}

func (s *dailyTaskService) ClaimTaskReward(ctx context.Context, userID uint, taskID uint) (*ClaimResult, error) {
	// Get task
	task, err := s.dailyTaskRepo.GetTaskByID(ctx, taskID)
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
	if err := s.dailyTaskRepo.ClaimTaskReward(ctx, userID, taskID); err != nil {
		return nil, err
	}

	// Get updated user XP
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	task.PopulateTaskInfo()
	return &ClaimResult{
		TaskID:     task.ID,
		TaskType:   string(task.TaskType),
		TaskName:   task.TaskName,
		XPEarned:   task.XPReward,
		CoinEarned: task.CoinReward,
		TotalXP:    user.Exp,
		ClaimedAt:  timeutil.Now(),
	}, nil
}

func (s *dailyTaskService) ClaimAllRewards(ctx context.Context, userID uint) (*ClaimAllResult, error) {
	today := timeutil.Now()

	// Get all tasks for today
	tasks, err := s.dailyTaskRepo.GetUserTasksForDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}

	result := &ClaimAllResult{
		ClaimedTasks: make([]ClaimResult, 0),
	}

	// Claim each completed but unclaimed task
	for _, task := range tasks {
		if task.IsCompleted && !task.IsClaimed {
			claimResult, err := s.ClaimTaskReward(ctx, userID, task.ID)
			if err != nil {
				continue // Skip errors and try next task
			}
			result.ClaimedTasks = append(result.ClaimedTasks, *claimResult)
			result.TotalXPEarned += claimResult.XPEarned
			result.TotalCoinsEarned += claimResult.CoinEarned
			result.TotalClaimed++
		}
	}

	return result, nil
}

func (s *dailyTaskService) GetTaskHistory(ctx context.Context, userID uint, page, pageSize int) (*TaskHistoryResult, error) {
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
	history, total, err := s.dailyTaskRepo.GetTaskHistory(ctx, userID, pageSize, offset)
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
