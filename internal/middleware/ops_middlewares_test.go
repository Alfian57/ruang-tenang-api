package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func resetRateLimitGlobals() {
	if strictLimiter != nil {
		strictLimiter.cleanup.Stop()
	}
	if moderateLimiter != nil {
		moderateLimiter.cleanup.Stop()
	}
	if relaxedLimiter != nil {
		relaxedLimiter.cleanup.Stop()
	}
	if openLimiter != nil {
		openLimiter.cleanup.Stop()
	}
	strictLimiter = nil
	moderateLimiter = nil
	relaxedLimiter = nil
	openLimiter = nil
	initOnce = sync.Once{}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware(&config.Config{CORSAllowedOrigins: []string{"http://localhost:3000"}}))
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/ok", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("expected CORS allow origin header")
	}
}

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()

	r := gin.New()
	r.Use(LoggerMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Set("requestId", "req-1")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping?x=1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiterAndMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetRateLimitGlobals()

	rl := NewRateLimiter("unknown")
	defer rl.cleanup.Stop()
	if rl.config.RequestsPerMinute != tierConfigs[TierModerate].RequestsPerMinute {
		t.Fatalf("expected default moderate config, got %+v", rl.config)
	}

	if !rl.Allow("1.1.1.1") {
		t.Fatal("expected first request allowed")
	}
	_ = rl.RemainingTokens("1.1.1.1")

	r := gin.New()
	r.Use(StrictRateLimit())
	r.GET("/limited", func(c *gin.Context) { c.Status(http.StatusOK) })

	first := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req1.Header.Set("X-Forwarded-For", "9.9.9.9")
	r.ServeHTTP(first, req1)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first strict request 200, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req2.Header.Set("X-Forwarded-For", "9.9.9.9")
	r.ServeHTTP(second, req2)

	third := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req3.Header.Set("X-Forwarded-For", "9.9.9.9")
	r.ServeHTTP(third, req3)

	if third.Code != http.StatusTooManyRequests {
		t.Fatalf("expected strict limiter to block third request, got %d", third.Code)
	}
	if third.Header().Get("X-RateLimit-Limit") == "" || third.Header().Get("Retry-After") == "" {
		t.Fatal("expected rate limit headers")
	}

	if formatInt(123) != "123" {
		t.Fatal("formatInt should stringify ints")
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	reqIP := httptest.NewRequest(http.MethodGet, "/", nil)
	reqIP.Header.Set("X-Real-IP", "2.2.2.2")
	c.Request = reqIP
	if got := getClientIP(c); got != "2.2.2.2" {
		t.Fatalf("expected X-Real-IP precedence, got %s", got)
	}

	_ = ModerateRateLimit()
	_ = RelaxedRateLimit()
	_ = OpenRateLimit()
}
