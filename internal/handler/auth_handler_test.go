package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAuthHandlerTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestAuthHandler_InvalidJSONBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil, nil)

	tests := []struct {
		name   string
		call   func(*gin.Context)
		method string
		target string
		body   string
	}{
		{name: "register-invalid-json", call: h.Register, method: http.MethodPost, target: "/auth/register", body: "{"},
		{name: "login-invalid-json", call: h.Login, method: http.MethodPost, target: "/auth/login", body: "{"},
		{name: "update-profile-invalid-json", call: h.UpdateProfile, method: http.MethodPut, target: "/auth/profile", body: "{"},
		{name: "update-password-invalid-json", call: h.UpdatePassword, method: http.MethodPut, target: "/auth/password", body: "{"},
		{name: "forgot-password-invalid-json", call: h.ForgotPassword, method: http.MethodPost, target: "/auth/forgot-password", body: "{"},
		{name: "reset-password-invalid-json", call: h.ResetPassword, method: http.MethodPost, target: "/auth/reset-password", body: "{"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newAuthHandlerTestContext(tc.method, tc.target, tc.body)
			tc.call(c)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func setupAuthHandler(t *testing.T, withSchema bool) *AuthHandler {
	t.Helper()
	config.AppConfig = &config.Config{JWTSecret: "test-secret", JWTExpiryHours: 24}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
			t.Fatalf("migrate tables: %v", err)
		}

		password1, _ := utils.HashPassword("secret123")
		password2, _ := utils.HashPassword("secret999")

		if err := db.Create(&model.User{ID: 1, Name: "User One", Email: "one@test.local", Password: password1, Role: model.RoleMember, Exp: 10}).Error; err != nil {
			t.Fatalf("seed user 1: %v", err)
		}
		if err := db.Create(&model.User{ID: 2, Name: "User Two", Email: "two@test.local", Password: password2, Role: model.RoleMember, Exp: 30}).Error; err != nil {
			t.Fatalf("seed user 2: %v", err)
		}
		if err := db.Create(&model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱"}).Error; err != nil {
			t.Fatalf("seed level config: %v", err)
		}
	}

	authService := service.NewAuthService(repository.NewUserRepository(db))
	levelService := service.NewLevelConfigService(repository.NewLevelConfigRepository(db), service.NewCacheService())
	return NewAuthHandler(authService, levelService)
}

func TestAuthHandler_AuthFlows_SuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAuthHandler(t, true)

	t.Run("register-success-and-duplicate", func(t *testing.T) {
		okCtx, okW := newAuthHandlerTestContext(http.MethodPost, "/auth/register", `{"name":"User Three","email":"three@test.local","password":"secret123"}`)
		h.Register(okCtx)
		if okW.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", okW.Code)
		}

		dupCtx, dupW := newAuthHandlerTestContext(http.MethodPost, "/auth/register", `{"name":"User Three","email":"three@test.local","password":"secret123"}`)
		h.Register(dupCtx)
		if dupW.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", dupW.Code)
		}
	})

	t.Run("login-success-and-fail", func(t *testing.T) {
		badCtx, badW := newAuthHandlerTestContext(http.MethodPost, "/auth/login", `{"email":"one@test.local","password":"wrong"}`)
		h.Login(badCtx)
		if badW.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", badW.Code)
		}

		okCtx, okW := newAuthHandlerTestContext(http.MethodPost, "/auth/login", `{"email":"one@test.local","password":"secret123"}`)
		h.Login(okCtx)
		if okW.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okW.Code)
		}
	})

	t.Run("profile-and-password", func(t *testing.T) {
		getCtx, getW := newAuthHandlerTestContext(http.MethodGet, "/auth/me", "")
		getCtx.Set("user_id", uint(1))
		h.GetProfile(getCtx)
		if getW.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", getW.Code)
		}

		missingCtx, missingW := newAuthHandlerTestContext(http.MethodGet, "/auth/me", "")
		missingCtx.Set("user_id", uint(99999))
		h.GetProfile(missingCtx)
		if missingW.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", missingW.Code)
		}

		updateCtx, updateW := newAuthHandlerTestContext(http.MethodPut, "/auth/profile", `{"name":"User One Updated","email":"one.updated@test.local","avatar":"a.png"}`)
		updateCtx.Set("user_id", uint(1))
		h.UpdateProfile(updateCtx)
		if updateW.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", updateW.Code)
		}

		dupCtx, dupW := newAuthHandlerTestContext(http.MethodPut, "/auth/profile", `{"name":"User One","email":"two@test.local"}`)
		dupCtx.Set("user_id", uint(1))
		h.UpdateProfile(dupCtx)
		if dupW.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", dupW.Code)
		}

		badPwdCtx, badPwdW := newAuthHandlerTestContext(http.MethodPut, "/auth/password", `{"current_password":"wrong","new_password":"newsecret123"}`)
		badPwdCtx.Set("user_id", uint(1))
		h.UpdatePassword(badPwdCtx)
		if badPwdW.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", badPwdW.Code)
		}

		okPwdCtx, okPwdW := newAuthHandlerTestContext(http.MethodPut, "/auth/password", `{"current_password":"secret123","new_password":"newsecret123"}`)
		okPwdCtx.Set("user_id", uint(1))
		h.UpdatePassword(okPwdCtx)
		if okPwdW.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", okPwdW.Code)
		}
	})

	t.Run("forgot-and-reset", func(t *testing.T) {
		forgotCtx, forgotW := newAuthHandlerTestContext(http.MethodPost, "/auth/forgot-password", `{"email":"one.updated@test.local"}`)
		h.ForgotPassword(forgotCtx)
		if forgotW.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", forgotW.Code)
		}

		resetBadCtx, resetBadW := newAuthHandlerTestContext(http.MethodPost, "/auth/reset-password", `{"token":"invalid","new_password":"secret987"}`)
		h.ResetPassword(resetBadCtx)
		if resetBadW.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resetBadW.Code)
		}
	})
}

func TestAuthHandler_ForgotPassword_GracefulOnLookupError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAuthHandler(t, false)

	c, w := newAuthHandlerTestContext(http.MethodPost, "/auth/forgot-password", `{"email":"x@test.local"}`)
	h.ForgotPassword(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
