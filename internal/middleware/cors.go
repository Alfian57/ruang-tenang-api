package middleware

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Authorization", "Accept",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		MaxAge: 12 * time.Hour,
	}

	// In development, allow all origins (needed for mobile apps & emulators)
	if cfg.AppEnv == "development" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false // AllowAllOrigins + AllowCredentials is not allowed
	} else {
		corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
		corsConfig.AllowCredentials = true
	}

	return cors.New(corsConfig)
}
