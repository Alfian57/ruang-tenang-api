package handler

import (
	"bytes"
	"fmt"
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

func newArticleHandlerTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, w
}

func TestArticleHandler_ProtectedEndpointsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ArticleHandler{}

	tests := []struct {
		name   string
		call   func(*gin.Context)
		method string
		target string
		setup  func(*gin.Context)
	}{
		{name: "get-my-articles", call: h.GetMyArticles, method: http.MethodGet, target: "/my-articles"},
		{name: "create-my-article", call: h.CreateMyArticle, method: http.MethodPost, target: "/my-articles"},
		{name: "update-my-article", call: h.UpdateMyArticle, method: http.MethodPut, target: "/my-articles/slug", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "slug", Value: "slug"}} }},
		{name: "delete-my-article", call: h.DeleteMyArticle, method: http.MethodDelete, target: "/my-articles/slug", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "slug", Value: "slug"}} }},
		{name: "get-article-by-id-for-user", call: h.GetArticleByIDForUser, method: http.MethodGet, target: "/my-articles/slug", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "slug", Value: "slug"}} }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newArticleHandlerTestContext(tc.method, tc.target)
			if tc.setup != nil {
				tc.setup(c)
			}
			tc.call(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
		})
	}
}

func newArticleHandlerWithDB(t *testing.T) (*ArticleHandler, *gorm.DB, uint, uint, uint, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ArticleCategory{}, &model.Article{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	owner := model.User{Name: "Owner", Username: "owner_a", Email: "owner_a@test.local", Password: "x"}
	other := model.User{Name: "Other", Username: "other_a", Email: "other_a@test.local", Password: "x"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("create other: %v", err)
	}

	cat := model.ArticleCategory{Name: "Mental", Slug: "mental"}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	pub := model.Article{Title: "Published", Slug: "published", Content: "ok", ArticleCategoryID: cat.ID, UserID: owner.ID, Status: model.ArticleStatusPublished}
	draft := model.Article{Title: "Draft", Slug: "draft", Content: "ok", ArticleCategoryID: cat.ID, UserID: owner.ID, Status: model.ArticleStatusDraft}
	if err := db.Create(&pub).Error; err != nil {
		t.Fatalf("create published: %v", err)
	}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatalf("create draft: %v", err)
	}

	articleSvc := service.NewArticleService(repository.NewArticleRepository(db), repository.NewArticleCategoryRepository(db), nil, nil, nil, nil)
	return NewArticleHandler(articleSvc), db, owner.ID, other.ID, cat.ID, draft.Slug
}

func TestArticleHandler_PublicEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _, _, _ := newArticleHandlerWithDB(t)

	r := gin.New()
	r.GET("/articles", h.GetArticles)
	r.GET("/articles/:slug", h.GetArticle)
	r.GET("/article-categories", h.GetCategories)

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/articles?category_id=bad", nil))
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/articles?page=1&limit=10", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/articles/not-found", nil))
	if w3.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w3.Code)
	}

	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/articles/published", nil))
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w4.Code)
	}

	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, httptest.NewRequest(http.MethodGet, "/article-categories", nil))
	if w5.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w5.Code)
	}
}

func TestArticleHandler_UserEndpointsWithOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, db, ownerID, otherID, catID, draftSlug := newArticleHandlerWithDB(t)

	r := gin.New()
	r.GET("/my-articles", func(c *gin.Context) { c.Set("user_id", ownerID); h.GetMyArticles(c) })
	r.POST("/my-articles", func(c *gin.Context) { c.Set("user_id", ownerID); h.CreateMyArticle(c) })
	r.PUT("/my-articles/:slug", func(c *gin.Context) { c.Set("user_id", ownerID); h.UpdateMyArticle(c) })
	r.DELETE("/my-articles/:slug", func(c *gin.Context) { c.Set("user_id", ownerID); h.DeleteMyArticle(c) })
	r.GET("/my-articles/:slug", func(c *gin.Context) { c.Set("user_id", otherID); h.GetArticleByIDForUser(c) })

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/my-articles?page=1&limit=10", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/my-articles", bytes.NewBufferString(`{}`))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/my-articles", bytes.NewBufferString(fmt.Sprintf(`{"title":"Mine","content":"Body","category_id":%d}`, catID)))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w3.Code)
	}

	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPut, "/my-articles/"+draftSlug, bytes.NewBufferString(fmt.Sprintf(`{"title":"Updated","content":"Body2","category_id":%d}`, catID)))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w4.Code)
	}

	// Create foreign draft article and ensure non-owner cannot see unpublished
	foreign := model.Article{Title: "Foreign Draft", Slug: "foreign-draft", Content: "x", ArticleCategoryID: catID, UserID: ownerID, Status: model.ArticleStatusDraft}
	if err := db.Create(&foreign).Error; err != nil {
		t.Fatalf("create foreign draft: %v", err)
	}

	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, httptest.NewRequest(http.MethodGet, "/my-articles/foreign-draft", nil))
	if w5.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w5.Code)
	}

	// owner can delete own article slug
	rDelete := gin.New()
	rDelete.DELETE("/my-articles/:slug", func(c *gin.Context) { c.Set("user_id", ownerID); h.DeleteMyArticle(c) })
	w6 := httptest.NewRecorder()
	rDelete.ServeHTTP(w6, httptest.NewRequest(http.MethodDelete, "/my-articles/"+draftSlug, nil))
	if w6.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w6.Code)
	}

}

