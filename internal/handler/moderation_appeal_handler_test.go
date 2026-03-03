package handler

import (
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

func newModerationAppealTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestModerationAppealHandler_InvalidBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ModerationHandler{}

	{
		c, w := newModerationAppealTestContext(http.MethodPost, "/appeals", "{")
		h.CreateAppeal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 CreateAppeal, got %d", w.Code)
		}
	}

	{
		c, w := newModerationAppealTestContext(http.MethodPut, "/moderation/appeals/bad", "{}")
		c.Params = gin.Params{{Key: "id", Value: "bad"}}
		h.ReviewAppeal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 ReviewAppeal invalid id, got %d", w.Code)
		}
	}

	{
		c, w := newModerationAppealTestContext(http.MethodPut, "/moderation/appeals/1", "{")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ReviewAppeal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 ReviewAppeal invalid body, got %d", w.Code)
		}
	}

	{
		h2 := setupModerationHandlerWithService(t)
		c, w := newModerationAppealTestContext(http.MethodGet, "/moderation/appeals?page=1&limit=20", "")
		h2.GetAppeals(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 GetAppeals service error, got %d", w.Code)
		}
	}
}

func setupModerationAppealHandlerDB(t *testing.T) *ModerationHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Appeal{}, &model.ModeratorAction{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	users := []model.User{
		{Name: "Member", Username: "member_appeal", Email: "member.appeal@test.local", Password: "x", Role: model.RoleMember},
		{Name: "Moderator", Username: "moderator_appeal", Email: "moderator.appeal@test.local", Password: "x", Role: model.RoleModerator},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	svc := service.NewModerationService(
		repository.NewModerationRepository(db),
		repository.NewUserRepository(db),
		repository.NewArticleRepository(db),
		repository.NewForumRepository(db),
		nil,
	)
	return NewModerationHandler(svc)
}

func TestModerationAppealHandler_SuccessAndServiceBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupModerationAppealHandlerDB(t)

	{
		c, w := newModerationAppealTestContext(http.MethodPost, "/appeals", `{"reason":"please reconsider","evidence":"new detail"}`)
		c.Set("user_id", uint(1))
		h.CreateAppeal(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 CreateAppeal success, got %d", w.Code)
		}
	}

	{
		c, w := newModerationAppealTestContext(http.MethodPost, "/appeals", `{"reason":"second attempt","evidence":"follow up"}`)
		c.Set("user_id", uint(1))
		h.CreateAppeal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 CreateAppeal duplicate active appeal, got %d", w.Code)
		}
	}

	{
		c, w := newModerationAppealTestContext(http.MethodGet, "/moderation/appeals?status=pending&page=1&limit=20", "")
		h.GetAppeals(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetAppeals success, got %d", w.Code)
		}
	}

	{
		c, w := newModerationAppealTestContext(http.MethodPut, "/moderation/appeals/1", `{"status":"approved","notes":"ok"}`)
		c.Set("user_id", uint(2))
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ReviewAppeal(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ReviewAppeal success, got %d", w.Code)
		}
	}

	{
		c, w := newModerationAppealTestContext(http.MethodPut, "/moderation/appeals/1", `{"status":"wrong","notes":"x"}`)
		c.Set("user_id", uint(2))
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ReviewAppeal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 ReviewAppeal invalid status, got %d", w.Code)
		}
	}
}
