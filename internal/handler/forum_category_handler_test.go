package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

type mockForumCategorySvcForHandler struct {
	createErr error
	listErr   error
	updateErr error
	deleteErr error

	lastCreateName string
	lastUpdateID   uint
	lastUpdateName string
	lastDeleteID   uint
}

func (m *mockForumCategorySvcForHandler) CreateCategory(_ context.Context, name string) error {
	m.lastCreateName = name
	return m.createErr
}

func (m *mockForumCategorySvcForHandler) GetAllCategories(_ context.Context) ([]model.ForumCategory, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return []model.ForumCategory{{ID: 1, Name: "General"}}, nil
}

func (m *mockForumCategorySvcForHandler) UpdateCategory(_ context.Context, id uint, name string) error {
	m.lastUpdateID = id
	m.lastUpdateName = name
	return m.updateErr
}

func (m *mockForumCategorySvcForHandler) DeleteCategory(_ context.Context, id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func TestForumCategoryHandler_CreateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockForumCategorySvcForHandler{}
	h := NewForumCategoryHandler(svc)

	r := gin.New()
	r.POST("/categories", h.CreateCategory)

	body := bytes.NewBufferString(`{"name":"Mindfulness"}`)
	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if svc.lastCreateName != "Mindfulness" {
		t.Fatalf("expected create name to be propagated, got %s", svc.lastCreateName)
	}
}

func TestForumCategoryHandler_GetAllCategories_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockForumCategorySvcForHandler{listErr: errors.New("db down")}
	h := NewForumCategoryHandler(svc)

	r := gin.New()
	r.GET("/categories", h.GetAllCategories)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestForumCategoryHandler_GetAllCategories_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockForumCategorySvcForHandler{}
	h := NewForumCategoryHandler(svc)

	r := gin.New()
	r.GET("/categories", h.GetAllCategories)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestForumCategoryHandler_UpdateAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockForumCategorySvcForHandler{}
	h := NewForumCategoryHandler(svc)

	r := gin.New()
	r.PUT("/categories/:id", h.UpdateCategory)
	r.DELETE("/categories/:id", h.DeleteCategory)

	updateBody := bytes.NewBufferString(`{"name":"Updated"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/categories/7", updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()
	r.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("expected update 200, got %d", updateW.Code)
	}
	if svc.lastUpdateID != 7 || svc.lastUpdateName != "Updated" {
		t.Fatalf("unexpected update inputs: id=%d name=%s", svc.lastUpdateID, svc.lastUpdateName)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/categories/7", nil)
	deleteW := httptest.NewRecorder()
	r.ServeHTTP(deleteW, deleteReq)

	if deleteW.Code != http.StatusOK {
		t.Fatalf("expected delete 200, got %d", deleteW.Code)
	}
	if svc.lastDeleteID != 7 {
		t.Fatalf("unexpected delete id: %d", svc.lastDeleteID)
	}

	var payload map[string]any
	if err := json.Unmarshal(deleteW.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json payload: %v", err)
	}
}

func TestForumCategoryHandler_AdminAndValidationBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("admin get all categories success", func(t *testing.T) {
		svc := &mockForumCategorySvcForHandler{}
		h := NewForumCategoryHandler(svc)

		r := gin.New()
		r.GET("/admin/categories", h.AdminGetAllCategories)

		req := httptest.NewRequest(http.MethodGet, "/admin/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("admin get all categories service error", func(t *testing.T) {
		svc := &mockForumCategorySvcForHandler{listErr: errors.New("db down")}
		h := NewForumCategoryHandler(svc)

		r := gin.New()
		r.GET("/admin/categories", h.AdminGetAllCategories)

		req := httptest.NewRequest(http.MethodGet, "/admin/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("create and update invalid payload", func(t *testing.T) {
		svc := &mockForumCategorySvcForHandler{}
		h := NewForumCategoryHandler(svc)

		r := gin.New()
		r.POST("/categories", h.CreateCategory)
		r.PUT("/categories/:id", h.UpdateCategory)

		reqCreate := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString("{"))
		reqCreate.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		r.ServeHTTP(wCreate, reqCreate)
		if wCreate.Code != http.StatusBadRequest {
			t.Fatalf("expected create 400, got %d", wCreate.Code)
		}

		reqUpdate := httptest.NewRequest(http.MethodPut, "/categories/1", bytes.NewBufferString("{"))
		reqUpdate.Header.Set("Content-Type", "application/json")
		wUpdate := httptest.NewRecorder()
		r.ServeHTTP(wUpdate, reqUpdate)
		if wUpdate.Code != http.StatusBadRequest {
			t.Fatalf("expected update 400, got %d", wUpdate.Code)
		}
	})

	t.Run("create service error", func(t *testing.T) {
		svc := &mockForumCategorySvcForHandler{createErr: errors.New("boom")}
		h := NewForumCategoryHandler(svc)

		r := gin.New()
		r.POST("/categories", h.CreateCategory)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":"X"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("update and delete service error", func(t *testing.T) {
		svc := &mockForumCategorySvcForHandler{updateErr: errors.New("boom"), deleteErr: errors.New("boom")}
		h := NewForumCategoryHandler(svc)

		r := gin.New()
		r.PUT("/categories/:id", h.UpdateCategory)
		r.DELETE("/categories/:id", h.DeleteCategory)

		reqUpdate := httptest.NewRequest(http.MethodPut, "/categories/3", bytes.NewBufferString(`{"name":"X"}`))
		reqUpdate.Header.Set("Content-Type", "application/json")
		wUpdate := httptest.NewRecorder()
		r.ServeHTTP(wUpdate, reqUpdate)
		if wUpdate.Code != http.StatusInternalServerError {
			t.Fatalf("expected update 500, got %d", wUpdate.Code)
		}

		reqDelete := httptest.NewRequest(http.MethodDelete, "/categories/3", nil)
		wDelete := httptest.NewRecorder()
		r.ServeHTTP(wDelete, reqDelete)
		if wDelete.Code != http.StatusInternalServerError {
			t.Fatalf("expected delete 500, got %d", wDelete.Code)
		}
	})
}
