package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInspiringStoryReadHandler(t *testing.T) *InspiringStoryHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	storyRepo := repository.NewInspiringStoryRepository(db)
	svc := service.NewInspiringStoryService(storyRepo, nil, nil, nil, nil, nil)
	return NewInspiringStoryHandler(svc)
}

func newStoryGuardContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestInspiringStoryHandler_UnauthorizedGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &InspiringStoryHandler{}

	tests := []struct {
		name string
		call func(*gin.Context)
	}{
		{"CreateStory", h.CreateStory},
		{"UpdateStory", h.UpdateStory},
		{"DeleteStory", h.DeleteStory},
		{"GetMyStories", h.GetMyStories},
		{"ToggleHeart", h.ToggleHeart},
		{"CreateComment", h.CreateComment},
		{"DeleteComment", h.DeleteComment},
		{"ToggleCommentHeart", h.ToggleCommentHeart},
		{"GetMyStats", h.GetMyStats},
		{"GetPendingStories", h.GetPendingStories},
		{"ModerateStory", h.ModerateStory},
		{"SetFeatured", h.SetFeatured},
		{"HideComment", h.HideComment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newStoryGuardContext(http.MethodPost, "/", "{}")
			c.Params = gin.Params{
				{Key: "id", Value: "bad-id"},
				{Key: "commentId", Value: "bad-comment"},
			}
			tt.call(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestInspiringStoryHandler_ForbiddenRoleGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &InspiringStoryHandler{}

	tests := []struct {
		name string
		call func(*gin.Context)
	}{
		{"GetPendingStories", h.GetPendingStories},
		{"ModerateStory", h.ModerateStory},
		{"SetFeatured", h.SetFeatured},
		{"HideComment", h.HideComment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newStoryGuardContext(http.MethodPost, "/", "{}")
			c.Set("user_role", model.RoleMember)
			c.Set("user_id", uint(1))
			c.Params = gin.Params{
				{Key: "id", Value: "bad-id"},
				{Key: "commentId", Value: "bad-comment"},
			}
			tt.call(c)
			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestInspiringStoryHandler_NewHandler(t *testing.T) {
	h := NewInspiringStoryHandler(nil)
	if h == nil {
		t.Fatal("expected handler instance")
	}
}

func TestInspiringStoryHandler_InvalidIDAndPayloadGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &InspiringStoryHandler{}
	validID := "11111111-1111-1111-1111-111111111111"

	t.Run("public-invalid-story-id", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodGet, "/stories/bad-id", "")
		c1.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.GetStory(c1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodGet, "/stories/bad-id/comments", "")
		c2.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.GetComments(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}
	})

	t.Run("auth-invalid-json-and-id", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodPost, "/stories", "{")
		c1.Set("user_id", uint(1))
		h.CreateStory(c1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodPut, "/stories/bad-id", "{}")
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.UpdateStory(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		c3, w3 := newStoryGuardContext(http.MethodPut, "/stories/valid-id", "{")
		c3.Set("user_id", uint(1))
		c3.Params = gin.Params{{Key: "id", Value: validID}}
		h.UpdateStory(c3)
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		c4, w4 := newStoryGuardContext(http.MethodDelete, "/stories/bad-id", "")
		c4.Set("user_id", uint(1))
		c4.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.DeleteStory(c4)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w4.Code)
		}

		c5, w5 := newStoryGuardContext(http.MethodPost, "/stories/bad-id/heart", "")
		c5.Set("user_id", uint(1))
		c5.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.ToggleHeart(c5)
		if w5.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w5.Code)
		}

		c6, w6 := newStoryGuardContext(http.MethodPost, "/stories/bad-id/comments", "{}")
		c6.Set("user_id", uint(1))
		c6.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.CreateComment(c6)
		if w6.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w6.Code)
		}

		c7, w7 := newStoryGuardContext(http.MethodPost, "/stories/valid-id/comments", "{")
		c7.Set("user_id", uint(1))
		c7.Params = gin.Params{{Key: "id", Value: validID}}
		h.CreateComment(c7)
		if w7.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w7.Code)
		}

		c8, w8 := newStoryGuardContext(http.MethodDelete, "/stories/valid-id/comments/bad-comment", "")
		c8.Set("user_id", uint(1))
		c8.Params = gin.Params{{Key: "commentId", Value: "bad-comment"}}
		h.DeleteComment(c8)
		if w8.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w8.Code)
		}

		c9, w9 := newStoryGuardContext(http.MethodPost, "/stories/valid-id/comments/bad-comment/heart", "")
		c9.Set("user_id", uint(1))
		c9.Params = gin.Params{{Key: "commentId", Value: "bad-comment"}}
		h.ToggleCommentHeart(c9)
		if w9.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w9.Code)
		}
	})

	t.Run("admin-invalid-id-and-json", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodPost, "/admin/stories/bad-id/moderate", "{}")
		c1.Set("user_id", uint(1))
		c1.Set("user_role", model.RoleAdmin)
		c1.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.ModerateStory(c1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodPost, "/admin/stories/valid-id/moderate", "{")
		c2.Set("user_id", uint(1))
		c2.Set("user_role", model.RoleAdmin)
		c2.Params = gin.Params{{Key: "id", Value: validID}}
		h.ModerateStory(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		c3, w3 := newStoryGuardContext(http.MethodPost, "/admin/stories/bad-id/featured", "")
		c3.Set("user_id", uint(1))
		c3.Set("user_role", model.RoleAdmin)
		c3.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		h.SetFeatured(c3)
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		c4, w4 := newStoryGuardContext(http.MethodPost, "/admin/stories/valid-id/comments/bad-comment/hide", "{}")
		c4.Set("user_role", model.RoleAdmin)
		c4.Params = gin.Params{{Key: "commentId", Value: "bad-comment"}}
		h.HideComment(c4)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w4.Code)
		}

		c5, w5 := newStoryGuardContext(http.MethodPost, "/admin/stories/valid-id/comments/valid-comment/hide", "{")
		c5.Set("user_role", model.RoleAdmin)
		c5.Params = gin.Params{{Key: "commentId", Value: validID}}
		h.HideComment(c5)
		if w5.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w5.Code)
		}
	})
}

