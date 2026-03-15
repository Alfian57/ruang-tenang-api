package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Global Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware(cfg))
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.InputValidationMiddleware())
	r.Use(middleware.MaxBodySizeMiddleware(10 * 1024 * 1024)) // 10MB max body size

	r.Static("/uploads", "./uploads")
	r.Static("/storage", "./storage")

	deps := initializeRouteDependencies(cfg)

	registerBaseRoutes(r, deps)
	registerAPIV1Routes(r, deps)

	return r
}
