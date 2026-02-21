package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAPIE2ERouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	original := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = original })

	cfg := &config.Config{
		AppEnv:             "test",
		JWTSecret:          "test-secret",
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	if err := logger.Init("test"); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	t.Cleanup(logger.Sync)

	return SetupRouter(cfg)
}

func TestAPIE2E_PublicSmokeEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupAPIE2ERouter(t)

	paths := []string{
		"/health",
		"/api/v1/search?q=",
		"/api/v1/features/categories",
		"/api/v1/forums/report-reasons",
		"/api/v1/forums/sort-options",
	}

	for _, path := range paths {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("path %s expected 200 got %d", path, w.Code)
		}
	}
}

func TestAPIE2E_HealthContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupAPIE2ERouter(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected health status ok, got %q", body["status"])
	}
}

func TestAPIE2E_AuthProtectedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupAPIE2ERouter(t)

	protectedPaths := []string{
		"/api/v1/features/my-features",
		"/api/v1/community/my-journey",
		"/api/v1/daily-tasks",
	}

	for _, path := range protectedPaths {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("path %s expected 401 got %d", path, w.Code)
		}
	}
}
