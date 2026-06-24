package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Trust only loopback proxies by default so c.ClientIP() and the rate
	// limiter's proxy-header parsing cannot be spoofed by direct clients.
	// Behind a real reverse proxy/CDN, set TRUSTED_PROXIES to those CIDRs.
	if err := r.SetTrustedProxies([]string{"127.0.0.0/8", "::1"}); err != nil {
		// Non-fatal: gin logs the warning; default trust policy still applies.
		_ = err
	}

	// Global Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware(cfg))
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.InputValidationMiddleware())
	r.Use(middleware.MaxBodySizeMiddleware(10 * 1024 * 1024)) // 10MB max body size

	r.GET("/uploads/*filepath", middleware.SafeStatic("/uploads", "./uploads"))
	r.GET("/storage/*filepath", middleware.SafeStatic("/storage", "./storage"))

	deps := initializeRouteDependencies(cfg)

	registerBaseRoutes(r, deps)
	registerAPIV1Routes(r, deps)

	return r
}
