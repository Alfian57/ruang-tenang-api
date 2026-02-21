package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

func hasRoute(routes []gin.RouteInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func TestRegisterBaseRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	deps := &routeDependencies{cacheService: service.NewCacheService()}

	registerBaseRoutes(r, deps)
	routes := r.Routes()

	if !hasRoute(routes, "GET", "/swagger/*any") {
		t.Fatalf("expected swagger route")
	}
	if !hasRoute(routes, "GET", "/health") {
		t.Fatalf("expected health route")
	}
	if !hasRoute(routes, "GET", "/api/v1/leaderboard") {
		t.Fatalf("expected leaderboard route")
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	r.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", healthRec.Code)
	}

	var healthBody map[string]string
	if err := json.Unmarshal(healthRec.Body.Bytes(), &healthBody); err != nil {
		t.Fatalf("unmarshal health body: %v", err)
	}
	if healthBody["status"] != "ok" {
		t.Fatalf("unexpected health status body: %+v", healthBody)
	}

	clearReq := httptest.NewRequest(http.MethodPost, "/dev/cache/clear", nil)
	clearRec := httptest.NewRecorder()
	r.ServeHTTP(clearRec, clearReq)
	if clearRec.Code != http.StatusOK {
		t.Fatalf("expected cache clear status 200, got %d", clearRec.Code)
	}

	var clearBody map[string]string
	if err := json.Unmarshal(clearRec.Body.Bytes(), &clearBody); err != nil {
		t.Fatalf("unmarshal cache clear body: %v", err)
	}
	if clearBody["message"] != "Cache cleared" {
		t.Fatalf("unexpected cache clear message: %+v", clearBody)
	}
}

func TestRegisterBaseRoutes_ReleaseModeSkipsDevRoute(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	r := gin.New()
	deps := &routeDependencies{cacheService: service.NewCacheService()}
	registerBaseRoutes(r, deps)

	if hasRoute(r.Routes(), "POST", "/dev/cache/clear") {
		t.Fatal("expected dev cache clear route to be absent in release mode")
	}
}

func TestRegisterAPIV1Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	deps := &routeDependencies{}

	registerAPIV1Routes(r, deps)
	routes := r.Routes()

	checks := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/auth/register"},
		{"POST", "/api/v1/auth/login"},
		{"GET", "/api/v1/articles"},
		{"GET", "/api/v1/article-categories"},
		{"GET", "/api/v1/forum-categories"},
		{"POST", "/api/v1/forums"},
		{"GET", "/api/v1/search"},
		{"GET", "/api/v1/community/stats"},
		{"GET", "/api/v1/features"},
		{"GET", "/api/v1/badges"},
		{"GET", "/api/v1/stories"},
		{"GET", "/api/v1/breathing/techniques"},
		{"GET", "/api/v1/daily-tasks"},
	}

	for _, check := range checks {
		if !hasRoute(routes, check.method, check.path) {
			t.Fatalf("expected route %s %s", check.method, check.path)
		}
	}
}
