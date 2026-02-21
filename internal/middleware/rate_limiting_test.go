package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewRateLimiter_DefaultTierFallback(t *testing.T) {
	rl := NewRateLimiter(RateLimitTier("unknown"))
	defer rl.cleanup.Stop()

	if rl.config.RequestsPerMinute != tierConfigs[TierModerate].RequestsPerMinute {
		t.Fatalf("expected default moderate tier requests, got %d", rl.config.RequestsPerMinute)
	}
	if rl.config.BurstSize != tierConfigs[TierModerate].BurstSize {
		t.Fatalf("expected default moderate burst, got %d", rl.config.BurstSize)
	}
}

func TestRateLimiter_AllowAndRemainingTokens(t *testing.T) {
	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		config:  RateLimitConfig{RequestsPerMinute: 60, BurstSize: 2},
		cleanup: time.NewTicker(time.Hour),
	}
	defer rl.cleanup.Stop()

	if !rl.Allow("1.1.1.1") {
		t.Fatal("first request should be allowed")
	}
	if !rl.Allow("1.1.1.1") {
		t.Fatal("second request should be allowed")
	}
	if rl.Allow("1.1.1.1") {
		t.Fatal("third request should be rate-limited")
	}

	remaining := rl.RemainingTokens("1.1.1.1")
	if remaining != 0 {
		t.Fatalf("expected 0 remaining tokens, got %d", remaining)
	}

	if unknown := rl.RemainingTokens("2.2.2.2"); unknown != 2 {
		t.Fatalf("expected full burst for unknown client, got %d", unknown)
	}
}

func TestRateLimiter_CleanupLoopRemovesStaleClients(t *testing.T) {
	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		config:  RateLimitConfig{RequestsPerMinute: 60, BurstSize: 2},
		cleanup: time.NewTicker(5 * time.Millisecond),
	}
	defer rl.cleanup.Stop()

	rl.clients["stale"] = &clientInfo{tokens: 1, lastAccess: time.Now().Add(-11 * time.Minute)}
	rl.clients["fresh"] = &clientInfo{tokens: 1, lastAccess: time.Now()}

	go rl.cleanupLoop()
	time.Sleep(20 * time.Millisecond)

	rl.mu.RLock()
	_, staleExists := rl.clients["stale"]
	_, freshExists := rl.clients["fresh"]
	rl.mu.RUnlock()

	if staleExists {
		t.Fatal("expected stale client to be removed")
	}
	if !freshExists {
		t.Fatal("expected fresh client to remain")
	}
}

func TestGetClientIP_HeaderPriority(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx1, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx1.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx1.Request.Header.Set("X-Forwarded-For", "10.0.0.1")
	if ip := getClientIP(ctx1); ip != "10.0.0.1" {
		t.Fatalf("expected X-Forwarded-For, got %s", ip)
	}

	ctx2, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx2.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx2.Request.Header.Set("X-Real-IP", "10.0.0.2")
	if ip := getClientIP(ctx2); ip != "10.0.0.2" {
		t.Fatalf("expected X-Real-IP, got %s", ip)
	}

	ctx3, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx3.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx3.Request.Header.Set("CF-Connecting-IP", "10.0.0.3")
	if ip := getClientIP(ctx3); ip != "10.0.0.3" {
		t.Fatalf("expected CF-Connecting-IP, got %s", ip)
	}
}

func TestRateLimitMiddleware_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	strictLimiter = nil
	moderateLimiter = nil
	relaxedLimiter = nil
	openLimiter = nil
	initOnce = sync.Once{}

	router := gin.New()
	router.Use(RateLimitMiddleware(TierStrict))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for request %d, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for third request, got %d", w.Code)
	}
	if w.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected remaining header 0, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
