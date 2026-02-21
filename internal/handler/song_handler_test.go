package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockDailyTaskForSongHandler struct {
	calls    int
	lastUser uint
	lastType model.DailyTaskType
}

func (m *mockDailyTaskForSongHandler) InitializeDailyTasks(_ context.Context, _ uint) error {
	return nil
}
func (m *mockDailyTaskForSongHandler) ProcessDailyLogin(_ context.Context, _ uint) (*service.DailyLoginResult, error) {
	return nil, nil
}
func (m *mockDailyTaskForSongHandler) UpdateTaskProgress(_ context.Context, userID uint, taskType model.DailyTaskType) error {
	m.calls++
	m.lastUser = userID
	m.lastType = taskType
	return nil
}
func (m *mockDailyTaskForSongHandler) GetTodayTasks(_ context.Context, _ uint) (*model.DailyTaskSummary, error) {
	return nil, nil
}
func (m *mockDailyTaskForSongHandler) ClaimTaskReward(_ context.Context, _ uint, _ uint) (*service.ClaimResult, error) {
	return nil, nil
}
func (m *mockDailyTaskForSongHandler) ClaimAllRewards(_ context.Context, _ uint) (*service.ClaimAllResult, error) {
	return nil, nil
}
func (m *mockDailyTaskForSongHandler) GetTaskHistory(_ context.Context, _ uint, _, _ int) (*service.TaskHistoryResult, error) {
	return nil, nil
}

func setupSongHandler(t *testing.T, withSchema bool) *SongHandler {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		schema := []string{
			`CREATE TABLE song_categories (
				id INTEGER PRIMARY KEY,
				name TEXT,
				slug TEXT,
				thumbnail TEXT,
				created_at DATETIME,
				updated_at DATETIME,
				deleted_at DATETIME
			)`,
			`CREATE TABLE songs (
				id INTEGER PRIMARY KEY,
				title TEXT,
				slug TEXT,
				file_path TEXT,
				thumbnail TEXT,
				song_category_id INTEGER,
				created_at DATETIME,
				updated_at DATETIME,
				deleted_at DATETIME
			)`,
			`INSERT INTO song_categories (id, name, slug, thumbnail, created_at, updated_at) VALUES (1, 'Calm', 'calm', 'thumb.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			`INSERT INTO songs (id, title, slug, file_path, thumbnail, song_category_id, created_at, updated_at) VALUES (1, 'Song A', 'song-a', '/a.mp3', 'a.png', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		}
		for _, q := range schema {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup song schema failed: %v", err)
			}
		}
	}

	svc := service.NewSongService(
		repository.NewSongRepository(db),
		repository.NewSongCategoryRepository(db),
		service.NewCacheService(),
	)

	return NewSongHandler(svc)
}

func TestSongHandler_Branches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("set daily task service", func(t *testing.T) {
		h := setupSongHandler(t, true)
		h.SetDailyTaskService(&mockDailyTaskForSongHandler{})
	})

	t.Run("get categories success and error", func(t *testing.T) {
		h1 := setupSongHandler(t, true)
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest(http.MethodGet, "/song-categories", nil)
		h1.GetCategories(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		h2 := setupSongHandler(t, false)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/song-categories", nil)
		h2.GetCategories(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("get songs by category success and error", func(t *testing.T) {
		h := setupSongHandler(t, true)

		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest(http.MethodGet, "/song-categories/calm/songs", nil)
		c1.Params = gin.Params{{Key: "slug", Value: "calm"}}
		h.GetSongsByCategory(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/song-categories/missing/songs", nil)
		c2.Params = gin.Params{{Key: "slug", Value: "missing"}}
		h.GetSongsByCategory(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("get song paths", func(t *testing.T) {
		h := setupSongHandler(t, true)
		daily := &mockDailyTaskForSongHandler{}
		h.SetDailyTaskService(daily)

		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest(http.MethodGet, "/songs/bad", nil)
		c1.Params = gin.Params{{Key: "id", Value: "bad"}}
		h.GetSong(c1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/songs/999", nil)
		c2.Params = gin.Params{{Key: "id", Value: "999"}}
		h.GetSong(c2)
		if w2.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		c3.Request = httptest.NewRequest(http.MethodGet, "/songs/1", nil)
		c3.Params = gin.Params{{Key: "id", Value: "1"}}
		c3.Set("user_id", uint(7))
		h.GetSong(c3)
		if w3.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w3.Code)
		}
		if daily.calls != 1 || daily.lastUser != 7 || daily.lastType != model.TaskTypeListenSongs {
			t.Fatalf("expected one task progress call, got calls=%d user=%d task=%s", daily.calls, daily.lastUser, daily.lastType)
		}
	})
}
