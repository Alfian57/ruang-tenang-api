package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/handlers"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	// Serve static files for uploads
	r.Static("/uploads", "./uploads")

	// Initialize dependencies
	db := database.GetDB()

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	articleCategoryRepo := repositories.NewArticleCategoryRepository(db)
	chatSessionRepo := repositories.NewChatSessionRepository(db)
	chatMessageRepo := repositories.NewChatMessageRepository(db)
	songRepo := repositories.NewSongRepository(db)
	songCategoryRepo := repositories.NewSongCategoryRepository(db)
	moodRepo := repositories.NewUserMoodRepository(db)

	// Services
	authService := services.NewAuthService(userRepo)
	articleService := services.NewArticleService(articleRepo, articleCategoryRepo)
	chatService := services.NewChatService(chatSessionRepo, chatMessageRepo)
	songService := services.NewSongService(songRepo, songCategoryRepo)
	moodService := services.NewMoodService(moodRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	articleHandler := handlers.NewArticleHandler(articleService)
	chatHandler := handlers.NewChatHandler(chatService)
	songHandler := handlers.NewSongHandler(songService)
	moodHandler := handlers.NewMoodHandler(moodService)
	adminHandler := handlers.NewAdminHandler(db, userRepo, articleRepo)
	uploadHandler := handlers.NewUploadHandler()

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected auth routes
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware())
		{
			authProtected.GET("/me", authHandler.GetProfile)
			authProtected.PUT("/profile", authHandler.UpdateProfile)
			authProtected.PUT("/password", authHandler.UpdatePassword)
		}

		// Upload routes (protected)
		upload := v1.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		{
			upload.POST("/image", uploadHandler.UploadImage)
			upload.POST("/audio", uploadHandler.UploadAudio)
		}

		// Articles (public)
		articles := v1.Group("/articles")
		{
			articles.GET("", articleHandler.GetArticles)
			articles.GET("/:id", articleHandler.GetArticle)
		}
		v1.GET("/article-categories", articleHandler.GetCategories)

		// User articles (protected) - for users to manage their own articles
		myArticles := v1.Group("/my-articles")
		myArticles.Use(middleware.AuthMiddleware())
		{
			myArticles.GET("", articleHandler.GetMyArticles)
			myArticles.POST("", articleHandler.CreateMyArticle)
			myArticles.GET("/:id", articleHandler.GetArticleByIDForUser)
			myArticles.PUT("/:id", articleHandler.UpdateMyArticle)
			myArticles.DELETE("/:id", articleHandler.DeleteMyArticle)
		}

		// Songs (public)
		v1.GET("/song-categories", songHandler.GetCategories)
		v1.GET("/song-categories/:id/songs", songHandler.GetSongsByCategory)
		v1.GET("/songs/:id", songHandler.GetSong)

		// Chat (protected)
		chat := v1.Group("/chat-sessions")
		chat.Use(middleware.AuthMiddleware())
		{
			chat.GET("", chatHandler.GetSessions)
			chat.POST("", chatHandler.CreateSession)
			chat.GET("/:id", chatHandler.GetSession)
			chat.POST("/:id/messages", chatHandler.SendMessage)
			chat.PUT("/:id/bookmark", chatHandler.ToggleBookmark)
			chat.PUT("/:id/favorite", chatHandler.ToggleFavorite)
			chat.DELETE("/:id", chatHandler.DeleteSession)
		}

		// Chat messages (protected)
		chatMessages := v1.Group("/chat-messages")
		chatMessages.Use(middleware.AuthMiddleware())
		{
			chatMessages.PUT("/:id/like", chatHandler.ToggleMessageLike)
			chatMessages.PUT("/:id/dislike", chatHandler.ToggleMessageDislike)
		}

		// Mood (protected)
		mood := v1.Group("/user-moods")
		mood.Use(middleware.AuthMiddleware())
		{
			mood.GET("", moodHandler.GetMoodHistory)
			mood.POST("", moodHandler.RecordMood)
			mood.GET("/latest", moodHandler.GetLatestMood)
			mood.GET("/stats", moodHandler.GetMoodStats)
		}

		// Admin routes (protected, admin only)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/stats", adminHandler.GetDashboardStats)

			// User management
			admin.GET("/users", adminHandler.GetUsers)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.PUT("/users/:id/block", adminHandler.BlockUser)
			admin.PUT("/users/:id/unblock", adminHandler.UnblockUser)

			// Article management
			admin.GET("/articles", adminHandler.GetAllArticles)
			admin.POST("/articles", adminHandler.CreateArticle)
			admin.PUT("/articles/:id", adminHandler.UpdateArticle)
			admin.DELETE("/articles/:id", adminHandler.DeleteArticle)
			admin.PUT("/articles/:id/block", adminHandler.BlockArticle)
			admin.PUT("/articles/:id/unblock", adminHandler.UnblockArticle)

			// Article category management
			admin.GET("/article-categories", adminHandler.GetArticleCategories)
			admin.POST("/article-categories", adminHandler.CreateArticleCategory)
			admin.PUT("/article-categories/:id", adminHandler.UpdateArticleCategory)
			admin.DELETE("/article-categories/:id", adminHandler.DeleteArticleCategory)

			// Song category management
			admin.POST("/song-categories", adminHandler.CreateSongCategory)
			admin.DELETE("/song-categories/:id", adminHandler.DeleteSongCategory)

			// Song management
			admin.GET("/songs", adminHandler.GetAllSongs)
			admin.POST("/songs", adminHandler.CreateSong)
			admin.DELETE("/songs/:id", adminHandler.DeleteSong)
		}
	}

	return r
}
