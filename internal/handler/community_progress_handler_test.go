package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCommunityCtx(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, w
}

func TestCommunityProgressHandler_UnauthorizedGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CommunityProgressHandler{}

	tests := []struct {
		name string
		call func(*gin.Context)
	}{
		{"GetPersonalJourney", h.GetPersonalJourney},
		{"GetWeeklyProgress", h.GetWeeklyProgress},
		{"GetMonthlyProgress", h.GetMonthlyProgress},
		{"GetAllTimeStats", h.GetAllTimeStats},
		{"GetLevelUpCelebration", h.GetLevelUpCelebration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newCommunityCtx(http.MethodGet, "/")
			c.Params = gin.Params{{Key: "level", Value: "2"}}
			tt.call(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestCommunityProgressHandler_PublicValidationBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CommunityProgressHandler{}

	// Invalid level path should fail before touching service
	{
		c, w := newCommunityCtx(http.MethodGet, "/community/hall-of-fame/level/99")
		c.Params = gin.Params{{Key: "level", Value: "99"}}
		h.GetLevelHallOfFame(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid level hall-of-fame input, got %d", w.Code)
		}
	}

	// Invalid month/year query should fail before touching service
	{
		c, w := newCommunityCtx(http.MethodGet, "/community/hall-of-fame/monthly?month=13&year=2026")
		h.GetMonthlyHallOfFame(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid month, got %d", w.Code)
		}
	}
	{
		c, w := newCommunityCtx(http.MethodGet, "/community/hall-of-fame/monthly?month=1&year=1800")
		h.GetMonthlyHallOfFame(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid year, got %d", w.Code)
		}
	}
}
