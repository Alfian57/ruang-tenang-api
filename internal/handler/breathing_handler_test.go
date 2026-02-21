package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockBreathingServiceForHandler struct {
	getAllTechniquesFn   func(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error)
	getTechniqueByIDFn   func(ctx context.Context, userID uint, techniqueID uuid.UUID) (*dto.BreathingTechniqueResponse, error)
	getTechniqueBySlugFn func(ctx context.Context, userID uint, slug string) (*dto.BreathingTechniqueResponse, error)
	createTechniqueFn    func(ctx context.Context, userID uint, req dto.CreateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error)
	updateTechniqueFn    func(ctx context.Context, userID uint, techniqueID uuid.UUID, req dto.UpdateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error)
	deleteTechniqueFn    func(ctx context.Context, userID uint, techniqueID uuid.UUID) error
	startSessionFn       func(ctx context.Context, userID uint, req dto.StartBreathingSessionRequest) (*dto.BreathingSessionResponse, error)
	completeSessionFn    func(ctx context.Context, userID uint, sessionID uuid.UUID, req dto.CompleteBreathingSessionRequest) (*dto.SessionCompletionResult, error)
	getSessionHistoryFn  func(ctx context.Context, userID uint, req dto.SessionHistoryRequest) (*dto.SessionHistoryResponse, error)
	getSessionByIDFn     func(ctx context.Context, userID uint, sessionID uuid.UUID) (*dto.BreathingSessionResponse, error)
	getPreferencesFn     func(ctx context.Context, userID uint) (*dto.BreathingPreferencesResponse, error)
	updatePreferencesFn  func(ctx context.Context, userID uint, req dto.UpdateBreathingPreferencesRequest) (*dto.BreathingPreferencesResponse, error)
	getFavoritesFn       func(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error)
	addFavoriteFn        func(ctx context.Context, userID uint, techniqueID uuid.UUID) error
	removeFavoriteFn     func(ctx context.Context, userID uint, techniqueID uuid.UUID) error
	reorderFavoritesFn   func(ctx context.Context, userID uint, ids []uuid.UUID) error
	getStatsFn           func(ctx context.Context, userID uint) (*dto.BreathingStatsResponse, error)
	getCalendarFn        func(ctx context.Context, userID uint, year, month int) (*dto.BreathingCalendarResponse, error)
	getTechniqueUsageFn  func(ctx context.Context, userID uint) ([]dto.TechniqueUsageStats, error)
	getWidgetDataFn      func(ctx context.Context, userID uint) (*dto.BreathingWidgetData, error)
	getRecommendationsFn func(ctx context.Context, userID uint, mood, timeOfDay string) (*dto.RecommendationsResponse, error)
}

