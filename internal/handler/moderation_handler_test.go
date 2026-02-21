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

func setupModerationHandlerWithService(t *testing.T) *ModerationHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
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

func setupModerationHandlerRich(t *testing.T) *ModerationHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.ArticleCategory{},
		&model.Article{},
		&model.UserReport{},
		&model.UserBlock{},
		&model.UserStrike{},
		&model.ModeratorAction{},
		&model.CrisisKeyword{},
		&model.ContentFlag{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cat := model.ArticleCategory{Name: "General", Description: "d"}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}

	moderator := model.User{Name: "Moderator", Username: "moderator1", Email: "moderator@test.local", Password: "x", Role: model.RoleModerator}
	reporter := model.User{Name: "Reporter", Username: "reporter1", Email: "reporter@test.local", Password: "x", Role: model.RoleMember}
	reported := model.User{Name: "Reported", Username: "reported1", Email: "reported@test.local", Password: "x", Role: model.RoleMember}
	if err := db.Create(&moderator).Error; err != nil {
		t.Fatalf("seed moderator: %v", err)
	}
	if err := db.Create(&reporter).Error; err != nil {
		t.Fatalf("seed reporter: %v", err)
	}
	if err := db.Create(&reported).Error; err != nil {
		t.Fatalf("seed reported: %v", err)
	}

	article := model.Article{Title: "Need moderation", Content: "content", UserID: reported.ID, ArticleCategoryID: cat.ID, ModerationStatus: model.ArticleModerationPending}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("seed article: %v", err)
	}

	report := model.UserReport{ReporterID: reporter.ID, ReportType: model.ReportTypeArticle, ReportedContentID: &article.ID, Reason: model.ReportReasonSpam, Description: "spam", Status: model.ReportStatusPending}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("seed report: %v", err)
	}

	strike := model.UserStrike{UserID: reported.ID, Reason: "old strike", IsActive: true}
	if err := db.Create(&strike).Error; err != nil {
		t.Fatalf("seed strike: %v", err)
	}

	keyword := model.CrisisKeyword{Keyword: "hopeless", Severity: "high", IsActive: true}
	if err := db.Create(&keyword).Error; err != nil {
		t.Fatalf("seed keyword: %v", err)
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

func newModerationTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestModerationHandler_InvalidInputBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewModerationHandler(nil)

	tests := []struct {
		name   string
		call   func(*gin.Context)
		method string
		target string
		body   string
		setup  func(*gin.Context)
		code   int
	}{
		{name: "get-moderation-queue-invalid-query", call: h.GetModerationQueue, method: http.MethodGet, target: "/moderation/queue?page=abc", code: http.StatusBadRequest},
		{name: "moderate-article-invalid-id", call: h.ModerateArticle, method: http.MethodPut, target: "/moderation/articles/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "moderate-article-invalid-json", call: h.ModerateArticle, method: http.MethodPut, target: "/moderation/articles/1", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "1"}} }, code: http.StatusBadRequest},
		{name: "get-reports-invalid-query", call: h.GetReports, method: http.MethodGet, target: "/moderation/reports?page=abc", code: http.StatusBadRequest},
		{name: "handle-report-invalid-id", call: h.HandleReport, method: http.MethodPut, target: "/moderation/reports/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "handle-report-invalid-json", call: h.HandleReport, method: http.MethodPut, target: "/moderation/reports/1", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "1"}} }, code: http.StatusBadRequest},
		{name: "create-report-invalid-json", call: h.CreateReport, method: http.MethodPost, target: "/reports", body: "{", code: http.StatusBadRequest},
		{name: "block-user-invalid-json", call: h.BlockUser, method: http.MethodPost, target: "/blocks", body: "{", code: http.StatusBadRequest},
		{name: "unblock-user-invalid-id", call: h.UnblockUser, method: http.MethodDelete, target: "/blocks/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "get-user-strikes-invalid-id", call: h.GetUserStrikes, method: http.MethodGet, target: "/moderation/users/bad/strikes", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "add-trigger-warnings-invalid-json", call: h.AddTriggerWarnings, method: http.MethodPost, target: "/moderation/trigger-warnings", body: "{", code: http.StatusBadRequest},
		{name: "update-content-warning-preference-invalid-json", call: h.UpdateContentWarningPreference, method: http.MethodPut, target: "/user/content-warning-preference", body: "{", code: http.StatusBadRequest},
		{name: "get-moderator-actions-invalid-query", call: h.GetModeratorActions, method: http.MethodGet, target: "/moderation/actions?page=abc", code: http.StatusBadRequest},
		{name: "create-crisis-keyword-invalid-json", call: h.CreateCrisisKeyword, method: http.MethodPost, target: "/moderation/crisis-keywords", body: "{", code: http.StatusBadRequest},
		{name: "delete-crisis-keyword-invalid-id", call: h.DeleteCrisisKeyword, method: http.MethodDelete, target: "/moderation/crisis-keywords/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newModerationTestContext(tc.method, tc.target, tc.body)
			if tc.setup != nil {
				tc.setup(c)
			}
			tc.call(c)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}

func TestModerationHandler_ServiceBackedBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupModerationHandlerWithService(t)

	t.Run("get moderation stats", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodGet, "/moderation/stats", "")
		h.GetModerationStats(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get blocked users service error", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodGet, "/blocks", "")
		c.Set("user_id", uint(1))
		h.GetBlockedUsers(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("accept ai disclaimer service error", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodPost, "/user/accept-ai-disclaimer", "")
		c.Set("user_id", uint(1))
		h.AcceptAIDisclaimer(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("get crisis keywords service error", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodGet, "/moderation/crisis-keywords", "")
		h.GetCrisisKeywords(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestModerationHandler_RichSuccessAndErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupModerationHandlerRich(t)

	t.Run("queue-and-reports-success", func(t *testing.T) {
		cq, wq := newModerationTestContext(http.MethodGet, "/moderation/queue?page=1&limit=10", "")
		h.GetModerationQueue(cq)
		if wq.Code != http.StatusOK {
			t.Fatalf("expected 200 queue, got %d", wq.Code)
		}

		cr, wr := newModerationTestContext(http.MethodGet, "/moderation/reports?page=1&limit=10", "")
		h.GetReports(cr)
		if wr.Code != http.StatusOK {
			t.Fatalf("expected 200 reports, got %d", wr.Code)
		}
	})

	t.Run("moderate-article-success-and-invalid-action", func(t *testing.T) {
		c1, w1 := newModerationTestContext(http.MethodPut, "/moderation/articles/1", `{"action":"approve","notes":"ok"}`)
		c1.Set("user_id", uint(1))
		c1.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ModerateArticle(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 moderate approve, got %d", w1.Code)
		}

		c2, w2 := newModerationTestContext(http.MethodPut, "/moderation/articles/1", `{"action":"bad_action","notes":"x"}`)
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ModerateArticle(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid action path, got %d", w2.Code)
		}
	})

	t.Run("create-and-handle-report", func(t *testing.T) {
		c1, w1 := newModerationTestContext(http.MethodPost, "/reports", `{"report_type":"user","reason":"harassment","user_id":3,"description":"toxic"}`)
		c1.Set("user_id", uint(2))
		h.CreateReport(c1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("expected 201 create report, got %d", w1.Code)
		}

		c2, w2 := newModerationTestContext(http.MethodPut, "/moderation/reports/1", `{"action":"dismiss","notes":"checked"}`)
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "id", Value: "1"}}
		h.HandleReport(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 handle report, got %d", w2.Code)
		}
	})

	t.Run("block-unblock-and-list-success", func(t *testing.T) {
		cb, wb := newModerationTestContext(http.MethodPost, "/blocks", `{"user_id":3,"reason":"mute"}`)
		cb.Set("user_id", uint(2))
		h.BlockUser(cb)
		if wb.Code != http.StatusCreated {
			t.Fatalf("expected 201 block user, got %d", wb.Code)
		}

		cget, wget := newModerationTestContext(http.MethodGet, "/blocks", "")
		cget.Set("user_id", uint(2))
		h.GetBlockedUsers(cget)
		if wget.Code != http.StatusOK {
			t.Fatalf("expected 200 get blocked users, got %d", wget.Code)
		}

		cu, wu := newModerationTestContext(http.MethodDelete, "/blocks/3", "")
		cu.Set("user_id", uint(2))
		cu.Params = gin.Params{{Key: "id", Value: "3"}}
		h.UnblockUser(cu)
		if wu.Code != http.StatusOK {
			t.Fatalf("expected 200 unblock user, got %d", wu.Code)
		}
	})

	t.Run("strikes-actions-keywords-and-preferences", func(t *testing.T) {
		cs, ws := newModerationTestContext(http.MethodGet, "/moderation/users/3/strikes?active_only=true", "")
		cs.Params = gin.Params{{Key: "id", Value: "3"}}
		h.GetUserStrikes(cs)
		if ws.Code != http.StatusOK {
			t.Fatalf("expected 200 get strikes, got %d", ws.Code)
		}

		ca, wa := newModerationTestContext(http.MethodGet, "/moderation/actions?page=1&limit=10", "")
		h.GetModeratorActions(ca)
		if wa.Code != http.StatusOK {
			t.Fatalf("expected 200 get actions, got %d", wa.Code)
		}

		ctw, wtw := newModerationTestContext(http.MethodPost, "/moderation/trigger-warnings", `{"content_type":"article","content_id":1,"trigger_warnings":["violence"]}`)
		ctw.Set("user_id", uint(1))
		h.AddTriggerWarnings(ctw)
		if wtw.Code != http.StatusOK {
			t.Fatalf("expected 200 add trigger warnings, got %d", wtw.Code)
		}

		cad, wad := newModerationTestContext(http.MethodPost, "/user/accept-ai-disclaimer", "")
		cad.Set("user_id", uint(2))
		h.AcceptAIDisclaimer(cad)
		if wad.Code != http.StatusOK {
			t.Fatalf("expected 200 accept disclaimer, got %d", wad.Code)
		}

		cup, wup := newModerationTestContext(http.MethodPut, "/user/content-warning-preference", `{"preference":"show"}`)
		cup.Set("user_id", uint(2))
		h.UpdateContentWarningPreference(cup)
		if wup.Code != http.StatusOK {
			t.Fatalf("expected 200 update preference path, got %d", wup.Code)
		}

		cgk, wgk := newModerationTestContext(http.MethodGet, "/moderation/crisis-keywords", "")
		h.GetCrisisKeywords(cgk)
		if wgk.Code != http.StatusOK {
			t.Fatalf("expected 200 get crisis keywords, got %d", wgk.Code)
		}

		cck, wck := newModerationTestContext(http.MethodPost, "/moderation/crisis-keywords", `{"keyword":"self harm","category":"self_harm","severity":"critical"}`)
		h.CreateCrisisKeyword(cck)
		if wck.Code != http.StatusCreated {
			t.Fatalf("expected 201 create crisis keyword, got %d", wck.Code)
		}

		dck, wdk := newModerationTestContext(http.MethodDelete, "/moderation/crisis-keywords/1", "")
		dck.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteCrisisKeyword(dck)
		if wdk.Code != http.StatusOK {
			t.Fatalf("expected 200 delete crisis keyword, got %d", wdk.Code)
		}
	})
}

func TestModerationHandler_InternalErrorBranches_Misc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupModerationHandlerWithService(t)

	t.Run("update-content-warning-preference-internal", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodPut, "/user/content-warning-preference", `{"preference":"show"}`)
		c.Set("user_id", uint(1))
		h.UpdateContentWarningPreference(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("create-crisis-keyword-internal", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodPost, "/moderation/crisis-keywords", `{"keyword":"urgent","category":"self_harm","severity":"high"}`)
		h.CreateCrisisKeyword(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("delete-crisis-keyword-internal", func(t *testing.T) {
		c, w := newModerationTestContext(http.MethodDelete, "/moderation/crisis-keywords/1", "")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteCrisisKeyword(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
