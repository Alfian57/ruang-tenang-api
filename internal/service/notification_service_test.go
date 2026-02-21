package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNotificationService(t *testing.T) (*NotificationService, *repository.NotificationRepository, *gorm.DB, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Notification{}); err != nil {
		t.Fatalf("migrate notification tables: %v", err)
	}

	user := model.User{Name: "Notif User", Username: "notifuser", Email: "notif@test.local", Password: "x"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	repo := repository.NewNotificationRepository(db)
	return NewNotificationService(repo), repo, db, user.ID
}

func TestNotificationService_CreateNotifications(t *testing.T) {
	svc, _, db, userID := setupNotificationService(t)
	ctx := context.Background()

	longTitle := strings.Repeat("A", 80)
	svc.CreateHeartNotification(ctx, userID, "Budi", longTitle, "story-1")
	svc.CreateStoryApprovedNotification(ctx, userID, "Judul Cerita", "story-2")
	svc.CreateStoryRejectedNotification(ctx, userID, "Judul Revisi", "story-3", "perlu perbaikan")
	svc.CreateStoryRejectedNotification(ctx, userID, "Judul Revisi 2", "story-4", "")

	var notifs []model.Notification
	if err := db.Order("created_at ASC").Find(&notifs).Error; err != nil {
		t.Fatalf("query notifications failed: %v", err)
	}
	if len(notifs) != 4 {
		t.Fatalf("expected 4 notifications, got %d", len(notifs))
	}

	if notifs[0].Type != model.NotificationTypeHeart {
		t.Fatalf("expected heart type, got %s", notifs[0].Type)
	}
	if !strings.Contains(notifs[0].Message, "...") {
		t.Fatalf("expected truncated title message, got %s", notifs[0].Message)
	}
	if !strings.Contains(notifs[0].Data, "story-1") {
		t.Fatalf("expected story_id in heart data, got %s", notifs[0].Data)
	}

	if notifs[1].Type != model.NotificationTypeStoryApproved {
		t.Fatalf("expected approved type, got %s", notifs[1].Type)
	}
	if notifs[2].Type != model.NotificationTypeStoryRejected || !strings.Contains(notifs[2].Data, "feedback") {
		t.Fatalf("expected rejected with feedback, got type=%s data=%s", notifs[2].Type, notifs[2].Data)
	}
	if notifs[3].Type != model.NotificationTypeStoryRejected || strings.Contains(notifs[3].Data, "feedback") {
		t.Fatalf("expected rejected without feedback key, got type=%s data=%s", notifs[3].Type, notifs[3].Data)
	}
}

func TestNotificationService_ListAndReadOperations(t *testing.T) {
	svc, repo, db, userID := setupNotificationService(t)
	ctx := context.Background()

	// Seed via service + repo mix to cover list/read branches
	svc.CreateHeartNotification(ctx, userID, "Ana", "Story 1", "s1")
	svc.CreateStoryApprovedNotification(ctx, userID, "Story 2", "s2")

	otherUser := model.User{Name: "Other", Username: "othernotif", Email: "othernotif@test.local", Password: "x"}
	if err := db.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}
	svc.CreateHeartNotification(ctx, otherUser.ID, "Ana", "Other Story", "s3")

	list, err := svc.GetNotifications(ctx, userID, 0, 999, false)
	if err != nil {
		t.Fatalf("get notifications failed: %v", err)
	}
	if list.Page != 1 || list.Limit != 20 {
		t.Fatalf("expected normalized pagination page=1 limit=20, got page=%d limit=%d", list.Page, list.Limit)
	}
	if list.Total != 2 || len(list.Notifications) != 2 {
		t.Fatalf("expected 2 notifications for user, total=%d len=%d", list.Total, len(list.Notifications))
	}
	if list.UnreadCount != 2 {
		t.Fatalf("expected unread count 2, got %d", list.UnreadCount)
	}

	count, err := svc.GetUnreadCount(ctx, userID)
	if err != nil || count != 2 {
		t.Fatalf("get unread count failed: err=%v count=%d", err, count)
	}

	firstID := list.Notifications[0].ID
	if err := svc.MarkAsRead(ctx, firstID, userID); err != nil {
		t.Fatalf("mark as read failed: %v", err)
	}

	unreadOnly, err := svc.GetNotifications(ctx, userID, 1, 20, true)
	if err != nil {
		t.Fatalf("get unread-only notifications failed: %v", err)
	}
	if unreadOnly.Total != 1 || len(unreadOnly.Notifications) != 1 {
		t.Fatalf("expected 1 unread left, total=%d len=%d", unreadOnly.Total, len(unreadOnly.Notifications))
	}

	if err := svc.MarkAllAsRead(ctx, userID); err != nil {
		t.Fatalf("mark all as read failed: %v", err)
	}

	countAfter, err := repo.GetUnreadCount(ctx, userID)
	if err != nil || countAfter != 0 {
		t.Fatalf("expected unread count 0 after mark all, err=%v count=%d", err, countAfter)
	}
}

func TestNotificationService_TruncateTitle(t *testing.T) {
	short := truncateTitle("halo", 10)
	if short != "halo" {
		t.Fatalf("expected unchanged short title, got %s", short)
	}

	long := truncateTitle("abcdefghijklmnopqrstuvwxyz", 10)
	if long != "abcdefg..." {
		t.Fatalf("unexpected truncated title: %s", long)
	}
}