func (m *mockBreathingServiceForHandler) GetAllTechniques(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error) {
	if m.getAllTechniquesFn != nil {
		return m.getAllTechniquesFn(ctx, userID)
	}
	return []dto.BreathingTechniqueResponse{}, nil
}
func (m *mockBreathingServiceForHandler) GetTechniqueByID(ctx context.Context, userID uint, techniqueID uuid.UUID) (*dto.BreathingTechniqueResponse, error) {
	if m.getTechniqueByIDFn != nil {
		return m.getTechniqueByIDFn(ctx, userID, techniqueID)
	}
	return &dto.BreathingTechniqueResponse{ID: techniqueID}, nil
}
func (m *mockBreathingServiceForHandler) GetTechniqueBySlug(_ context.Context, _ uint, _ string) (*dto.BreathingTechniqueResponse, error) {
	if m.getTechniqueBySlugFn != nil {
		return m.getTechniqueBySlugFn(context.Background(), 0, "")
	}
	return &dto.BreathingTechniqueResponse{}, nil
}
func (m *mockBreathingServiceForHandler) CreateCustomTechnique(ctx context.Context, userID uint, req dto.CreateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
	if m.createTechniqueFn != nil {
		return m.createTechniqueFn(ctx, userID, req)
	}
	return &dto.BreathingTechniqueResponse{Name: req.Name}, nil
}
func (m *mockBreathingServiceForHandler) UpdateCustomTechnique(_ context.Context, _ uint, _ uuid.UUID, _ dto.UpdateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
	if m.updateTechniqueFn != nil {
		return m.updateTechniqueFn(context.Background(), 0, uuid.Nil, dto.UpdateBreathingTechniqueRequest{})
	}
	return &dto.BreathingTechniqueResponse{}, nil
}
func (m *mockBreathingServiceForHandler) DeleteCustomTechnique(_ context.Context, _ uint, _ uuid.UUID) error {
	if m.deleteTechniqueFn != nil {
		return m.deleteTechniqueFn(context.Background(), 0, uuid.Nil)
	}
	return nil
}
func (m *mockBreathingServiceForHandler) StartSession(_ context.Context, _ uint, _ dto.StartBreathingSessionRequest) (*dto.BreathingSessionResponse, error) {
	if m.startSessionFn != nil {
		return m.startSessionFn(context.Background(), 0, dto.StartBreathingSessionRequest{})
	}
	return &dto.BreathingSessionResponse{}, nil
}
func (m *mockBreathingServiceForHandler) CompleteSession(ctx context.Context, userID uint, sessionID uuid.UUID, req dto.CompleteBreathingSessionRequest) (*dto.SessionCompletionResult, error) {
	if m.completeSessionFn != nil {
		return m.completeSessionFn(ctx, userID, sessionID, req)
	}
	return &dto.SessionCompletionResult{TotalXP: 10, Session: dto.BreathingSessionResponse{ID: sessionID}}, nil
}
func (m *mockBreathingServiceForHandler) GetSessionHistory(_ context.Context, _ uint, _ dto.SessionHistoryRequest) (*dto.SessionHistoryResponse, error) {
	if m.getSessionHistoryFn != nil {
		return m.getSessionHistoryFn(context.Background(), 0, dto.SessionHistoryRequest{})
	}
	return &dto.SessionHistoryResponse{}, nil
}
func (m *mockBreathingServiceForHandler) GetSessionByID(_ context.Context, _ uint, _ uuid.UUID) (*dto.BreathingSessionResponse, error) {
	if m.getSessionByIDFn != nil {
		return m.getSessionByIDFn(context.Background(), 0, uuid.Nil)
	}
	return &dto.BreathingSessionResponse{}, nil
}
func (m *mockBreathingServiceForHandler) GetPreferences(_ context.Context, _ uint) (*dto.BreathingPreferencesResponse, error) {
	if m.getPreferencesFn != nil {
		return m.getPreferencesFn(context.Background(), 0)
	}
	return &dto.BreathingPreferencesResponse{}, nil
}
func (m *mockBreathingServiceForHandler) UpdatePreferences(_ context.Context, _ uint, _ dto.UpdateBreathingPreferencesRequest) (*dto.BreathingPreferencesResponse, error) {
	if m.updatePreferencesFn != nil {
		return m.updatePreferencesFn(context.Background(), 0, dto.UpdateBreathingPreferencesRequest{})
	}
	return &dto.BreathingPreferencesResponse{}, nil
}
func (m *mockBreathingServiceForHandler) GetFavorites(_ context.Context, _ uint) ([]dto.BreathingTechniqueResponse, error) {
	if m.getFavoritesFn != nil {
		return m.getFavoritesFn(context.Background(), 0)
	}
	return []dto.BreathingTechniqueResponse{}, nil
}
func (m *mockBreathingServiceForHandler) AddFavorite(_ context.Context, _ uint, _ uuid.UUID) error {
	if m.addFavoriteFn != nil {
		return m.addFavoriteFn(context.Background(), 0, uuid.Nil)
	}
	return nil
}
func (m *mockBreathingServiceForHandler) RemoveFavorite(_ context.Context, _ uint, _ uuid.UUID) error {
	if m.removeFavoriteFn != nil {
		return m.removeFavoriteFn(context.Background(), 0, uuid.Nil)
	}
	return nil
}
func (m *mockBreathingServiceForHandler) ReorderFavorites(_ context.Context, _ uint, _ []uuid.UUID) error {
	if m.reorderFavoritesFn != nil {
		return m.reorderFavoritesFn(context.Background(), 0, nil)
	}
	return nil
}
func (m *mockBreathingServiceForHandler) GetStats(_ context.Context, _ uint) (*dto.BreathingStatsResponse, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(context.Background(), 0)
	}
	return &dto.BreathingStatsResponse{}, nil
}
func (m *mockBreathingServiceForHandler) GetCalendar(_ context.Context, _ uint, _, _ int) (*dto.BreathingCalendarResponse, error) {
	if m.getCalendarFn != nil {
		return m.getCalendarFn(context.Background(), 0, 0, 0)
	}
	return &dto.BreathingCalendarResponse{}, nil
}
func (m *mockBreathingServiceForHandler) GetTechniqueUsage(_ context.Context, _ uint) ([]dto.TechniqueUsageStats, error) {
	if m.getTechniqueUsageFn != nil {
		return m.getTechniqueUsageFn(context.Background(), 0)
	}
	return []dto.TechniqueUsageStats{}, nil
}
func (m *mockBreathingServiceForHandler) GetWidgetData(_ context.Context, _ uint) (*dto.BreathingWidgetData, error) {
	if m.getWidgetDataFn != nil {
		return m.getWidgetDataFn(context.Background(), 0)
	}
	return &dto.BreathingWidgetData{}, nil
}
func (m *mockBreathingServiceForHandler) GetRecommendations(_ context.Context, _ uint, _, _ string) (*dto.RecommendationsResponse, error) {
	if m.getRecommendationsFn != nil {
		return m.getRecommendationsFn(context.Background(), 0, "", "")
	}
	return &dto.RecommendationsResponse{}, nil
}