func TestArticleHandler_UserMutationForbiddenAndInternalBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, db, ownerID, otherID, catID, draftSlug := newArticleHandlerWithDB(t)

	ownerDraft := model.Article{Title: "Owner Draft", Slug: "owner-draft", Content: "x", ArticleCategoryID: catID, UserID: ownerID, Status: model.ArticleStatusDraft}
	if err := db.Create(&ownerDraft).Error; err != nil {
		t.Fatalf("create owner draft: %v", err)
	}

	rUpdateForbidden := gin.New()
	rUpdateForbidden.PUT("/my-articles/:slug", func(c *gin.Context) {
		c.Set("user_id", otherID)
		h.UpdateMyArticle(c)
	})
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPut, "/my-articles/owner-draft", bytes.NewBufferString(fmt.Sprintf(`{"title":"No","content":"No","category_id":%d}`, catID)))
	req1.Header.Set("Content-Type", "application/json")
	rUpdateForbidden.ServeHTTP(w1, req1)
	if w1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 update forbidden, got %d", w1.Code)
	}

	rDeleteForbidden := gin.New()
	rDeleteForbidden.DELETE("/my-articles/:slug", func(c *gin.Context) {
		c.Set("user_id", otherID)
		h.DeleteMyArticle(c)
	})
	w2 := httptest.NewRecorder()
	rDeleteForbidden.ServeHTTP(w2, httptest.NewRequest(http.MethodDelete, "/my-articles/owner-draft", nil))
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 delete forbidden, got %d", w2.Code)
	}

	rUpdateInternal := gin.New()
	rUpdateInternal.PUT("/my-articles/:slug", func(c *gin.Context) {
		c.Set("user_id", ownerID)
		h.UpdateMyArticle(c)
	})
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPut, "/my-articles/missing-slug", bytes.NewBufferString(fmt.Sprintf(`{"title":"No","content":"No","category_id":%d}`, catID)))
	req3.Header.Set("Content-Type", "application/json")
	rUpdateInternal.ServeHTTP(w3, req3)
	if w3.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 update internal, got %d", w3.Code)
	}

	rDeleteInternal := gin.New()
	rDeleteInternal.DELETE("/my-articles/:slug", func(c *gin.Context) {
		c.Set("user_id", ownerID)
		h.DeleteMyArticle(c)
	})
	w4 := httptest.NewRecorder()
	rDeleteInternal.ServeHTTP(w4, httptest.NewRequest(http.MethodDelete, "/my-articles/"+draftSlug, nil))
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200 delete existing, got %d", w4.Code)
	}
	w5 := httptest.NewRecorder()
	rDeleteInternal.ServeHTTP(w5, httptest.NewRequest(http.MethodDelete, "/my-articles/"+draftSlug, nil))
	if w5.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 delete internal for missing slug, got %d", w5.Code)
	}
}

func TestArticleHandler_GetCategories_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	articleSvc := service.NewArticleService(repository.NewArticleRepository(db), repository.NewArticleCategoryRepository(db), nil, nil, nil, nil)
	h := NewArticleHandler(articleSvc)

	r := gin.New()
	r.GET("/article-categories", h.GetCategories)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/article-categories", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestArticleHandler_GetArticles_InternalErrorAndDefaultPaging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	h := NewArticleHandler(service.NewArticleService(repository.NewArticleRepository(db), repository.NewArticleCategoryRepository(db), nil, nil, nil, nil))

	r := gin.New()
	r.GET("/articles", h.GetArticles)

	// Missing schema should trigger internal error branch
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/articles?page=1&limit=10", nil))
	if w1.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w1.Code)
	}

	// Ensure page/limit defaults branch runs on normal flow
	h2, _, _, _, _, _ := newArticleHandlerWithDB(t)
	r2 := gin.New()
	r2.GET("/articles", h2.GetArticles)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/articles?page=0&limit=999", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
}
