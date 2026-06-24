package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"
)

// RateLimitTier defines different rate limiting tiers
type RateLimitTier string

const (
	// TierStrict - 30 requests per minute (auth endpoints, password reset)
	TierStrict RateLimitTier = "strict"
	// TierModerate - 180 requests per minute (write operations: POST, PUT, DELETE)
	TierModerate RateLimitTier = "moderate"
	// TierRelaxed - 600 requests per minute (read operations: GET authenticated)
	TierRelaxed RateLimitTier = "relaxed"
	// TierOpen - 1200 requests per minute (public endpoints)
	TierOpen RateLimitTier = "open"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// Rate limit configurations for each tier
var tierConfigs = map[RateLimitTier]RateLimitConfig{
	TierStrict:   {RequestsPerMinute: 30, BurstSize: 10},
	TierModerate: {RequestsPerMinute: 180, BurstSize: 60},
	TierRelaxed:  {RequestsPerMinute: 600, BurstSize: 120},
	TierOpen:     {RequestsPerMinute: 1200, BurstSize: 240},
}

// clientInfo stores rate limiting data for a client
type clientInfo struct {
	tokens     float64
	lastAccess time.Time
	mu         sync.Mutex
}

// RateLimiter manages rate limiting for all clients
type RateLimiter struct {
	clients map[string]*clientInfo
	mu      sync.RWMutex
	config  RateLimitConfig
	cleanup *time.Ticker
}

// NewRateLimiter creates a new rate limiter with the specified tier
func NewRateLimiter(tier RateLimitTier) *RateLimiter {
	config := tierConfigs[tier]
	if config.RequestsPerMinute == 0 {
		config = tierConfigs[TierModerate] // default to moderate
	}

	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		config:  config,
		cleanup: time.NewTicker(5 * time.Minute),
	}

	// Start cleanup goroutine to remove stale clients
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop removes inactive clients to prevent memory leaks
func (rl *RateLimiter) cleanupLoop() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, client := range rl.clients {
			client.mu.Lock()
			if client.lastAccess.Before(cutoff) {
				delete(rl.clients, ip)
			}
			client.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	client, exists := rl.clients[ip]
	if !exists {
		client = &clientInfo{
			tokens:     float64(rl.config.BurstSize),
			lastAccess: time.Now(),
		}
		rl.clients[ip] = client
	}
	rl.mu.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(client.lastAccess).Seconds()
	client.lastAccess = now

	// Token bucket algorithm: refill tokens based on time elapsed
	tokensPerSecond := float64(rl.config.RequestsPerMinute) / 60.0
	client.tokens += elapsed * tokensPerSecond
	if client.tokens > float64(rl.config.BurstSize) {
		client.tokens = float64(rl.config.BurstSize)
	}

	// Check if request can be allowed
	if client.tokens >= 1 {
		client.tokens--
		return true
	}

	return false
}

// RemainingTokens returns the number of remaining requests for an IP
func (rl *RateLimiter) RemainingTokens(ip string) int {
	rl.mu.RLock()
	client, exists := rl.clients[ip]
	rl.mu.RUnlock()

	if !exists {
		return rl.config.BurstSize
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	return int(client.tokens)
}

// Global rate limiters for different tiers
var (
	strictLimiter   *RateLimiter
	moderateLimiter *RateLimiter
	relaxedLimiter  *RateLimiter
	openLimiter     *RateLimiter
	initOnce        sync.Once
)

// initLimiters initializes all rate limiters
func initLimiters() {
	initOnce.Do(func() {
		strictLimiter = NewRateLimiter(TierStrict)
		moderateLimiter = NewRateLimiter(TierModerate)
		relaxedLimiter = NewRateLimiter(TierRelaxed)
		openLimiter = NewRateLimiter(TierOpen)
	})
}

// RateLimitMiddleware creates a rate limiting middleware for the specified tier
func RateLimitMiddleware(tier RateLimitTier) gin.HandlerFunc {
	initLimiters()

	var limiter *RateLimiter
	switch tier {
	case TierStrict:
		limiter = strictLimiter
	case TierModerate:
		limiter = moderateLimiter
	case TierRelaxed:
		limiter = relaxedLimiter
	case TierOpen:
		limiter = openLimiter
	default:
		limiter = moderateLimiter
	}

	return func(c *gin.Context) {
		ip := getClientIP(c)

		if !limiter.Allow(ip) {
			c.Header("X-RateLimit-Limit", formatInt(tierConfigs[tier].RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", formatInt(int(time.Now().Add(time.Minute).Unix())))
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, dto.ErrorResponseWithCode(
				"ERR_RATE_LIMIT",
				"Too many requests. Please try again later.",
			))
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", formatInt(tierConfigs[tier].RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", formatInt(limiter.RemainingTokens(ip)))

		c.Next()
	}
}

// StrictRateLimit returns middleware for strict rate limiting (auth endpoints)
func StrictRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(TierStrict)
}

// ModerateRateLimit returns middleware for moderate rate limiting (write operations)
func ModerateRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(TierModerate)
}

// RelaxedRateLimit returns middleware for relaxed rate limiting (read operations)
func RelaxedRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(TierRelaxed)
}

// OpenRateLimit returns middleware for open rate limiting (public endpoints)
func OpenRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(TierOpen)
}

// getClientIP extracts the real client IP from request
func getClientIP(c *gin.Context) string {
	// X-Forwarded-For may contain a comma-separated chain: "client, proxy1, proxy2".
	// The left-most entry is the original client. Returning the whole header value
	// (as before) made every request share one bucket when a proxy appends IPs, or
	// let an attacker spoof a unique IP per request to bypass rate limiting.
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		if ip := firstIP(xff); ip != "" {
			return ip
		}
	}
	if ip := strings.TrimSpace(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); ip != "" {
		return ip
	}
	return c.ClientIP()
}

// firstIP extracts the left-most IP from a comma-separated X-Forwarded-For value.
func firstIP(xff string) string {
	for i := 0; i < len(xff); i++ {
		if xff[i] == ',' {
			return strings.TrimSpace(xff[:i])
		}
	}
	return strings.TrimSpace(xff)
}

// formatInt converts int to string for headers
func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}
