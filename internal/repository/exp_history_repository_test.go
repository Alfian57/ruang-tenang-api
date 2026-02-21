package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newExpHistoryRepoForTest(t *testing.T) (*ExpHistoryRepository, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ExpHistory{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}

	user := &model.User{Name: "Repo User", Username: "repouser", Email: "repo@test.local", Password: "x"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	repo := NewExpHistoryRepository(db)
	return repo, user.ID
}

func TestExpHistoryRepository_CreateGetAndAggregate(t *testing.T) {
	ctx := context.Background()
	repo, userID := newExpHistoryRepoForTest(t)

	entries := []model.ExpHistory{
		{UserID: userID, ActivityType: "chat_ai", Points: 10, Description: "A", CreatedAt: time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)},
		{UserID: userID, ActivityType: "article_read", Points: 5, Description: "B", CreatedAt: time.Date(2026, 2, 2, 9, 0, 0, 0, time.UTC)},
		{UserID: userID, ActivityType: "chat_ai", Points: 15, Description: "C", CreatedAt: time.Date(2026, 2, 3, 9, 0, 0, 0, time.UTC)},
	}
	for _, entry := range entries {
		copy := entry
		if err := repo.Create(ctx, &copy); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	totalExp, err := repo.GetTotalExpByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetTotalExpByUserID failed: %v", err)
	}
	if totalExp != 30 {
		t.Fatalf("expected total exp 30, got %d", totalExp)
	}

	types, err := repo.GetActivityTypes(ctx)
	if err != nil {
		t.Fatalf("GetActivityTypes failed: %v", err)
	}
	if len(types) < 2 {
		t.Fatalf("expected at least 2 distinct types, got %d", len(types))
	}
}

func TestExpHistoryRepository_GetByUserIDFilters(t *testing.T) {
	ctx := context.Background()
	repo, userID := newExpHistoryRepoForTest(t)

	seed := []model.ExpHistory{
		{UserID: userID, ActivityType: "chat_ai", Points: 10, Description: "A", CreatedAt: time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)},
		{UserID: userID, ActivityType: "chat_ai", Points: 15, Description: "B", CreatedAt: time.Date(2026, 2, 5, 8, 0, 0, 0, time.UTC)},
		{UserID: userID, ActivityType: "article_read", Points: 5, Description: "C", CreatedAt: time.Date(2026, 2, 10, 8, 0, 0, 0, time.UTC)},
	}
	for _, entry := range seed {
		copy := entry
		if err := repo.Create(ctx, &copy); err != nil {
			t.Fatalf("seed create failed: %v", err)
		}
	}

	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 6, 0, 0, 0, 0, time.UTC)
	histories, total, err := repo.GetByUserID(ctx, ExpHistoryFilter{
		UserID:       userID,
		ActivityType: "chat_ai",
		StartDate:    &start,
		EndDate:      &end,
		Page:         1,
		Limit:        1,
	})
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(histories) != 1 {
		t.Fatalf("expected paginated len 1, got %d", len(histories))
	}
}
