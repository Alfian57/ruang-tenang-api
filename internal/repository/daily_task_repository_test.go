package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDailyTaskRepo(t *testing.T, withSchema bool) DailyTaskRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		queries := []string{
			`CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				name TEXT,
				username TEXT,
				email TEXT,
				password TEXT,
				role TEXT,
				exp INTEGER,
				avatar TEXT,
				is_blocked BOOLEAN,
				is_forum_blocked BOOLEAN,
				reset_token TEXT,
				reset_token_expiry DATETIME,
				suspension_end DATETIME,
				suspension_reason TEXT,
				is_banned BOOLEAN,
				ban_reason TEXT,
				has_accepted_ai_disclaimer BOOLEAN,
				content_warning_preference TEXT,
				profile_theme TEXT,
				profile_banner TEXT,
				avatar_border_color TEXT,
				tagline TEXT,
				bio TEXT,
				current_streak INTEGER,
				last_activity_date DATETIME,
				total_activities INTEGER,
				login_streak INTEGER,
				longest_streak INTEGER,
				last_login_date DATETIME,
				streak_freeze_available BOOLEAN,
				streak_freeze_used_at DATETIME,
				created_at DATETIME,
				updated_at DATETIME,
				deleted_at DATETIME
			)`,
			`CREATE TABLE daily_tasks (
				id INTEGER PRIMARY KEY,
				user_id INTEGER,
				task_type TEXT,
				task_date DATETIME,
				target_count INTEGER,
				current_count INTEGER,
				is_completed BOOLEAN,
				is_claimed BOOLEAN,
				xp_reward INTEGER,
				completed_at DATETIME,
				claimed_at DATETIME,
				created_at DATETIME,
				updated_at DATETIME
			)`,
			`INSERT INTO users (id, name, username, email, password, role, exp, avatar, is_blocked, is_forum_blocked, login_streak, longest_streak, streak_freeze_available, total_activities, created_at, updated_at)
				VALUES (1, 'U1', 'u1', 'u1@test.id', 'x', 'member', 0, '', 0, 0, 0, 0, 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		}
		for _, q := range queries {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup query failed: %v", err)
			}
		}
	}

	return NewDailyTaskRepository(db)
}

func TestDailyTaskRepository_MainFlows(t *testing.T) {
	ctx := context.Background()
	repo := setupDailyTaskRepo(t, true)
	today := time.Now()

	base := &model.DailyTask{UserID: 1, TaskType: model.TaskTypeDailyLogin, TaskDate: today.Truncate(24 * time.Hour), TargetCount: 1, XPReward: 5}
	if err := repo.CreateTask(ctx, base); err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	got, err := repo.GetTaskByID(ctx, base.ID)
	if err != nil || got == nil || got.TaskName == "" {
		t.Fatalf("get task by id failed: %+v err=%v", got, err)
	}

	got.CurrentCount = 1
	if err := repo.UpdateTask(ctx, got); err != nil {
		t.Fatalf("update task failed: %v", err)
	}

	if err := repo.CreateDailyTasksForUser(ctx, 1, today); err != nil {
		t.Fatalf("create daily tasks failed: %v", err)
	}
	if err := repo.CreateDailyTasksForUser(ctx, 1, today); err != nil {
		t.Fatalf("create daily tasks idempotent failed: %v", err)
	}

	tasks, err := repo.GetUserTasksForDate(ctx, 1, today)
	if err != nil || len(tasks) == 0 {
		t.Fatalf("get user tasks for date failed: len=%d err=%v", len(tasks), err)
	}

	one, err := repo.GetUserTask(ctx, 1, model.TaskTypeDailyLogin, today)
	if err != nil || one == nil {
		t.Fatalf("get user task failed: %+v err=%v", one, err)
	}

	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeRecordMood, today); err != nil {
		t.Fatalf("increment missing task failed: %v", err)
	}
	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeRecordMood, today); err != nil {
		t.Fatalf("increment existing task failed: %v", err)
	}

	if err := repo.MarkTaskCompleted(ctx, 1, model.TaskTypeReadArticle, today); err != nil {
		t.Fatalf("mark task completed create-if-missing failed: %v", err)
	}
	if err := repo.MarkTaskCompleted(ctx, 1, model.TaskTypeReadArticle, today); err != nil {
		t.Fatalf("mark task completed idempotent failed: %v", err)
	}

	claimTask := &model.DailyTask{UserID: 1, TaskType: model.TaskTypeCommentForum, TaskDate: today.Truncate(24 * time.Hour), TargetCount: 1, CurrentCount: 1, IsCompleted: true, XPReward: 15}
	if err := repo.CreateTask(ctx, claimTask); err != nil {
		t.Fatalf("create claim task failed: %v", err)
	}
	if err := repo.ClaimTaskReward(ctx, 1, claimTask.ID); err != nil {
		t.Fatalf("claim task reward failed: %v", err)
	}

	streak, isNewDay, err := repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil || !isNewDay || streak < 1 {
		t.Fatalf("update login streak first login failed: streak=%d new=%v err=%v", streak, isNewDay, err)
	}
	streak2, isNewDay2, err := repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil || isNewDay2 || streak2 < 1 {
		t.Fatalf("update login streak same day failed: streak=%d new=%v err=%v", streak2, isNewDay2, err)
	}

	summary, err := repo.GetDailyTaskSummary(ctx, 1, today)
	if err != nil || summary == nil || summary.TotalTasks == 0 {
		t.Fatalf("get daily task summary failed: %+v err=%v", summary, err)
	}

	history, total, err := repo.GetTaskHistory(ctx, 1, 10, 0)
	if err != nil || total == 0 || len(history) == 0 {
		t.Fatalf("get task history failed: total=%d len=%d err=%v", total, len(history), err)
	}
}

func TestDailyTaskRepository_ErrorBranches(t *testing.T) {
	ctx := context.Background()
	repo := setupDailyTaskRepo(t, false)
	today := time.Now()

	if err := repo.CreateTask(ctx, &model.DailyTask{UserID: 1, TaskType: model.TaskTypeDailyLogin, TaskDate: today}); err == nil {
		t.Fatal("expected create task error on missing schema")
	}
	if _, err := repo.GetTaskByID(ctx, 1); err == nil {
		t.Fatal("expected get task by id error on missing schema")
	}
	if err := repo.UpdateTask(ctx, &model.DailyTask{ID: 1}); err == nil {
		t.Fatal("expected update task error on missing schema")
	}
	if _, err := repo.GetUserTasksForDate(ctx, 1, today); err == nil {
		t.Fatal("expected get user tasks error on missing schema")
	}
	if _, err := repo.GetUserTask(ctx, 1, model.TaskTypeDailyLogin, today); err == nil {
		t.Fatal("expected get user task error on missing schema")
	}
	if err := repo.CreateDailyTasksForUser(ctx, 1, today); err != nil {
		t.Fatalf("expected tolerant create daily tasks behavior, got %v", err)
	}
	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeDailyLogin, today); err == nil {
		t.Fatal("expected increment task progress error on missing schema")
	}
	if err := repo.MarkTaskCompleted(ctx, 1, model.TaskTypeDailyLogin, today); err == nil {
		t.Fatal("expected mark task completed error on missing schema")
	}
	if err := repo.ClaimTaskReward(ctx, 1, 1); err == nil {
		t.Fatal("expected claim reward error on missing schema")
	}
	if _, _, err := repo.UpdateUserLoginStreak(ctx, 1, today); err == nil {
		t.Fatal("expected update login streak error on missing schema")
	}
	if _, err := repo.GetDailyTaskSummary(ctx, 1, today); err == nil {
		t.Fatal("expected get daily summary error on missing schema")
	}
	if _, _, err := repo.GetTaskHistory(ctx, 1, 10, 0); err == nil {
		t.Fatal("expected get task history error on missing schema")
	}
}

func TestDailyTaskRepository_IncrementAndClaimBranches(t *testing.T) {
	ctx := context.Background()
	repo := setupDailyTaskRepo(t, true)
	r := repo.(*dailyTaskRepository)
	today := time.Now().Truncate(24 * time.Hour)

	completedTask := &model.DailyTask{
		UserID:       1,
		TaskType:     model.TaskTypeReadArticle,
		TaskDate:     today,
		TargetCount:  1,
		CurrentCount: 1,
		IsCompleted:  true,
		XPReward:     20,
	}
	if err := repo.CreateTask(ctx, completedTask); err != nil {
		t.Fatalf("create completed task: %v", err)
	}

	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeReadArticle, today); err != nil {
		t.Fatalf("increment completed task should be no-op, got %v", err)
	}

	var after model.DailyTask
	if err := r.db.Where("id = ?", completedTask.ID).First(&after).Error; err != nil {
		t.Fatalf("reload completed task: %v", err)
	}
	if after.CurrentCount != 1 {
		t.Fatalf("expected completed task count unchanged, got %d", after.CurrentCount)
	}

	if err := repo.IncrementTaskProgress(ctx, 1, model.DailyTaskType("unknown_task"), today); err != nil {
		t.Fatalf("unknown task type should be ignored, got %v", err)
	}

	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeDailyLogin, today.AddDate(0, 0, 1)); err != nil {
		t.Fatalf("create-and-complete branch should succeed for target=1 task, got %v", err)
	}
	createdCompleted, err := repo.GetUserTask(ctx, 1, model.TaskTypeDailyLogin, today.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("load created-completed task: %v", err)
	}
	if !createdCompleted.IsCompleted || createdCompleted.CompletedAt == nil {
		t.Fatalf("expected created task to be immediately completed, got %+v", createdCompleted)
	}

	almostDone := &model.DailyTask{UserID: 1, TaskType: model.TaskTypeChatAI, TaskDate: today.AddDate(0, 0, 2), TargetCount: 2, CurrentCount: 1, IsCompleted: false, XPReward: 30}
	if err := repo.CreateTask(ctx, almostDone); err != nil {
		t.Fatalf("create almost-done task: %v", err)
	}
	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeChatAI, today.AddDate(0, 0, 2)); err != nil {
		t.Fatalf("increment should complete existing task at target, got %v", err)
	}
	completedExisting, err := repo.GetUserTask(ctx, 1, model.TaskTypeChatAI, today.AddDate(0, 0, 2))
	if err != nil {
		t.Fatalf("load completed existing task: %v", err)
	}
	if !completedExisting.IsCompleted || completedExisting.CompletedAt == nil {
		t.Fatalf("expected existing task to be marked completed, got %+v", completedExisting)
	}

	notCompleted := &model.DailyTask{UserID: 1, TaskType: model.TaskTypeWriteJournal, TaskDate: today, TargetCount: 1, CurrentCount: 0, IsCompleted: false, XPReward: 40}
	if err := repo.CreateTask(ctx, notCompleted); err != nil {
		t.Fatalf("create not-completed task: %v", err)
	}
	if err := repo.ClaimTaskReward(ctx, 1, notCompleted.ID); err == nil {
		t.Fatal("expected claim reward error for incomplete task")
	}

	alreadyClaimed := &model.DailyTask{UserID: 1, TaskType: model.TaskTypeChatAI, TaskDate: today, TargetCount: 1, CurrentCount: 1, IsCompleted: true, IsClaimed: true, XPReward: 30}
	if err := repo.CreateTask(ctx, alreadyClaimed); err != nil {
		t.Fatalf("create claimed task: %v", err)
	}
	if err := repo.ClaimTaskReward(ctx, 1, alreadyClaimed.ID); err == nil {
		t.Fatal("expected claim reward error for already claimed task")
	}
}

func TestDailyTaskRepository_UpdateUserLoginStreakBranches(t *testing.T) {
	ctx := context.Background()
	repo := setupDailyTaskRepo(t, true)
	r := repo.(*dailyTaskRepository)

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	if err := r.db.Model(&model.User{}).Where("id = 1").Updates(map[string]any{
		"login_streak":            3,
		"longest_streak":          3,
		"last_login_date":         yesterday,
		"streak_freeze_available": true,
		"streak_freeze_used_at":   nil,
	}).Error; err != nil {
		t.Fatalf("seed yesterday login state: %v", err)
	}

	streak, isNewDay, err := repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil || !isNewDay || streak != 4 {
		t.Fatalf("expected yesterday streak continuation to 4, streak=%d new=%v err=%v", streak, isNewDay, err)
	}

	if err := r.db.Model(&model.User{}).Where("id = 1").Updates(map[string]any{
		"login_streak":            5,
		"longest_streak":          5,
		"last_login_date":         twoDaysAgo,
		"streak_freeze_available": true,
		"streak_freeze_used_at":   nil,
	}).Error; err != nil {
		t.Fatalf("seed two-days-ago state: %v", err)
	}

	streak, isNewDay, err = repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil || !isNewDay || streak != 6 {
		t.Fatalf("expected streak freeze branch to continue to 6, streak=%d new=%v err=%v", streak, isNewDay, err)
	}

	var user model.User
	if err := r.db.First(&user, 1).Error; err != nil {
		t.Fatalf("load user after freeze usage: %v", err)
	}
	if user.StreakFreezeAvailable {
		t.Fatal("expected streak freeze to be consumed")
	}
	if user.StreakFreezeUsedAt == nil {
		t.Fatal("expected streak freeze used date to be set")
	}

	if err := r.db.Model(&model.User{}).Where("id = 1").Updates(map[string]any{
		"login_streak":            9,
		"longest_streak":          9,
		"last_login_date":         today.AddDate(0, 0, -4),
		"streak_freeze_available": false,
	}).Error; err != nil {
		t.Fatalf("seed broken streak state: %v", err)
	}

	streak, isNewDay, err = repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil || !isNewDay || streak != 1 {
		t.Fatalf("expected broken streak reset to 1, streak=%d new=%v err=%v", streak, isNewDay, err)
	}
}
