package middleware

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	origins := []string{
		"http://ruang-tenang.site",
		"https://ruang-tenang.site",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}

	if cfg.ClientOrigin != "" {
		origins = append(origins, cfg.ClientOrigin)
	}

	config := cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}
