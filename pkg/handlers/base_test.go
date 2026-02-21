package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/pkg/apperror"
	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/gin-gonic/gin"
)

type bindReq struct {
	Name string `json:"name" form:"name" binding:"required"`
}

func newTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req
	return c, w
}

func TestBaseHandler_ResponseMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBaseHandler()

	tests := []struct {
		name     string
		call     func(*gin.Context)
		expected int
	}{
		{"ok", func(c *gin.Context) { h.OK(c, gin.H{"x": 1}, "") }, http.StatusOK},
		{"created", func(c *gin.Context) { h.Created(c, gin.H{"x": 1}, "") }, http.StatusCreated},
		{"no-content", h.NoContent, http.StatusNoContent},
		{"deleted", func(c *gin.Context) { h.Deleted(c, "removed") }, http.StatusOK},
		{"updated", func(c *gin.Context) { h.Updated(c, gin.H{"x": 1}, "") }, http.StatusOK},
		{"paginated", func(c *gin.Context) { h.Paginated(c, []int{1, 2}, 1, 10, 2) }, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, "/", "")
			tt.call(c)
			status := c.Writer.Status()
			if status != tt.expected {
				t.Fatalf("expected status %d, got %d (recorder=%d)", tt.expected, status, w.Code)
			}
		})
	}
}

func TestBaseHandler_ErrorMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBaseHandler()

	tests := []struct {
		name     string
		call     func(*gin.Context)
		expected int
	}{
		{"error", func(c *gin.Context) { h.Error(c, apperror.BadRequest("bad")) }, http.StatusBadRequest},
		{"bad-request", func(c *gin.Context) { h.BadRequest(c, "bad") }, http.StatusBadRequest},
		{"unauthorized", func(c *gin.Context) { h.Unauthorized(c, "unauth") }, http.StatusUnauthorized},
		{"forbidden", func(c *gin.Context) { h.Forbidden(c, "forbidden") }, http.StatusForbidden},
		{"not-found", func(c *gin.Context) { h.NotFound(c, "item") }, http.StatusNotFound},
		{"internal", func(c *gin.Context) { h.InternalError(c, "internal") }, http.StatusInternalServerError},
		{"validation", func(c *gin.Context) {
			h.ValidationError(c, []apperror.FieldError{{Field: "name", Message: "required"}})
		}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, "/", "")
			tt.call(c)
			if w.Code != tt.expected {
				t.Fatalf("expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

func TestBaseHandler_ContextAndQueryHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBaseHandler()

	c, _ := newTestContext(http.MethodGet, "/?page=2&limit=5&active=true&n=12&uid=8&name=abc", "")
	ctxutil.SetUserInfo(c, 42, "u@u", "admin")

	uid, ok := h.GetUserID(c)
	if !ok || uid != 42 {
		t.Fatalf("expected user id 42, got %d (ok=%v)", uid, ok)
	}
	if h.MustGetUserID(c) != 42 {
		t.Fatalf("expected must user id 42")
	}
	if _, ok := h.GetUserInfo(c); !ok {
		t.Fatalf("expected user info")
	}
	if !h.IsAdmin(c) || !h.IsModerator(c) {
		t.Fatalf("expected admin/moderator to be true")
	}
	if !h.IsOwnerOrAdmin(c, 999) {
		t.Fatalf("expected owner/admin check true for admin")
	}

	p := h.GetPagination(c)
	if p.Page != 2 || p.Limit != 5 {
		t.Fatalf("unexpected pagination: %+v", p)
	}
	if h.GetIntQuery(c, "n", 0) != 12 {
		t.Fatalf("expected int query 12")
	}
	if h.GetUintQuery(c, "uid", 0) != 8 {
		t.Fatalf("expected uint query 8")
	}
	if !h.GetBoolQuery(c, "active", false) {
		t.Fatalf("expected bool query true")
	}
	if h.GetStringQuery(c, "name", "") != "abc" {
		t.Fatalf("expected name abc")
	}
	if v := h.GetOptionalUint(c, "uid"); v == nil || *v != 8 {
		t.Fatalf("expected optional uint 8")
	}
}

func TestBaseHandler_ParamBindingAndAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBaseHandler()

	{
		c, _ := newTestContext(http.MethodGet, "/items/7", "")
		c.Params = gin.Params{{Key: "id", Value: "7"}}
		if id, ok := h.GetUintParam(c, "id"); !ok || id != 7 {
			t.Fatalf("expected uint param 7")
		}
		if h.MustGetUintParam(c, "id") != 7 {
			t.Fatalf("expected must uint param 7")
		}
		if id, ok := h.GetIntParam(c, "id"); !ok || id != 7 {
			t.Fatalf("expected int param 7")
		}
	}

	{
		c, _ := newTestContext(http.MethodPost, "/", `{"name":"alf"}`)
		var req bindReq
		if !h.BindJSON(c, &req) || req.Name != "alf" {
			t.Fatalf("expected bind json success")
		}
	}

	{
		c, w := newTestContext(http.MethodPost, "/", `{}`)
		var req bindReq
		if h.BindJSON(c, &req) {
			t.Fatalf("expected bind json fail")
		}
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, got %d", w.Code)
		}
	}

	{
		c, _ := newTestContext(http.MethodGet, "/?name=ok", "")
		var req bindReq
		if !h.BindQuery(c, &req) || req.Name != "ok" {
			t.Fatalf("expected bind query success")
		}
	}

	{
		c, w := newTestContext(http.MethodGet, "/", "")
		var req bindReq
		if h.BindQuery(c, &req) {
			t.Fatalf("expected bind query fail")
		}
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, got %d", w.Code)
		}
	}

	{
		c, _ := newTestContext(http.MethodGet, "/", "")
		ctxutil.SetUserInfo(c, 42, "u@u", "member")
		if !h.RequireOwnerOrAdmin(c, 42) {
			t.Fatalf("expected owner authorized")
		}
	}

	{
		c, w := newTestContext(http.MethodGet, "/", "")
		ctxutil.SetUserInfo(c, 42, "u@u", "member")
		if h.RequireOwnerOrAdmin(c, 99) {
			t.Fatalf("expected owner/admin unauthorized")
		}
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected forbidden, got %d", w.Code)
		}
	}

	{
		c, _ := newTestContext(http.MethodGet, "/", "")
		ctxutil.SetUserInfo(c, 1, "a@a", "admin")
		if !h.RequireAdmin(c) || !h.RequireModerator(c) {
			t.Fatalf("expected admin and moderator pass")
		}
	}

	{
		c, w := newTestContext(http.MethodGet, "/", "")
		ctxutil.SetUserInfo(c, 1, "m@m", "member")
		if h.RequireAdmin(c) || h.RequireModerator(c) {
			t.Fatalf("expected require checks fail")
		}
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected forbidden, got %d", w.Code)
		}
	}
}

func TestBaseHandler_HandleServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBaseHandler()

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"nil", nil, 200},
		{"not-found", errors.New("not found"), http.StatusNotFound},
		{"record-not-found", errors.New("record not found"), http.StatusNotFound},
		{"unauthorized", errors.New("unauthorized"), http.StatusForbidden},
		{"forbidden", errors.New("forbidden"), http.StatusForbidden},
		{"app-error", apperror.BadRequest("bad"), http.StatusBadRequest},
		{"unknown", errors.New("something broke"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, "/", "")
			if tt.err == nil {
				c.Status(200)
			}
			h.HandleServiceError(c, tt.err, "forum")
			if w.Code != tt.expected {
				t.Fatalf("expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}