func TestInspiringStoryHandler_PublicReadBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupInspiringStoryReadHandler(t)

	t.Run("get-stories-invalid-query", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories?page=abc", "")
		h.GetStories(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("get-categories-service-error", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/categories", "")
		h.GetCategories(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("get-featured-service-error", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/featured?limit=3", "")
		h.GetFeaturedStories(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("get-most-appreciated-service-error", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/most-appreciated?month=1&year=2025&limit=5", "")
		h.GetMostAppreciated(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func setupInspiringStorySuccessHandler(t *testing.T) *InspiringStoryHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	now := time.Now()
	schema := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, avatar TEXT, exp INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE level_configs (id INTEGER PRIMARY KEY, level INTEGER, min_exp INTEGER, badge_name TEXT, badge_icon TEXT, tier_name TEXT, tier_color TEXT, description TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (
			id TEXT PRIMARY KEY,
			author_id INTEGER,
			title TEXT,
			content TEXT,
			cover_image TEXT,
			is_anonymous BOOLEAN,
			has_trigger_warning BOOLEAN,
			trigger_warning_text TEXT,
			status TEXT,
			moderated_by INTEGER,
			moderator_feedback TEXT,
			moderator_id INTEGER,
			moderation_feedback TEXT,
			moderated_at DATETIME,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_featured BOOLEAN,
			featured_by INTEGER,
			featured_at DATETIME,
			featured_until DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			published_at DATETIME
		)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, inspiring_story_id TEXT, story_category_id TEXT)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, heart_count INTEGER, is_hidden BOOLEAN, hidden_reason TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO users (id, name, avatar, exp, deleted_at) VALUES (1, 'Author', 'a.png', 700, NULL), (2, 'Viewer', 'b.png', 100, NULL)`).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (1, 1, 0, 'Pemula', '🌱', 'Bronze', '#A97142', 'd', ?, ?), (2, 7, 500, 'Pro', '🏆', 'Gold', '#FFD700', 'd', ?, ?)`, now, now, now, now).Error; err != nil {
		t.Fatalf("seed levels: %v", err)
	}

	storyID := "11111111-1111-1111-1111-111111111111"
	catID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES (?, 1, 'Story A', 'Story content', 'approved', 0, 3, 2, 1, 0, ?, ?, ?)`, storyID, now, now, now).Error; err != nil {
		t.Fatalf("seed story: %v", err)
	}
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES (?, 1, 'Story Pending', 'Pending content', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL)`, "33333333-3333-3333-3333-333333333333", now, now).Error; err != nil {
		t.Fatalf("seed pending story: %v", err)
	}
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES
		('44444444-4444-4444-4444-444444444444', 2, 'U2 Story 1', 'content1', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL),
		('55555555-5555-5555-5555-555555555555', 2, 'U2 Story 2', 'content2', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL),
		('66666666-6666-6666-6666-666666666666', 2, 'U2 Story 3', 'content3', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL)`, now, now, now, now, now, now).Error; err != nil {
		t.Fatalf("seed user2 monthly stories: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_categories (id, name, slug, description, icon, display_order, is_active, created_at, updated_at) VALUES (?, 'Hope', 'hope', 'desc', '🌟', 1, 1, ?, ?)`, catID, now, now).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_category_relations (story_id, category_id, inspiring_story_id, story_category_id) VALUES (?, ?, ?, ?)`, storyID, catID, storyID, catID).Error; err != nil {
		t.Fatalf("seed relation: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_tags (id, story_id, tag, created_at) VALUES (?, ?, 'hope', ?)`, uuid.New().String(), storyID, now).Error; err != nil {
		t.Fatalf("seed tags: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, heart_count, is_hidden, created_at, updated_at) VALUES (?, ?, 2, 'supportive', 0, 0, ?, ?)`, "22222222-2222-2222-2222-222222222222", storyID, now, now).Error; err != nil {
		t.Fatalf("seed comments: %v", err)
	}

	storyRepo := repository.NewInspiringStoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	svc := service.NewInspiringStoryService(storyRepo, userRepo, levelRepo, nil, nil, nil)
	return NewInspiringStoryHandler(svc)
}

func TestInspiringStoryHandler_PublicAndMyStories_SuccessPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupInspiringStorySuccessHandler(t)

	t.Run("get-story-success-and-notfound", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodGet, "/stories/11111111-1111-1111-1111-111111111111", "")
		c1.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.GetStory(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		cAuth, wAuth := newStoryGuardContext(http.MethodGet, "/stories/11111111-1111-1111-1111-111111111111", "")
		cAuth.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		cAuth.Set("user_id", uint(2))
		h.GetStory(cAuth)
		if wAuth.Code != http.StatusOK {
			t.Fatalf("expected 200 for authenticated viewer, got %d", wAuth.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodGet, "/stories/99999999-9999-9999-9999-999999999999", "")
		c2.Params = gin.Params{{Key: "id", Value: "99999999-9999-9999-9999-999999999999"}}
		h.GetStory(c2)
		if w2.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w2.Code)
		}
	})

	t.Run("get-stories-success", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories?page=1&limit=10", "")
		h.GetStories(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-my-stories-success", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/my-stories?page=1&limit=10", "")
		c.Set("user_id", uint(1))
		h.GetMyStories(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-comments-success", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/11111111-1111-1111-1111-111111111111/comments?page=1&limit=10", "")
		c.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.GetComments(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestInspiringStoryHandler_AdminAndInteractionBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupInspiringStorySuccessHandler(t)

	t.Run("toggle-heart-service-error", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodPost, "/stories/33333333-3333-3333-3333-333333333333/heart", "")
		c.Set("user_id", uint(2))
		c.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.ToggleHeart(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("create-update-delete-story-branches", func(t *testing.T) {
		createReq := `{"title":"Cerita Panjang Sekali","content":"` + strings.Repeat("a", 220) + `","category_ids":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"]}`

		partialDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := partialDB.AutoMigrate(&model.User{}); err != nil {
			t.Fatalf("migrate users: %v", err)
		}
		if err := partialDB.Create(&model.User{ID: 1, Name: "User One", Username: "uone", Email: "uone@test.local", Password: "x", Role: model.RoleMember, Exp: 10}).Error; err != nil {
			t.Fatalf("seed partial user: %v", err)
		}
		hCreateErr := NewInspiringStoryHandler(service.NewInspiringStoryService(
			repository.NewInspiringStoryRepository(partialDB),
			repository.NewUserRepository(partialDB),
			repository.NewLevelConfigRepository(partialDB),
			nil, nil, nil,
		))

		createInternalErr, wcie := newStoryGuardContext(http.MethodPost, "/stories", createReq)
		createInternalErr.Set("user_id", uint(1))
		hCreateErr.CreateStory(createInternalErr)
		if wcie.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 create story internal error, got %d", wcie.Code)
		}

		createSvcErr, wcse := newStoryGuardContext(http.MethodPost, "/stories", createReq)
		createSvcErr.Set("user_id", uint(2))
		h.CreateStory(createSvcErr)
		if wcse.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 monthly limit, got %d", wcse.Code)
		}

		updateOK, wuo := newStoryGuardContext(http.MethodPut, "/stories/33333333-3333-3333-3333-333333333333", `{"title":"Updated Pending"}`)
		updateOK.Set("user_id", uint(1))
		updateOK.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.UpdateStory(updateOK)
		if wuo.Code != http.StatusOK {
			t.Fatalf("expected 200 update, got %d", wuo.Code)
		}

		updateForbidden, wuf := newStoryGuardContext(http.MethodPut, "/stories/33333333-3333-3333-3333-333333333333", `{"title":"Valid Title"}`)
		updateForbidden.Set("user_id", uint(2))
		updateForbidden.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.UpdateStory(updateForbidden)
		if wuf.Code != http.StatusForbidden {
			t.Fatalf("expected 403 update forbidden, got %d", wuf.Code)
		}

		updateMissing, wum := newStoryGuardContext(http.MethodPut, "/stories/missing", `{"title":"Valid Title"}`)
		updateMissing.Set("user_id", uint(1))
		updateMissing.Params = gin.Params{{Key: "id", Value: "99999999-9999-9999-9999-999999999999"}}
		h.UpdateStory(updateMissing)
		if wum.Code != http.StatusNotFound {
			t.Fatalf("expected 404 update not found, got %d", wum.Code)
		}

		deleteForbidden, wdf := newStoryGuardContext(http.MethodDelete, "/stories/33333333-3333-3333-3333-333333333333", "")
		deleteForbidden.Set("user_id", uint(2))
		deleteForbidden.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.DeleteStory(deleteForbidden)
		if wdf.Code != http.StatusForbidden {
			t.Fatalf("expected 403 delete forbidden, got %d", wdf.Code)
		}

		deleteOK, wdo := newStoryGuardContext(http.MethodDelete, "/stories/33333333-3333-3333-3333-333333333333", "")
		deleteOK.Set("user_id", uint(1))
		deleteOK.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.DeleteStory(deleteOK)
		if wdo.Code != http.StatusOK {
			t.Fatalf("expected 200 delete success, got %d", wdo.Code)
		}

		deleteMissing, wdm := newStoryGuardContext(http.MethodDelete, "/stories/missing", "")
		deleteMissing.Set("user_id", uint(1))
		deleteMissing.Params = gin.Params{{Key: "id", Value: "99999999-9999-9999-9999-999999999999"}}
		h.DeleteStory(deleteMissing)
		if wdm.Code != http.StatusNotFound {
			t.Fatalf("expected 404 delete missing, got %d", wdm.Code)
		}
	})

	t.Run("toggle-comment-heart-success-and-notfound", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodPost, "/stories/111/comments/222/heart", "")
		c1.Set("user_id", uint(1))
		c1.Params = gin.Params{{Key: "commentId", Value: "22222222-2222-2222-2222-222222222222"}}
		h.ToggleCommentHeart(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodPost, "/stories/111/comments/missing/heart", "")
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "commentId", Value: "99999999-9999-9999-9999-999999999999"}}
		h.ToggleCommentHeart(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}
	})

	t.Run("delete-comment-forbidden-and-notfound", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodDelete, "/stories/111/comments/222", "")
		c1.Set("user_id", uint(1))
		c1.Params = gin.Params{{Key: "commentId", Value: "22222222-2222-2222-2222-222222222222"}}
		h.DeleteComment(c1)
		if w1.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodDelete, "/stories/111/comments/missing", "")
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "commentId", Value: "99999999-9999-9999-9999-999999999999"}}
		h.DeleteComment(c2)
		if w2.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w2.Code)
		}
	})

	t.Run("my-stats-success", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/my-stats", "")
		c.Set("user_id", uint(1))
		h.GetMyStats(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("admin-pending-feature-hide-and-moderate-error", func(t *testing.T) {
		pending, wp := newStoryGuardContext(http.MethodGet, "/admin/stories/pending?page=1&limit=10", "")
		pending.Set("user_role", model.RoleAdmin)
		h.GetPendingStories(pending)
		if wp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wp.Code)
		}

		featured, wf := newStoryGuardContext(http.MethodPost, "/admin/stories/111/featured?featured=true", "")
		featured.Set("user_id", uint(1))
		featured.Set("user_role", model.RoleAdmin)
		featured.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.SetFeatured(featured)
		if wf.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wf.Code)
		}

		featuredOff, wfo := newStoryGuardContext(http.MethodPost, "/admin/stories/111/featured?featured=false", "")
		featuredOff.Set("user_id", uint(1))
		featuredOff.Set("user_role", model.RoleModerator)
		featuredOff.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.SetFeatured(featuredOff)
		if wfo.Code != http.StatusOK {
			t.Fatalf("expected 200 featured off, got %d", wfo.Code)
		}

		hide, wh := newStoryGuardContext(http.MethodPost, "/admin/stories/111/comments/222/hide", `{"reason":"moderation"}`)
		hide.Set("user_role", model.RoleAdmin)
		hide.Params = gin.Params{{Key: "commentId", Value: "22222222-2222-2222-2222-222222222222"}}
		h.HideComment(hide)
		if wh.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wh.Code)
		}

		moderate, wm := newStoryGuardContext(http.MethodPost, "/admin/stories/missing/moderate", `{"status":"approved"}`)
		moderate.Set("user_id", uint(1))
		moderate.Set("user_role", model.RoleAdmin)
		moderate.Params = gin.Params{{Key: "id", Value: "99999999-9999-9999-9999-999999999999"}}
		h.ModerateStory(moderate)
		if wm.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", wm.Code)
		}
	})
}

