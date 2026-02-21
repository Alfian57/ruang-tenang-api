package repository

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNotificationRepo(t *testing.T, withSchema bool) *NotificationRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		queries := []string{
			`CREATE TABLE notifications (
				id TEXT PRIMARY KEY,
				user_id INTEGER,
				type TEXT,
				title TEXT,
				message TEXT,
				is_read BOOLEAN,
				data TEXT,
				created_at DATETIME
			)`,
			`INSERT INTO notifications (id, user_id, type, title, message, is_read, created_at)
				VALUES ('00000000-0000-0000-0000-000000000001', 1, 'heart', 'T1', 'M1', 0, CURRENT_TIMESTAMP),
				       ('00000000-0000-0000-0000-000000000002', 1, 'heart', 'T2', 'M2', 1, CURRENT_TIMESTAMP),
				       ('00000000-0000-0000-0000-000000000003', 2, 'heart', 'T3', 'M3', 0, CURRENT_TIMESTAMP)`,
		}
		for _, q := range queries {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup query failed: %v", err)
			}
		}
	}

	return NewNotificationRepository(db)
}

func TestNotificationRepository_BasicFlows(t *testing.T) {
	ctx := context.Background()
	repo := setupNotificationRepo(t, true)

	n := &model.Notification{ID: uuid.New(), UserID: 1, Type: model.NotificationTypeHeart, Title: "new", Message: "msg"}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("create notification failed: %v", err)
	}

	all, total, err := repo.GetByUserID(ctx, 1, 1, 10, false)
	if err != nil || total == 0 || len(all) == 0 {
		t.Fatalf("get by user all failed: total=%d len=%d err=%v", total, len(all), err)
	}

	unread, unreadTotal, err := repo.GetByUserID(ctx, 1, 1, 10, true)
	if err != nil || unreadTotal == 0 || len(unread) == 0 {
		t.Fatalf("get by user unread failed: total=%d len=%d err=%v", unreadTotal, len(unread), err)
	}

	count, err := repo.GetUnreadCount(ctx, 1)
	if err != nil || count == 0 {
		t.Fatalf("get unread count failed: count=%d err=%v", count, err)
	}

	if err := repo.MarkAsRead(ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001"), 1); err != nil {
		t.Fatalf("mark as read failed: %v", err)
	}

	if err := repo.MarkAllAsRead(ctx, 1); err != nil {
		t.Fatalf("mark all as read failed: %v", err)
	}
}

func TestNotificationRepository_DBErrorBranches(t *testing.T) {
	ctx := context.Background()
	repo := setupNotificationRepo(t, false)

	if err := repo.Create(ctx, &model.Notification{ID: uuid.New(), UserID: 1, Type: model.NotificationTypeHeart, Title: "x", Message: "y"}); err == nil {
		t.Fatal("expected create error on missing schema")
	}
	if _, _, err := repo.GetByUserID(ctx, 1, 1, 10, false); err == nil {
		t.Fatal("expected get by user error on missing schema")
	}
	if _, err := repo.GetUnreadCount(ctx, 1); err == nil {
		t.Fatal("expected get unread count error on missing schema")
	}
	if err := repo.MarkAsRead(ctx, uuid.New(), 1); err == nil {
		t.Fatal("expected mark as read error on missing schema")
	}
	if err := repo.MarkAllAsRead(ctx, 1); err == nil {
		t.Fatal("expected mark all as read error on missing schema")
	}
}
