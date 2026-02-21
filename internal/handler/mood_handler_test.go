package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type moodDailyTaskMock struct {
	called bool
}

func (m *moodDailyTaskMock) InitializeDailyTasks(_ context.Context, _ uint) error { return nil }
func (m *moodDailyTaskMock) ProcessDailyLogin(_ context.Context, _ uint) (*service.DailyLoginResult, error) {
	return &service.DailyLoginResult{}, nil
}
func (m *moodDailyTaskMock) UpdateTaskProgress(_ context.Context, _ uint, _ model.DailyTaskType) error {
	m.called = true
	return nil
}
func (m *moodDailyTaskMock) GetTodayTasks(_ context.Context, _ uint) (*model.DailyTaskSummary, error) {
	return &model.DailyTaskSummary{}, nil
}
func (m *moodDailyTaskMock) ClaimTaskReward(_ context.Context, _ uint, _ uint) (*service.ClaimResult, error) {
	return &service.ClaimResult{}, nil
}
func (m *moodDailyTaskMock) ClaimAllRewards(_ context.Context, _ uint) (*service.ClaimAllResult, error) {
	return &service.ClaimAllResult{}, nil
}
func (m *moodDailyTaskMock) GetTaskHistory(_ context.Context, _ uint, _, _ int) (*service.TaskHistoryResult, error) {
	return &service.TaskHistoryResult{}, nil
}

func setupMoodHandlerWithDB(t *testing.T) (*MoodHandler, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE user_moods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		mood TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create user_moods: %v", err)
	}

	svc := service.NewMoodService(repository.NewUserMoodRepository(db))
	return NewMoodHandler(svc), db
}

func newMoodTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestMoodHandler_InvalidInputBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMoodHandler(nil)

	{
		c, w := newMoodTestContext(http.MethodPost, "/user-moods", "{")
		h.RecordMood(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 RecordMood, got %d", w.Code)
		}
	}

	{
		c, w := newMoodTestContext(http.MethodGet, "/user-moods?page=abc", "")
		h.GetMoodHistory(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 GetMoodHistory, got %d", w.Code)
		}
	}
}

func TestMoodHandler_RecordMoodBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success and daily task update", func(t *testing.T) {
		h, _ := setupMoodHandlerWithDB(t)
		daily := &moodDailyTaskMock{}
		h.SetDailyTaskService(daily)

		r := gin.New()
		r.POST("/user-moods", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.RecordMood(c)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/user-moods", strings.NewReader(`{"mood":"happy"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
		if !daily.called {
			t.Fatal("expected daily task UpdateTaskProgress to be called")
		}
	})

	t.Run("service error", func(t *testing.T) {
		h, db := setupMoodHandlerWithDB(t)
		if err := db.Exec(`DROP TABLE user_moods`).Error; err != nil {
			t.Fatalf("drop table: %v", err)
		}

		r := gin.New()
		r.POST("/user-moods", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.RecordMood(c)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/user-moods", strings.NewReader(`{"mood":"happy"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestMoodHandler_ReadBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get mood history success and server error", func(t *testing.T) {
		h, db := setupMoodHandlerWithDB(t)
		if err := db.Exec(`INSERT INTO user_moods (user_id, mood, created_at, updated_at) VALUES (1, 'happy', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
			t.Fatalf("seed mood: %v", err)
		}

		r := gin.New()
		r.GET("/user-moods", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.GetMoodHistory(c)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user-moods?page=1&limit=10", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		if err := db.Exec(`DROP TABLE user_moods`).Error; err != nil {
			t.Fatalf("drop table: %v", err)
		}
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/user-moods?page=1&limit=10", nil)
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("get latest mood success and not found", func(t *testing.T) {
		h, db := setupMoodHandlerWithDB(t)

		r := gin.New()
		r.GET("/user-moods/latest", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.GetLatestMood(c)
		})

		wNotFound := httptest.NewRecorder()
		reqNotFound := httptest.NewRequest(http.MethodGet, "/user-moods/latest", nil)
		r.ServeHTTP(wNotFound, reqNotFound)
		if wNotFound.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", wNotFound.Code)
		}

		if err := db.Exec(`INSERT INTO user_moods (user_id, mood, created_at, updated_at) VALUES (1, 'neutral', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
			t.Fatalf("seed latest mood: %v", err)
		}
		wOK := httptest.NewRecorder()
		reqOK := httptest.NewRequest(http.MethodGet, "/user-moods/latest", nil)
		r.ServeHTTP(wOK, reqOK)
		if wOK.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wOK.Code)
		}
	})

	t.Run("get stats success and error", func(t *testing.T) {
		h, db := setupMoodHandlerWithDB(t)

		r := gin.New()
		r.GET("/user-moods/stats", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.GetMoodStats(c)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user-moods/stats", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		if err := db.Exec(`DROP TABLE user_moods`).Error; err != nil {
			t.Fatalf("drop table: %v", err)
		}
		wErr := httptest.NewRecorder()
		reqErr := httptest.NewRequest(http.MethodGet, "/user-moods/stats", nil)
		r.ServeHTTP(wErr, reqErr)
		if wErr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", wErr.Code)
		}
	})

	t.Run("check today mood", func(t *testing.T) {
		h, _ := setupMoodHandlerWithDB(t)
		r := gin.New()
		r.GET("/user-moods/today", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.CheckTodayMood(c)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user-moods/today", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
