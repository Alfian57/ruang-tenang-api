package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.ForumCategory{}, &model.DailyTask{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	return db
}

func seedUser(t *testing.T, db *gorm.DB, id uint) {
	t.Helper()
	u := model.User{
		ID:       id,
		Name:     "Tester",
		Username: "tester_user",
		Email:    "tester@example.com",
		Password: "hashed",
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
}

func TestForumCategoryRepository_SQLiteCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewForumCategoryRepository(db)
	ctx := context.Background()

	category := &model.ForumCategory{Name: "General"}
	if err := repo.Create(ctx, category); err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	if category.ID == 0 || category.Slug == "" {
		t.Fatalf("expected id and slug to be set, got %+v", category)
	}

	all, err := repo.FindAll(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("find all failed: err=%v len=%d", err, len(all))
	}

	byID, err := repo.FindByID(ctx, category.ID)
	if err != nil || byID.ID != category.ID {
		t.Fatalf("find by id failed: err=%v result=%+v", err, byID)
	}

	bySlug, err := repo.FindBySlug(ctx, category.Slug)
	if err != nil || bySlug.ID != category.ID {
		t.Fatalf("find by slug failed: err=%v result=%+v", err, bySlug)
	}

	bySlug.Name = "Updated"
	if err := repo.Update(ctx, bySlug); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, _ := repo.FindByID(ctx, category.ID)
	if updated.Name != "Updated" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}

	if err := repo.Delete(ctx, category.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, category.ID); err == nil {
		t.Fatal("expected find by id to fail after delete")
	}
}

func TestDailyTaskRepository_SQLiteFlow(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db, 1)

	repo := NewDailyTaskRepository(db)
	ctx := context.Background()
	today := time.Now().Truncate(24 * time.Hour)

	if err := repo.CreateDailyTasksForUser(ctx, 1, today); err != nil {
		t.Fatalf("create daily tasks failed: %v", err)
	}

	tasks, err := repo.GetUserTasksForDate(ctx, 1, today)
	if err != nil || len(tasks) == 0 {
		t.Fatalf("get user tasks failed: err=%v len=%d", err, len(tasks))
	}

	if err := repo.IncrementTaskProgress(ctx, 1, model.TaskTypeChatAI, today); err != nil {
		t.Fatalf("increment progress failed: %v", err)
	}
	task, err := repo.GetUserTask(ctx, 1, model.TaskTypeChatAI, today)
	if err != nil {
		t.Fatalf("get user task failed: %v", err)
	}
	if task.CurrentCount < 1 {
		t.Fatalf("expected current count to increment, got %d", task.CurrentCount)
	}

	if err := repo.MarkTaskCompleted(ctx, 1, model.TaskTypeDailyLogin, today); err != nil {
		t.Fatalf("mark task completed failed: %v", err)
	}
	loginTask, err := repo.GetUserTask(ctx, 1, model.TaskTypeDailyLogin, today)
	if err != nil {
		t.Fatalf("get daily login task failed: %v", err)
	}
	if !loginTask.IsCompleted {
		t.Fatal("expected daily login task to be completed")
	}

	if err := repo.ClaimTaskReward(ctx, 1, loginTask.ID); err != nil {
		t.Fatalf("claim task reward failed: %v", err)
	}

	var user model.User
	if err := db.First(&user, 1).Error; err != nil {
		t.Fatalf("load user failed: %v", err)
	}
	if user.Exp <= 0 {
		t.Fatalf("expected user exp to increase, got %d", user.Exp)
	}

	streak, isNewDay, err := repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil {
		t.Fatalf("update login streak failed: %v", err)
	}
	if !isNewDay || streak <= 0 {
		t.Fatalf("expected first login update to be new day with streak > 0, got new=%v streak=%d", isNewDay, streak)
	}

	streak2, isNewDay2, err := repo.UpdateUserLoginStreak(ctx, 1, today)
	if err != nil {
		t.Fatalf("second same-day login update failed: %v", err)
	}
	if isNewDay2 {
		t.Fatal("expected same-day login to return isNewDay=false")
	}
	if streak2 != streak {
		t.Fatalf("expected same streak on same day, got %d vs %d", streak2, streak)
	}

	summary, err := repo.GetDailyTaskSummary(ctx, 1, today)
	if err != nil {
		t.Fatalf("get daily summary failed: %v", err)
	}
	if summary.TotalTasks == 0 {
		t.Fatal("expected summary to include tasks")
	}

	history, total, err := repo.GetTaskHistory(ctx, 1, 10, 0)
	if err != nil {
		t.Fatalf("get task history failed: %v", err)
	}
	if total == 0 || len(history) == 0 {
		t.Fatalf("expected non-empty history, total=%d len=%d", total, len(history))
	}
}