func TestBreathingHandler_BasicPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetTechniques success", func(t *testing.T) {
		svc := &mockBreathingServiceForHandler{
			getAllTechniquesFn: func(_ context.Context, _ uint) ([]dto.BreathingTechniqueResponse, error) {
				return []dto.BreathingTechniqueResponse{{Name: "Box"}}, nil
			},
		}
		h := NewBreathingHandler(svc)
		r := gin.New()
		r.GET("/techniques", func(c *gin.Context) {
			c.Set("user_id", uint(9))
			h.GetTechniques(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/techniques", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetTechniques service error", func(t *testing.T) {
		svc := &mockBreathingServiceForHandler{
			getAllTechniquesFn: func(_ context.Context, _ uint) ([]dto.BreathingTechniqueResponse, error) {
				return nil, errors.New("repo fail")
			},
		}
		h := NewBreathingHandler(svc)
		r := gin.New()
		r.GET("/techniques", func(c *gin.Context) {
			c.Set("user_id", uint(9))
			h.GetTechniques(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/techniques", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestBreathingHandler_InvalidIDsAndCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetTechniqueByID invalid UUID", func(t *testing.T) {
		h := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r := gin.New()
		r.GET("/techniques/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.GetTechniqueByID(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/techniques/not-uuid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetTechniqueByID not found and success", func(t *testing.T) {
		techID := uuid.New()

		hNotFound := NewBreathingHandler(&mockBreathingServiceForHandler{getTechniqueByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*dto.BreathingTechniqueResponse, error) {
			return nil, errors.New("not found")
		}})
		r1 := gin.New()
		r1.GET("/techniques/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			hNotFound.GetTechniqueByID(c)
		})
		w1 := httptest.NewRecorder()
		r1.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/techniques/"+techID.String(), nil))
		if w1.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w1.Code)
		}

		hOK := NewBreathingHandler(&mockBreathingServiceForHandler{getTechniqueByIDFn: func(_ context.Context, _ uint, id uuid.UUID) (*dto.BreathingTechniqueResponse, error) {
			return &dto.BreathingTechniqueResponse{ID: id, Name: "Box Breath"}, nil
		}})
		r2 := gin.New()
		r2.GET("/techniques/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			hOK.GetTechniqueByID(c)
		})
		w2 := httptest.NewRecorder()
		r2.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/techniques/"+techID.String(), nil))
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}
	})

	t.Run("CreateTechnique bad payload", func(t *testing.T) {
		h := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r := gin.New()
		r.POST("/techniques", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.CreateTechnique(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/techniques", bytes.NewBufferString(`{"name":123}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("CreateTechnique service error and success", func(t *testing.T) {
		validPayload := `{"name":"Calm Box","inhale_duration":4,"exhale_duration":4}`

		hErr := NewBreathingHandler(&mockBreathingServiceForHandler{createTechniqueFn: func(_ context.Context, _ uint, _ dto.CreateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
			return nil, errors.New("cannot create")
		}})
		r1 := gin.New()
		r1.POST("/techniques", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			hErr.CreateTechnique(c)
		})
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPost, "/techniques", bytes.NewBufferString(validPayload))
		req1.Header.Set("Content-Type", "application/json")
		r1.ServeHTTP(w1, req1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		hOK := NewBreathingHandler(&mockBreathingServiceForHandler{createTechniqueFn: func(_ context.Context, _ uint, req dto.CreateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
			return &dto.BreathingTechniqueResponse{Name: req.Name}, nil
		}})
		r2 := gin.New()
		r2.POST("/techniques", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			hOK.CreateTechnique(c)
		})
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/techniques", bytes.NewBufferString(validPayload))
		req2.Header.Set("Content-Type", "application/json")
		r2.ServeHTTP(w2, req2)
		if w2.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w2.Code)
		}
	})

	t.Run("CompleteSession invalid session ID", func(t *testing.T) {
		h := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r := gin.New()
		r.POST("/sessions/:id/complete", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.CompleteSession(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/sessions/bad-id/complete", bytes.NewBufferString(`{"duration_seconds":120}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("CompleteSession success", func(t *testing.T) {
		sessionID := uuid.New()
		svc := &mockBreathingServiceForHandler{
			completeSessionFn: func(_ context.Context, _ uint, id uuid.UUID, _ dto.CompleteBreathingSessionRequest) (*dto.SessionCompletionResult, error) {
				return &dto.SessionCompletionResult{TotalXP: 15, Session: dto.BreathingSessionResponse{ID: id, StartedAt: time.Now()}}, nil
			},
		}
		h := NewBreathingHandler(svc)
		r := gin.New()
		r.POST("/sessions/:id/complete", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			h.CompleteSession(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/complete", bytes.NewBufferString(`{"duration_seconds":120,"completed":true}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestBreathingHandler_RemainingEndpointBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetTechniqueBySlug error and success", func(t *testing.T) {
		hErr := NewBreathingHandler(&mockBreathingServiceForHandler{getTechniqueBySlugFn: func(_ context.Context, _ uint, _ string) (*dto.BreathingTechniqueResponse, error) {
			return nil, errors.New("not found")
		}})
		r1 := gin.New()
		r1.GET("/techniques/slug/:slug", func(c *gin.Context) { c.Set("user_id", uint(1)); hErr.GetTechniqueBySlug(c) })
		w1 := httptest.NewRecorder()
		r1.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/techniques/slug/box", nil))
		if w1.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w1.Code)
		}

		hOK := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r2 := gin.New()
		r2.GET("/techniques/slug/:slug", func(c *gin.Context) { c.Set("user_id", uint(1)); hOK.GetTechniqueBySlug(c) })
		w2 := httptest.NewRecorder()
		r2.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/techniques/slug/box", nil))
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}
	})

	t.Run("Update and delete technique branches", func(t *testing.T) {
		h := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r := gin.New()
		r.PUT("/techniques/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.UpdateTechnique(c) })
		r.DELETE("/techniques/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.DeleteTechnique(c) })

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPut, "/techniques/bad", bytes.NewBufferString(`{}`))
		req1.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		techID := uuid.New()
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPut, "/techniques/"+techID.String(), bytes.NewBufferString(`{"name":"Updated"}`))
		req2.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, "/techniques/bad", nil))
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, httptest.NewRequest(http.MethodDelete, "/techniques/"+techID.String(), nil))
		if w4.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w4.Code)
		}
	})

	t.Run("Session, preferences, favorites branches", func(t *testing.T) {
		h := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r := gin.New()
		r.POST("/sessions", func(c *gin.Context) { c.Set("user_id", uint(1)); h.StartSession(c) })
		r.GET("/sessions", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetSessionHistory(c) })
		r.GET("/sessions/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetSessionByID(c) })
		r.GET("/preferences", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetPreferences(c) })
		r.PUT("/preferences", func(c *gin.Context) { c.Set("user_id", uint(1)); h.UpdatePreferences(c) })
		r.GET("/favorites", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetFavorites(c) })
		r.POST("/favorites/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.AddFavorite(c) })
		r.DELETE("/favorites/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.RemoveFavorite(c) })
		r.PUT("/favorites/reorder", func(c *gin.Context) { c.Set("user_id", uint(1)); h.ReorderFavorites(c) })

		techID := uuid.New()
		sessionID := uuid.New()

		badStart := httptest.NewRecorder()
		reqBadStart := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{}`))
		reqBadStart.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(badStart, reqBadStart)
		if badStart.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badStart.Code)
		}

		okStart := httptest.NewRecorder()
		reqStart := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{"technique_id":"`+techID.String()+`","target_duration_seconds":120}`))
		reqStart.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(okStart, reqStart)
		if okStart.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", okStart.Code)
		}

		history := httptest.NewRecorder()
		r.ServeHTTP(history, httptest.NewRequest(http.MethodGet, "/sessions?page=1&limit=10", nil))
		if history.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", history.Code)
		}

		badSessionID := httptest.NewRecorder()
		r.ServeHTTP(badSessionID, httptest.NewRequest(http.MethodGet, "/sessions/bad", nil))
		if badSessionID.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badSessionID.Code)
		}

		okSessionID := httptest.NewRecorder()
		r.ServeHTTP(okSessionID, httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID.String(), nil))
		if okSessionID.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okSessionID.Code)
		}

		prefs := httptest.NewRecorder()
		r.ServeHTTP(prefs, httptest.NewRequest(http.MethodGet, "/preferences", nil))
		if prefs.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", prefs.Code)
		}

		badUpdatePrefs := httptest.NewRecorder()
		reqBadPref := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewBufferString(`{"voice_guidance":"invalid"}`))
		reqBadPref.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(badUpdatePrefs, reqBadPref)
		if badUpdatePrefs.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badUpdatePrefs.Code)
		}

		okUpdatePrefs := httptest.NewRecorder()
		reqPref := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewBufferString(`{"voice_guidance":"ask"}`))
		reqPref.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(okUpdatePrefs, reqPref)
		if okUpdatePrefs.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okUpdatePrefs.Code)
		}

		favs := httptest.NewRecorder()
		r.ServeHTTP(favs, httptest.NewRequest(http.MethodGet, "/favorites", nil))
		if favs.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", favs.Code)
		}

		badAddFav := httptest.NewRecorder()
		r.ServeHTTP(badAddFav, httptest.NewRequest(http.MethodPost, "/favorites/bad", nil))
		if badAddFav.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badAddFav.Code)
		}

		okAddFav := httptest.NewRecorder()
		r.ServeHTTP(okAddFav, httptest.NewRequest(http.MethodPost, "/favorites/"+techID.String(), nil))
		if okAddFav.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okAddFav.Code)
		}

		badRemoveFav := httptest.NewRecorder()
		r.ServeHTTP(badRemoveFav, httptest.NewRequest(http.MethodDelete, "/favorites/bad", nil))
		if badRemoveFav.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badRemoveFav.Code)
		}

		okRemoveFav := httptest.NewRecorder()
		r.ServeHTTP(okRemoveFav, httptest.NewRequest(http.MethodDelete, "/favorites/"+techID.String(), nil))
		if okRemoveFav.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okRemoveFav.Code)
		}

		badReorder := httptest.NewRecorder()
		reqBadReorder := httptest.NewRequest(http.MethodPut, "/favorites/reorder", bytes.NewBufferString(`{}`))
		reqBadReorder.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(badReorder, reqBadReorder)
		if badReorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badReorder.Code)
		}

		okReorder := httptest.NewRecorder()
		reqReorder := httptest.NewRequest(http.MethodPut, "/favorites/reorder", bytes.NewBufferString(`[`+`"`+techID.String()+`"`+`]`))
		reqReorder.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(okReorder, reqReorder)
		if okReorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okReorder.Code)
		}
	})

	t.Run("Stats, calendar, usage, widget, recommendations", func(t *testing.T) {
		h := NewBreathingHandler(&mockBreathingServiceForHandler{})
		r := gin.New()
		r.GET("/stats", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetStats(c) })
		r.GET("/calendar", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetCalendar(c) })
		r.GET("/usage", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetTechniqueUsage(c) })
		r.GET("/widget", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetWidgetData(c) })
		r.GET("/recommendations", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetRecommendations(c) })

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/stats", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/calendar?year=x&month=2", nil))
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/calendar?year=2026&month=13", nil))
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/calendar?year=2026&month=2", nil))
		if w4.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w4.Code)
		}

		w5 := httptest.NewRecorder()
		r.ServeHTTP(w5, httptest.NewRequest(http.MethodGet, "/usage", nil))
		if w5.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w5.Code)
		}

		w6 := httptest.NewRecorder()
		r.ServeHTTP(w6, httptest.NewRequest(http.MethodGet, "/widget", nil))
		if w6.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w6.Code)
		}

		w7 := httptest.NewRecorder()
		r.ServeHTTP(w7, httptest.NewRequest(http.MethodGet, "/recommendations?mood=stressed&time_of_day=night", nil))
		if w7.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w7.Code)
		}
	})
}

