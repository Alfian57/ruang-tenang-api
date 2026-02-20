package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerBaseRoutes(r *gin.Engine, deps *routeDependencies) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if gin.Mode() != gin.ReleaseMode {
		r.POST("/dev/cache/clear", func(c *gin.Context) {
			deps.cacheService.Clear()
			c.JSON(200, gin.H{"status": "ok", "message": "Cache cleared"})
		})
	}

	r.GET("/api/v1/leaderboard", deps.userHandler.GetLeaderboard)
}
