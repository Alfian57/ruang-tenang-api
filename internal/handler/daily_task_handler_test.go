package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type mockDailyTaskSvcForHandler struct {
	getTodayErr error
	claimErr    error
	loginErr    error
	claimAllErr error
	historyErr  error
	claimAllRes *service.ClaimAllResult

	lastClaimTaskID uint
}

func (m *mockDailyTaskSvcForHandler) InitializeDailyTasks(_ context.Context, _ uint) error {
	return nil
}
func (m *mockDailyTaskSvcForHandler) ProcessDailyLogin(_ context.Context, _ uint) (*service.DailyLoginResult, error) {
	if m.loginErr != nil {
		return nil, m.loginErr
	}
	return &service.DailyLoginResult{Message: "ok"}, nil
}
func (m *mockDailyTaskSvcForHandler) UpdateTaskProgress(_ context.Context, _ uint, _ model.DailyTaskType) error {
	return nil
}
func (m *mockDailyTaskSvcForHandler) GetTodayTasks(_ context.Context, _ uint) (*model.DailyTaskSummary, error) {
	if m.getTodayErr != nil {
		return nil, m.getTodayErr
	}
	return &model.DailyTaskSummary{TotalTasks: 3, CompletedTasks: 1}, nil
}
func (m *mockDailyTaskSvcForHandler) ClaimTaskReward(_ context.Context, _ uint, taskID uint) (*service.ClaimResult, error) {
	m.lastClaimTaskID = taskID
	if m.claimErr != nil {
		return nil, m.claimErr
	}
	return &service.ClaimResult{TaskID: taskID, XPEarned: 20}, nil
}
func (m *mockDailyTaskSvcForHandler) ClaimAllRewards(_ context.Context, _ uint) (*service.ClaimAllResult, error) {
	if m.claimAllErr != nil {
		return nil, m.claimAllErr
	}
	if m.claimAllRes != nil {
		return m.claimAllRes, nil
	}
	return &service.ClaimAllResult{TotalClaimed: 1, TotalXPEarned: 20}, nil
}
func (m *mockDailyTaskSvcForHandler) GetTaskHistory(_ context.Context, _ uint, _, _ int) (*service.TaskHistoryResult, error) {
	if m.historyErr != nil {
		return nil, m.historyErr
	}
	return &service.TaskHistoryResult{Page: 1, PageSize: 7}, nil
}

func TestDailyTaskHandler_GetDailyTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockDailyTaskSvcForHandler{}
	h := NewDailyTaskHandler(svc)

	r := gin.New()
	r.GET("/daily-tasks", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		h.GetDailyTasks(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/daily-tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDailyTaskHandler_ClaimTaskRewardPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid id", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/:id/claim", func(c *gin.Context) {
			c.Set("user_id", uint(5))
			h.ClaimTaskReward(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/daily-tasks/abc/claim", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{claimErr: service.ErrTaskNotFound}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/:id/claim", func(c *gin.Context) {
			c.Set("user_id", uint(5))
			h.ClaimTaskReward(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/daily-tasks/12/claim", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("task not completed", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{claimErr: service.ErrTaskNotCompleted}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/:id/claim", func(c *gin.Context) {
			c.Set("user_id", uint(5))
			h.ClaimTaskReward(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/daily-tasks/12/claim", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("task already claimed", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{claimErr: service.ErrTaskAlreadyClaimed}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/:id/claim", func(c *gin.Context) {
			c.Set("user_id", uint(5))
			h.ClaimTaskReward(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/daily-tasks/12/claim", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{claimErr: errors.New("boom")}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/:id/claim", func(c *gin.Context) {
			c.Set("user_id", uint(5))
			h.ClaimTaskReward(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/daily-tasks/12/claim", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/:id/claim", func(c *gin.Context) {
			c.Set("user_id", uint(5))
			h.ClaimTaskReward(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/daily-tasks/12/claim", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if svc.lastClaimTaskID != 12 {
			t.Fatalf("expected claim task id 12, got %d", svc.lastClaimTaskID)
		}
	})
}

func TestDailyTaskHandler_OtherEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get daily tasks error", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{getTodayErr: errors.New("boom")}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.GET("/daily-tasks", func(c *gin.Context) {
			c.Set("user_id", uint(7))
			h.GetDailyTasks(c)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/daily-tasks", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("claim daily login success and error", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/login", func(c *gin.Context) {
			c.Set("user_id", uint(7))
			h.ClaimDailyLogin(c)
		})

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodPost, "/daily-tasks/login", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		svc.loginErr = errors.New("boom")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/daily-tasks/login", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("claim all rewards branches", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{claimAllRes: &service.ClaimAllResult{TotalClaimed: 0, TotalXPEarned: 0}}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.POST("/daily-tasks/claim-all", func(c *gin.Context) {
			c.Set("user_id", uint(7))
			h.ClaimAllRewards(c)
		})

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodPost, "/daily-tasks/claim-all", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		svc.claimAllErr = errors.New("boom")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/daily-tasks/claim-all", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("get task history success and error", func(t *testing.T) {
		svc := &mockDailyTaskSvcForHandler{}
		h := NewDailyTaskHandler(svc)
		r := gin.New()
		r.GET("/daily-tasks/history", func(c *gin.Context) {
			c.Set("user_id", uint(7))
			h.GetTaskHistory(c)
		})

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/daily-tasks/history?page=2&page_size=9", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		svc.historyErr = errors.New("boom")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/daily-tasks/history", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})
}
