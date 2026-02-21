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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserHandler_GetLeaderboardInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/leaderboard?limit=abc", nil)

	h.GetLeaderboard(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func setupUserHandlerSuccess(t *testing.T) *UserHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)

	for _, u := range []model.User{
		{Name: "U1", Username: "u1", Email: "u1@test.local", Password: "x", Exp: 900},
		{Name: "U2", Username: "u2", Email: "u2@test.local", Password: "x", Exp: 200},
		{Name: "U3", Username: "u3", Email: "u3@test.local", Password: "x", Exp: 50},
	} {
		if err := userRepo.Create(ctx, &u); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	_ = levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱", TierName: "Bronze", TierColor: "#A97142"})
	_ = levelRepo.Create(ctx, &model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐", TierName: "Silver", TierColor: "#C0C0C0"})
	_ = levelRepo.Create(ctx, &model.LevelConfig{Level: 3, MinExp: 500, BadgeName: "Pro", BadgeIcon: "🏆", TierName: "Gold", TierColor: "#FFD700"})

	return NewUserHandler(
		service.NewUserService(userRepo),
		service.NewLevelConfigService(levelRepo, service.NewCacheService()),
	)
}

func TestUserHandler_GetLeaderboardSuccessAndClamp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupUserHandlerSuccess(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/leaderboard?limit=500", nil)
	h.GetLeaderboard(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string][]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	items := payload["data"]
	if len(items) != 3 {
		t.Fatalf("expected 3 leaderboard users, got %d", len(items))
	}
	if items[0]["name"] != "U1" {
		t.Fatalf("expected top user U1, got %v", items[0]["name"])
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/leaderboard", nil)
	h.GetLeaderboard(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 with default limit, got %d", w2.Code)
	}
}

func TestUserHandler_GetLeaderboardInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	h := NewUserHandler(
		service.NewUserService(repository.NewUserRepository(db)),
		service.NewLevelConfigService(repository.NewLevelConfigRepository(db), service.NewCacheService()),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/leaderboard?limit=10", nil)
	h.GetLeaderboard(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