func TestBreathingHandler_RemainingEndpointErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	techID := uuid.New()

	h := NewBreathingHandler(&mockBreathingServiceForHandler{
		getSessionByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*dto.BreathingSessionResponse, error) {
			return nil, errors.New("missing session")
		},
		getPreferencesFn: func(_ context.Context, _ uint) (*dto.BreathingPreferencesResponse, error) {
			return nil, errors.New("prefs fail")
		},
		updatePreferencesFn: func(_ context.Context, _ uint, _ dto.UpdateBreathingPreferencesRequest) (*dto.BreathingPreferencesResponse, error) {
			return nil, errors.New("update prefs fail")
		},
		getFavoritesFn: func(_ context.Context, _ uint) ([]dto.BreathingTechniqueResponse, error) {
			return nil, errors.New("favorites fail")
		},
		addFavoriteFn: func(_ context.Context, _ uint, _ uuid.UUID) error {
			return errors.New("add favorite fail")
		},
		removeFavoriteFn: func(_ context.Context, _ uint, _ uuid.UUID) error {
			return errors.New("remove favorite fail")
		},
		reorderFavoritesFn: func(_ context.Context, _ uint, _ []uuid.UUID) error {
			return errors.New("reorder fail")
		},
		getStatsFn: func(_ context.Context, _ uint) (*dto.BreathingStatsResponse, error) {
			return nil, errors.New("stats fail")
		},
		getCalendarFn: func(_ context.Context, _ uint, _, _ int) (*dto.BreathingCalendarResponse, error) {
			return nil, errors.New("calendar fail")
		},
		getTechniqueUsageFn: func(_ context.Context, _ uint) ([]dto.TechniqueUsageStats, error) {
			return nil, errors.New("usage fail")
		},
		getWidgetDataFn: func(_ context.Context, _ uint) (*dto.BreathingWidgetData, error) {
			return nil, errors.New("widget fail")
		},
		getRecommendationsFn: func(_ context.Context, _ uint, _, _ string) (*dto.RecommendationsResponse, error) {
			return nil, errors.New("recommendation fail")
		},
	})

	r := gin.New()
	r.GET("/sessions/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetSessionByID(c) })
	r.GET("/preferences", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetPreferences(c) })
	r.PUT("/preferences", func(c *gin.Context) { c.Set("user_id", uint(1)); h.UpdatePreferences(c) })
	r.GET("/favorites", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetFavorites(c) })
	r.POST("/favorites/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.AddFavorite(c) })
	r.DELETE("/favorites/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.RemoveFavorite(c) })
	r.PUT("/favorites/reorder", func(c *gin.Context) { c.Set("user_id", uint(1)); h.ReorderFavorites(c) })
	r.GET("/stats", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetStats(c) })
	r.GET("/calendar", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetCalendar(c) })
	r.GET("/usage", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetTechniqueUsage(c) })
	r.GET("/widget", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetWidgetData(c) })
	r.GET("/recommendations", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetRecommendations(c) })

	t.Run("GetSessionByID not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions/"+uuid.New().String(), nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("GetPreferences and UpdatePreferences service errors", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/preferences", nil))
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewBufferString(`{"voice_guidance":"ask"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w2, req)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}
	})

	t.Run("Favorites service errors", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/favorites", nil))
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/favorites/"+techID.String(), nil))
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, "/favorites/"+techID.String(), nil))
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		w4 := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/favorites/reorder", bytes.NewBufferString(`[`+`"`+techID.String()+`"`+`]`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w4, req)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w4.Code)
		}
	})

	t.Run("Stats and widgets service errors", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/stats", nil))
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/calendar?year=2026&month=2", nil))
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/usage", nil))
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/widget", nil))
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w4.Code)
		}

		w5 := httptest.NewRecorder()
		r.ServeHTTP(w5, httptest.NewRequest(http.MethodGet, "/recommendations?mood=tired&time_of_day=morning", nil))
		if w5.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w5.Code)
		}
	})
}

func TestBreathingHandler_SessionAndTechniqueErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	techID := uuid.New()
	sessionID := uuid.New()

	h := NewBreathingHandler(&mockBreathingServiceForHandler{
		updateTechniqueFn: func(_ context.Context, _ uint, _ uuid.UUID, _ dto.UpdateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
			return nil, errors.New("update technique fail")
		},
		deleteTechniqueFn: func(_ context.Context, _ uint, _ uuid.UUID) error {
			return errors.New("delete technique fail")
		},
		startSessionFn: func(_ context.Context, _ uint, _ dto.StartBreathingSessionRequest) (*dto.BreathingSessionResponse, error) {
			return nil, errors.New("start session fail")
		},
		completeSessionFn: func(_ context.Context, _ uint, _ uuid.UUID, _ dto.CompleteBreathingSessionRequest) (*dto.SessionCompletionResult, error) {
			return nil, errors.New("complete session fail")
		},
		getSessionHistoryFn: func(_ context.Context, _ uint, _ dto.SessionHistoryRequest) (*dto.SessionHistoryResponse, error) {
			return nil, errors.New("history fail")
		},
	})

	r := gin.New()
	r.PUT("/techniques/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.UpdateTechnique(c) })
	r.DELETE("/techniques/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); h.DeleteTechnique(c) })
	r.POST("/sessions", func(c *gin.Context) { c.Set("user_id", uint(1)); h.StartSession(c) })
	r.POST("/sessions/:id/complete", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CompleteSession(c) })
	r.GET("/sessions", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetSessionHistory(c) })

	t.Run("UpdateTechnique invalid payload then service error", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPut, "/techniques/"+techID.String(), bytes.NewBufferString(`{"name":123}`))
		req1.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPut, "/techniques/"+techID.String(), bytes.NewBufferString(`{"name":"Updated"}`))
		req2.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}
	})

	t.Run("DeleteTechnique service error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/techniques/"+techID.String(), nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("StartSession service error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{"technique_id":"`+techID.String()+`","target_duration_seconds":120}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("CompleteSession invalid payload then service error", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/complete", bytes.NewBufferString(`{"duration_seconds":"bad"}`))
		req1.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/complete", bytes.NewBufferString(`{"duration_seconds":120}`))
		req2.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}
	})

	t.Run("GetSessionHistory service error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions?page=1&limit=10", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}