func TestInspiringStoryHandler_CreateComment_AdditionalBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupInspiringStorySuccessHandler(t)

	t.Run("create-comment-story-not-approved", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodPost, "/stories/33333333-3333-3333-3333-333333333333/comments", `{"content":"Semangat terus"}`)
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.CreateComment(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("create-comment-internal-error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		now := time.Now()
		if err := db.Exec(`CREATE TABLE inspiring_stories (
			id TEXT PRIMARY KEY,
			author_id INTEGER,
			title TEXT,
			content TEXT,
			status TEXT,
			is_anonymous BOOLEAN,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_featured BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME,
			published_at DATETIME
		)`).Error; err != nil {
			t.Fatalf("schema story: %v", err)
		}
		if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at)
			VALUES ('77777777-7777-7777-7777-777777777777', 1, 'Approved Story', 'content', 'approved', 0, 0, 0, 0, 0, ?, ?, ?)`, now, now, now).Error; err != nil {
			t.Fatalf("seed story: %v", err)
		}

		storyRepo := repository.NewInspiringStoryRepository(db)
		svc := service.NewInspiringStoryService(storyRepo, nil, nil, nil, nil, nil)
		h2 := NewInspiringStoryHandler(svc)

		c, w := newStoryGuardContext(http.MethodPost, "/stories/77777777-7777-7777-7777-777777777777/comments", `{"content":"Komentar dukungan"}`)
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "id", Value: "77777777-7777-7777-7777-777777777777"}}
		h2.CreateComment(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestInspiringStoryHandler_Admin_MissingRoleBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupInspiringStorySuccessHandler(t)

	c1, w1 := newStoryGuardContext(http.MethodPost, "/admin/stories/11111111-1111-1111-1111-111111111111/moderate", `{"status":"approved"}`)
	c1.Set("user_id", uint(1))
	c1.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
	h.ModerateStory(c1)
	if w1.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 missing role on moderate, got %d", w1.Code)
	}

	c2, w2 := newStoryGuardContext(http.MethodPost, "/admin/stories/11111111-1111-1111-1111-111111111111/featured?featured=true", "")
	c2.Set("user_id", uint(1))
	c2.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
	h.SetFeatured(c2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 missing role on set-featured, got %d", w2.Code)
	}
}

func TestInspiringStoryHandler_ToggleHeartAndModerate_SuccessBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupInspiringStorySuccessHandler(t)

	t.Run("toggle-heart-add-remove-success", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodPost, "/stories/11111111-1111-1111-1111-111111111111/heart", "")
		c1.Set("user_id", uint(2))
		c1.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.ToggleHeart(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 add heart, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodPost, "/stories/11111111-1111-1111-1111-111111111111/heart", "")
		c2.Set("user_id", uint(2))
		c2.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		h.ToggleHeart(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 remove heart, got %d", w2.Code)
		}
	})

	t.Run("moderate-story-success-status-branches", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodPost, "/admin/stories/33333333-3333-3333-3333-333333333333/moderate", `{"status":"approved","feedback":"ok"}`)
		c1.Set("user_id", uint(1))
		c1.Set("user_role", model.RoleAdmin)
		c1.Params = gin.Params{{Key: "id", Value: "33333333-3333-3333-3333-333333333333"}}
		h.ModerateStory(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 approve moderation, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodPost, "/admin/stories/44444444-4444-4444-4444-444444444444/moderate", `{"status":"rejected","feedback":"tidak sesuai"}`)
		c2.Set("user_id", uint(1))
		c2.Set("user_role", model.RoleAdmin)
		c2.Params = gin.Params{{Key: "id", Value: "44444444-4444-4444-4444-444444444444"}}
		h.ModerateStory(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 reject moderation, got %d", w2.Code)
		}

		c3, w3 := newStoryGuardContext(http.MethodPost, "/admin/stories/55555555-5555-5555-5555-555555555555/moderate", `{"status":"revision_requested","feedback":"revisi"}`)
		c3.Set("user_id", uint(1))
		c3.Set("user_role", model.RoleModerator)
		c3.Params = gin.Params{{Key: "id", Value: "55555555-5555-5555-5555-555555555555"}}
		h.ModerateStory(c3)
		if w3.Code != http.StatusOK {
			t.Fatalf("expected 200 revision moderation, got %d", w3.Code)
		}
	})

	t.Run("set-featured-service-error-branch", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodPost, "/admin/stories/99999999-9999-9999-9999-999999999999/featured?featured=true", "")
		c.Set("user_id", uint(1))
		c.Set("user_role", model.RoleAdmin)
		c.Params = gin.Params{{Key: "id", Value: "99999999-9999-9999-9999-999999999999"}}
		h.SetFeatured(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 set featured service error, got %d", w.Code)
		}
	})

	t.Run("delete-comment-success", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodDelete, "/stories/111/comments/222", "")
		c.Set("user_id", uint(2))
		c.Params = gin.Params{{Key: "commentId", Value: "22222222-2222-2222-2222-222222222222"}}
		h.DeleteComment(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 delete comment success, got %d", w.Code)
		}
	})
}

func TestInspiringStoryHandler_ExtraCoverageBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	successH := setupInspiringStorySuccessHandler(t)
	errorH := setupInspiringStoryReadHandler(t)

	t.Run("categories success", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/categories", "")
		successH.GetCategories(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("public read internal errors", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodGet, "/stories/11111111-1111-1111-1111-111111111111", "")
		c1.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		errorH.GetStory(c1)
		if w1.Code != http.StatusNotFound {
			t.Fatalf("expected 404 get story, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodGet, "/stories?page=1&limit=10", "")
		errorH.GetStories(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 get stories, got %d", w2.Code)
		}

		c3, w3 := newStoryGuardContext(http.MethodGet, "/stories/11111111-1111-1111-1111-111111111111/comments?page=1&limit=10", "")
		c3.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
		errorH.GetComments(c3)
		if w3.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 get comments, got %d", w3.Code)
		}
	})

	t.Run("my stories internal errors", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodGet, "/stories/my-stories?page=1&limit=10", "")
		c1.Set("user_id", uint(1))
		errorH.GetMyStories(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 my stories, got %d", w1.Code)
		}
	})

	t.Run("my stats internal error", func(t *testing.T) {
		c, w := newStoryGuardContext(http.MethodGet, "/stories/my-stats", "")
		c.Set("user_id", uint(1))
		errorH.GetMyStats(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 my stats, got %d", w.Code)
		}
	})

	t.Run("featured and most appreciated success", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodGet, "/stories/featured?limit=3", "")
		successH.GetFeaturedStories(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 featured, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodGet, "/stories/most-appreciated?month=1&year=2025&limit=5", "")
		successH.GetMostAppreciated(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 most appreciated, got %d", w2.Code)
		}
	})

	t.Run("admin pending internal error and hide internal error", func(t *testing.T) {
		c1, w1 := newStoryGuardContext(http.MethodGet, "/admin/stories/pending?page=1&limit=10", "")
		c1.Set("user_role", model.RoleAdmin)
		errorH.GetPendingStories(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 pending stories, got %d", w1.Code)
		}

		c2, w2 := newStoryGuardContext(http.MethodPost, "/admin/stories/111/comments/222/hide", `{"reason":"x"}`)
		c2.Set("user_role", model.RoleAdmin)
		c2.Params = gin.Params{{Key: "commentId", Value: "11111111-1111-1111-1111-111111111111"}}
		errorH.HideComment(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 hide comment, got %d", w2.Code)
		}
	})
}
