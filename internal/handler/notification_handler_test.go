package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNotificationHandler(t *testing.T, withSchema bool) *NotificationHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		if err := db.Exec(`CREATE TABLE notifications (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			type TEXT,
			title TEXT,
			message TEXT,
			is_read BOOLEAN,
			data TEXT,
			created_at DATETIME
		)`).Error; err != nil {
			t.Fatalf("create notifications table: %v", err)
		}
		if err := db.Exec(`INSERT INTO notifications (id, user_id, type, title, message, is_read, data, created_at) VALUES ('11111111-1111-1111-1111-111111111111', 1, 'heart', 't', 'm', 0, '{}', CURRENT_TIMESTAMP)`).Error; err != nil {
			t.Fatalf("seed notification: %v", err)
		}
	}

	svc := service.NewNotificationService(repository.NewNotificationRepository(db))
	return NewNotificationHandler(svc)
}

func TestNotificationHandler_MarkAsReadInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNotificationHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/notifications/bad/read", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}

	h.MarkAsRead(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationHandler_Branches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get notifications success and error", func(t *testing.T) {
		h1 := setupNotificationHandler(t, true)
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest(http.MethodGet, "/notifications?page=1&limit=10&unread_only=true", nil)
		c1.Set("user_id", uint(1))
		h1.GetNotifications(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		h2 := setupNotificationHandler(t, false)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/notifications", nil)
		c2.Set("user_id", uint(1))
		h2.GetNotifications(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("get unread count success and error", func(t *testing.T) {
		h1 := setupNotificationHandler(t, true)
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
		c1.Set("user_id", uint(1))
		h1.GetUnreadCount(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		h2 := setupNotificationHandler(t, false)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
		c2.Set("user_id", uint(1))
		h2.GetUnreadCount(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("mark as read success", func(t *testing.T) {
		h := setupNotificationHandler(t, true)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/notifications/11111111-1111-1111-1111-111111111111/read", nil)
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.MarkAsRead(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("mark all as read success and error", func(t *testing.T) {
		h1 := setupNotificationHandler(t, true)
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		c1.Set("user_id", uint(1))
		h1.MarkAllAsRead(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		h2 := setupNotificationHandler(t, false)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		c2.Set("user_id", uint(1))
		h2.MarkAllAsRead(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})
}
