package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareRequiresVerifiedPhoneClaim(t *testing.T) {
	previousConfig := config.AppConfig
	previousResolver := accountStatusResolver
	accountStatusResolver = nil
	config.AppConfig = &config.Config{JWTSecret: "unit-test-signing-key"}
	defer func() {
		config.AppConfig = previousConfig
		accountStatusResolver = previousResolver
		if previousConfig != nil {
			utils.InitializeKeyManager()
		}
	}()
	utils.InitializeKeyManager()
	router := gin.New()
	router.GET("/private", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, tc := range []struct {
		verified bool
		want     int
	}{
		{false, http.StatusUnauthorized},
		{true, http.StatusNoContent},
	} {
		token, err := utils.GenerateToken(1, "user@example.com", "user", time.Hour, tc.verified)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(http.MethodGet, "/private", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != tc.want {
			t.Fatalf("verified=%v: status=%d, want=%d", tc.verified, recorder.Code, tc.want)
		}
	}
	accountStatusResolver = func(uint) AccountStatus {
		return AccountStatus{Allowed: false, RequirePhoneVerification: true}
	}
	token, err := utils.GenerateToken(1, "user@example.com", "user", time.Hour, true)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("changed phone: status=%d", recorder.Code)
	}
}
