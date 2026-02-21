package service

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newExpHistoryServiceForTest(t *testing.T) (*ExpHistoryService, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ExpHistory{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}

	user := &model.User{Name: "Exp User", Username: "expuser", Email: "exp@test.local", Password: "x"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	entries := []model.ExpHistory{
		{UserID: user.ID, ActivityType: "chat_ai", Points: 10, Description: "Chat", CreatedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},
		{UserID: user.ID, ActivityType: "article_read", Points: 5, Description: "Read", CreatedAt: time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC)},
		{UserID: user.ID, ActivityType: "chat_ai", Points: 15, Description: "Chat 2", CreatedAt: time.Date(2026, 2, 5, 10, 0, 0, 0, time.UTC)},
	}
	for _, entry := range entries {
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("create exp history: %v", err)
		}
	}

	repo := repository.NewExpHistoryRepository(db)
	return NewExpHistoryService(repo), user.ID
}

func TestExpHistoryService_GetHistoryAndActivityTypes(t *testing.T) {
	ctx := context.Background()
	svc, userID := newExpHistoryServiceForTest(t)

	histories, total, err := svc.GetHistory(ctx, userID, &dto.ExpHistoryFilterRequest{
		ActivityType: "chat_ai",
		StartDate:    "2026-02-01",
		EndDate:      "2026-02-28",
		Page:         1,
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if total != 2 || len(histories) != 2 {
		t.Fatalf("expected 2 chat_ai histories, total=%d len=%d", total, len(histories))
	}

	// Invalid date should be ignored gracefully
	histories2, total2, err := svc.GetHistory(ctx, userID, &dto.ExpHistoryFilterRequest{
		StartDate: "invalid-date",
		EndDate:   "invalid-date",
		Page:      1,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("GetHistory with invalid date failed: %v", err)
	}
	if total2 != 3 || len(histories2) != 3 {
		t.Fatalf("expected all histories when date parsing fails, total=%d len=%d", total2, len(histories2))
	}

	types, err := svc.GetActivityTypes(ctx)
	if err != nil {
		t.Fatalf("GetActivityTypes failed: %v", err)
	}
	if len(types) < 2 {
		t.Fatalf("expected at least 2 activity types, got %d", len(types))
	}
}
