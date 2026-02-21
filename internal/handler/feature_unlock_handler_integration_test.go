package handler

import (
	"context"
	"encoding/json"
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

func setupFeatureUnlockIntegrationHandler(t *testing.T) (*FeatureUnlockHandler, uint) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("migrate base tables: %v", err)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS feature_definitions (
		id TEXT PRIMARY KEY,
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
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		feature_id TEXT NOT NULL,
		unlocked_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create user_feature_unlocks table: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	featureRepo := repository.NewFeatureUnlockRepository(db)

	user := &model.User{Name: "Integration User", Username: "integration_user", Email: "integration@example.id", Password: "x", Exp: 140}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := levelRepo.Create(context.Background(), &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱", TierName: "Bronze", TierColor: "#cd7f32"}); err != nil {
		t.Fatalf("create level 1: %v", err)
	}
	if err := levelRepo.Create(context.Background(), &model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Tumbuh", BadgeIcon: "🌿", TierName: "Silver", TierColor: "#c0c0c0"}); err != nil {
		t.Fatalf("create level 2: %v", err)
	}
	if err := levelRepo.Create(context.Background(), &model.LevelConfig{Level: 3, MinExp: 200, BadgeName: "Maju", BadgeIcon: "🌳", TierName: "Gold", TierColor: "#ffd700"}); err != nil {
		t.Fatalf("create level 3: %v", err)
	}

	seedFeatures := []model.FeatureDefinition{
		{ID: uuid.New(), FeatureKey: "chat_pro", FeatureName: "Chat Pro", Description: "AI assistant", Icon: "bot", RequiredLevel: 1, Category: "ai", IsActive: true},
		{ID: uuid.New(), FeatureKey: "story_plus", FeatureName: "Story Plus", Description: "Story tools", Icon: "edit", RequiredLevel: 2, Category: "content", IsActive: true},
		{ID: uuid.New(), FeatureKey: "elite_zone", FeatureName: "Elite Zone", Description: "Special area", Icon: "crown", RequiredLevel: 3, Category: "special", IsActive: true},
	}

	for _, f := range seedFeatures {
		feature := f
		if err := featureRepo.CreateFeatureDefinition(context.Background(), &feature); err != nil {
			t.Fatalf("create feature definition %s: %v", feature.FeatureKey, err)
		}
	}

	h := NewFeatureUnlockHandler(service.NewFeatureUnlockService(featureRepo, levelRepo, userRepo))
	return h, user.ID
}

func TestFeatureUnlockHandler_IntegrationEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, userID := setupFeatureUnlockIntegrationHandler(t)
	r := gin.New()

	r.GET("/api/v1/features", h.GetAllFeatures)
	r.GET("/api/v1/features/categories", h.GetFeatureCategories)
	r.GET("/api/v1/features/category/:category", h.GetFeaturesByCategory)
	r.GET("/api/v1/features/my-features", func(c *gin.Context) { c.Set("user_id", userID); h.GetUserFeatures(c) })
	r.GET("/api/v1/features/check/:featureKey", func(c *gin.Context) { c.Set("user_id", userID); h.CheckFeatureAccess(c) })
	r.GET("/api/v1/features/upcoming", func(c *gin.Context) { c.Set("user_id", userID); h.GetUpcomingFeatures(c) })

	r.GET("/api/v1/features/my-features/unauth", h.GetUserFeatures)
	r.GET("/api/v1/features/check/:featureKey/unauth", h.CheckFeatureAccess)
	r.GET("/api/v1/features/upcoming/unauth", h.GetUpcomingFeatures)

	t.Run("public endpoints", func(t *testing.T) {
		cases := []string{
			"/api/v1/features",
			"/api/v1/features/categories",
			"/api/v1/features/category/ai",
		}

		for _, path := range cases {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("path %s expected 200 got %d", path, w.Code)
			}
		}
	})

	t.Run("authorized endpoints", func(t *testing.T) {
		cases := []string{
			"/api/v1/features/my-features",
			"/api/v1/features/check/chat_pro",
			"/api/v1/features/check/non-existent-feature",
			"/api/v1/features/upcoming",
		}

		for _, path := range cases {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("path %s expected 200 got %d", path, w.Code)
			}
		}
	})

	t.Run("unauthorized endpoints", func(t *testing.T) {
		cases := []string{
			"/api/v1/features/my-features/unauth",
			"/api/v1/features/check/chat_pro/unauth",
			"/api/v1/features/upcoming/unauth",
		}

		for _, path := range cases {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("path %s expected 401 got %d", path, w.Code)
			}
		}
	})

	t.Run("response payload includes data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/features", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", w.Code)
		}

		var payload map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if payload["data"] == nil {
			t.Fatal("expected response data to be present")
		}
	})
}
