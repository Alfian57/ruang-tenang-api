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
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newFeatureUnlockHandlerForTest(t *testing.T) (*FeatureUnlockHandler, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("migrate base tables: %v", err)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS feature_definitions (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		feature_key TEXT NOT NULL UNIQUE,
		feature_name TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		required_level INTEGER NOT NULL DEFAULT 1,
		category TEXT DEFAULT 'general',
		is_active NUMERIC DEFAULT 1,
		display_order INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create feature_definitions table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS user_feature_unlocks (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		user_id INTEGER NOT NULL,
		feature_id TEXT NOT NULL,
		unlocked_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create user_feature_unlocks table: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	featureRepo := repository.NewFeatureUnlockRepository(db)

	user := &model.User{Name: "FU Handler", Username: "fuhandler", Email: "fuhandler@example.id", Password: "x", Exp: 120}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	_ = levelRepo.Create(context.Background(), &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱"})
	_ = levelRepo.Create(context.Background(), &model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐"})
	_ = levelRepo.Create(context.Background(), &model.LevelConfig{Level: 3, MinExp: 200, BadgeName: "Maju", BadgeIcon: "🚀"})

	_ = featureRepo.CreateFeatureDefinition(context.Background(), &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "chat_pro", FeatureName: "Chat Pro", RequiredLevel: 1, Category: "ai", IsActive: true})
	_ = featureRepo.CreateFeatureDefinition(context.Background(), &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "story_plus", FeatureName: "Story Plus", RequiredLevel: 2, Category: "content", IsActive: true})
	_ = featureRepo.CreateFeatureDefinition(context.Background(), &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "pro_zone", FeatureName: "Pro Zone", RequiredLevel: 3, Category: "special", IsActive: true})

	svc := service.NewFeatureUnlockService(featureRepo, levelRepo, userRepo)
	return NewFeatureUnlockHandler(svc), user.ID
}

func newHandlerContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, w
}

func TestFeatureUnlockHandler_PublicEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := newFeatureUnlockHandlerForTest(t)

	{
		c, w := newHandlerContext(http.MethodGet, "/features")
		h.GetAllFeatures(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetAllFeatures, got %d", w.Code)
		}
	}

	{
		c, w := newHandlerContext(http.MethodGet, "/features/category/ai")
		c.Params = gin.Params{{Key: "category", Value: "ai"}}
		h.GetFeaturesByCategory(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetFeaturesByCategory, got %d", w.Code)
		}
	}

	{
		c, w := newHandlerContext(http.MethodGet, "/features/categories")
		h.GetFeatureCategories(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetFeatureCategories, got %d", w.Code)
		}
	}
}

func TestFeatureUnlockHandler_AuthRequiredEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, userID := newFeatureUnlockHandlerForTest(t)

	// Unauthorized branches
	for name, call := range map[string]func(*gin.Context){
		"user-features": h.GetUserFeatures,
		"check-access":  h.CheckFeatureAccess,
		"upcoming":      h.GetUpcomingFeatures,
	} {
		t.Run("unauthorized-"+name, func(t *testing.T) {
			c, w := newHandlerContext(http.MethodGet, "/")
			if name == "check-access" {
				c.Params = gin.Params{{Key: "featureKey", Value: "chat_pro"}}
			}
			call(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 for %s, got %d", name, w.Code)
			}
		})
	}

	// Authorized: user features
	{
		c, w := newHandlerContext(http.MethodGet, "/features/my-features")
		c.Set("user_id", userID)
		h.GetUserFeatures(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetUserFeatures, got %d", w.Code)
		}
	}

	// Authorized: check feature access existing and unknown
	{
		c, w := newHandlerContext(http.MethodGet, "/features/check/chat_pro")
		c.Set("user_id", userID)
		c.Params = gin.Params{{Key: "featureKey", Value: "chat_pro"}}
		h.CheckFeatureAccess(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 CheckFeatureAccess existing, got %d", w.Code)
		}
	}
	{
		c, w := newHandlerContext(http.MethodGet, "/features/check/unknown")
		c.Set("user_id", userID)
		c.Params = gin.Params{{Key: "featureKey", Value: "unknown"}}
		h.CheckFeatureAccess(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 CheckFeatureAccess unknown, got %d", w.Code)
		}
	}

	// Authorized: upcoming features
	{
		c, w := newHandlerContext(http.MethodGet, "/features/upcoming")
		c.Set("user_id", userID)
		h.GetUpcomingFeatures(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetUpcomingFeatures, got %d", w.Code)
		}
	}
}

func TestFeatureUnlockHandler_InternalErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("migrate base tables: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	featureRepo := repository.NewFeatureUnlockRepository(db)

	user := &model.User{Name: "Err User", Username: "erruser", Email: "erruser@example.id", Password: "x", Exp: 10}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := NewFeatureUnlockHandler(service.NewFeatureUnlockService(featureRepo, levelRepo, userRepo))

	t.Run("public endpoints return 500", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			call       func(*gin.Context)
			set        func(*gin.Context)
			expectCode int
		}{
			{name: "all-features", call: h.GetAllFeatures, expectCode: http.StatusInternalServerError},
			{name: "by-category", call: h.GetFeaturesByCategory, set: func(c *gin.Context) { c.Params = gin.Params{{Key: "category", Value: "ai"}} }, expectCode: http.StatusInternalServerError},
			{name: "categories", call: h.GetFeatureCategories, expectCode: http.StatusOK},
		} {
			t.Run(tc.name, func(t *testing.T) {
				c, w := newHandlerContext(http.MethodGet, "/")
				if tc.set != nil {
					tc.set(c)
				}
				tc.call(c)
				if w.Code != tc.expectCode {
					t.Fatalf("expected %d for %s, got %d", tc.expectCode, tc.name, w.Code)
				}
			})
		}
	})

	t.Run("auth endpoints return 500", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			call       func(*gin.Context)
			set        func(*gin.Context)
			expectCode int
		}{
			{name: "user-features", call: h.GetUserFeatures, expectCode: http.StatusInternalServerError},
			{name: "check-access", call: h.CheckFeatureAccess, set: func(c *gin.Context) { c.Params = gin.Params{{Key: "featureKey", Value: "chat_pro"}} }, expectCode: http.StatusOK},
			{name: "upcoming", call: h.GetUpcomingFeatures, expectCode: http.StatusInternalServerError},
		} {
			t.Run(tc.name, func(t *testing.T) {
				c, w := newHandlerContext(http.MethodGet, "/")
				c.Set("user_id", user.ID)
				if tc.set != nil {
					tc.set(c)
				}
				tc.call(c)
				if w.Code != tc.expectCode {
					t.Fatalf("expected %d for %s, got %d", tc.expectCode, tc.name, w.Code)
				}
			})
		}
	})
}
