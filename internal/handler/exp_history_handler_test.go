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

func TestExpHistoryHandler_GetHistoryInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExpHistoryHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/exp-history?page=abc", nil)

	h.GetHistory(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func setupExpHistoryHandler(t *testing.T, withSchema bool) *ExpHistoryHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		if err := db.Exec(`CREATE TABLE exp_histories (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			points INTEGER,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`).Error; err != nil {
			t.Fatalf("create exp_histories table: %v", err)
		}
		if err := db.Exec(`INSERT INTO exp_histories (id, user_id, activity_type, points, description, created_at, updated_at) VALUES (1, 7, 'chat_ai', 10, 'test', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
			t.Fatalf("seed exp_history: %v", err)
		}
	}

	svc := service.NewExpHistoryService(repository.NewExpHistoryRepository(db))
	return NewExpHistoryHandler(svc, nil)
}

func TestExpHistoryHandler_SuccessAndErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get-history-success", func(t *testing.T) {
		h := setupExpHistoryHandler(t, true)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/exp-history?page=1&limit=10", nil)
		c.Set("user_id", uint(7))

		h.GetHistory(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-history-service-error", func(t *testing.T) {
		h := setupExpHistoryHandler(t, false)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/exp-history", nil)
		c.Set("user_id", uint(7))

		h.GetHistory(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("get-activity-types-success", func(t *testing.T) {
		h := setupExpHistoryHandler(t, true)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/exp-history/activity-types", nil)

		h.GetActivityTypes(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-activity-types-error", func(t *testing.T) {
		h := setupExpHistoryHandler(t, false)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/exp-history/activity-types", nil)

		h.GetActivityTypes(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
