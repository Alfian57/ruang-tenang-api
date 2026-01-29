package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/handlers"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Global Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware(cfg))
	r.Use(middleware.InputValidationMiddleware())
	r.Use(middleware.MaxBodySizeMiddleware(10 * 1024 * 1024)) // 10MB max body size

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
	forumRepo := repositories.NewForumRepository(db)
	forumCategoryRepo := repositories.NewForumCategoryRepository(db)
	levelConfigRepo := repositories.NewLevelConfigRepository(db)
	expHistoryRepo := repositories.NewExpHistoryRepository(db)
	moderationRepo := repositories.NewModerationRepository(db)
	communityProgressRepo := repositories.NewCommunityProgressRepository(db)
	featureUnlockRepo := repositories.NewFeatureUnlockRepository(db)
	badgeRepo := repositories.NewBadgeRepository(db)
	inspiringStoryRepo := repositories.NewInspiringStoryRepository(db)
	breathingRepo := repositories.NewBreathingRepository(db)
	playlistRepo := repositories.NewPlaylistRepository(db)
	playlistItemRepo := repositories.NewPlaylistItemRepository(db)

	// Services
	cacheService := services.NewCacheService()
	gamificationService := services.NewGamificationService(db)
	contentContextService := services.NewContentContextService(articleRepo, songRepo, songCategoryRepo, forumRepo)
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	articleService := services.NewArticleService(articleRepo, articleCategoryRepo, gamificationService, contentContextService, cacheService)
	songService := services.NewSongService(songRepo, songCategoryRepo, cacheService)
	moodService := services.NewMoodService(moodRepo)
	forumService := services.NewForumService(forumRepo, gamificationService, contentContextService)
	forumCategoryService := services.NewForumCategoryService(forumCategoryRepo, cacheService)
	levelConfigService := services.NewLevelConfigService(levelConfigRepo, cacheService)
	expHistoryService := services.NewExpHistoryService(expHistoryRepo)
	chatService := services.NewChatService(chatSessionRepo, chatMessageRepo, cfg, gamificationService, contentContextService)
	aiModerationService := services.NewAIModerationService(moderationRepo, cfg)
	moderationService := services.NewModerationService(moderationRepo, userRepo, articleRepo, forumRepo, aiModerationService)
	communityProgressService := services.NewCommunityProgressService(communityProgressRepo, levelConfigRepo, featureUnlockRepo, badgeRepo, userRepo)
	featureUnlockService := services.NewFeatureUnlockService(featureUnlockRepo, levelConfigRepo, userRepo)
	badgeService := services.NewBadgeService(badgeRepo, userRepo, levelConfigRepo)
	inspiringStoryService := services.NewInspiringStoryService(inspiringStoryRepo, userRepo, levelConfigRepo, badgeService)
	breathingService := services.NewBreathingService(breathingRepo, gamificationService)
	playlistService := services.NewPlaylistService(playlistRepo, playlistItemRepo, songRepo)

	// Inject moderation repository into chat service for crisis detection
	chatService.SetModerationRepo(moderationRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, levelConfigService)
	userHandler := handlers.NewUserHandler(userService, levelConfigService)
	articleHandler := handlers.NewArticleHandler(articleService)
	chatHandler := handlers.NewChatHandler(chatService)
	uploadHandler := handlers.NewUploadHandler()
	songHandler := handlers.NewSongHandler(songService)
	moodHandler := handlers.NewMoodHandler(moodService)
	adminHandler := handlers.NewAdminHandler(db, userRepo, articleRepo)
	searchHandler := handlers.NewSearchHandler(articleRepo, songRepo)
	forumHandler := handlers.NewForumHandler(forumService)
	forumCategoryHandler := handlers.NewForumCategoryHandler(forumCategoryService)
	levelConfigHandler := handlers.NewLevelConfigHandler(levelConfigService)
	expHistoryHandler := handlers.NewExpHistoryHandler(expHistoryService, levelConfigService)
	moderationHandler := handlers.NewModerationHandler(moderationService)
	communityProgressHandler := handlers.NewCommunityProgressHandler(communityProgressService)
	featureUnlockHandler := handlers.NewFeatureUnlockHandler(featureUnlockService)
	badgeHandler := handlers.NewBadgeHandler(badgeService)
	inspiringStoryHandler := handlers.NewInspiringStoryHandler(inspiringStoryService)
	breathingHandler := handlers.NewBreathingHandler(breathingService)
	playlistHandler := handlers.NewPlaylistHandler(playlistService)

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Leaderboard (public)
	r.GET("/api/v1/leaderboard", userHandler.GetLeaderboard)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public) - STRICT rate limiting
		auth := v1.Group("/auth")
		auth.Use(middleware.StrictRateLimit())
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected auth routes - RELAXED rate limiting
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware())
		authProtected.Use(middleware.RelaxedRateLimit())
		{
			authProtected.GET("/me", authHandler.GetProfile)
			authProtected.PUT("/profile", authHandler.UpdateProfile)
			authProtected.PUT("/password", authHandler.UpdatePassword)
		}

		// Upload routes (protected) - MODERATE rate limiting
		upload := v1.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		upload.Use(middleware.ModerateRateLimit())
		{
			upload.POST("/image", uploadHandler.UploadImage)
			upload.POST("/audio", uploadHandler.UploadAudio)
		}

		// Articles (public) - OPEN rate limiting
		articles := v1.Group("/articles")
		articles.Use(middleware.OpenRateLimit())
		{
			articles.GET("", articleHandler.GetArticles)
			articles.GET("/:id", articleHandler.GetArticle)
		}
		v1.GET("/article-categories", middleware.OpenRateLimit(), articleHandler.GetCategories)

		// User articles (protected) - MODERATE rate limiting for writes
		myArticles := v1.Group("/my-articles")
		myArticles.Use(middleware.AuthMiddleware())
		myArticles.Use(middleware.ModerateRateLimit())
		{
			myArticles.GET("", articleHandler.GetMyArticles)
			myArticles.POST("", articleHandler.CreateMyArticle)
			myArticles.GET("/:id", articleHandler.GetArticleByIDForUser)
			myArticles.PUT("/:id", articleHandler.UpdateMyArticle)
			myArticles.DELETE("/:id", articleHandler.DeleteMyArticle)
		}

		// Songs (public) - OPEN rate limiting
		v1.GET("/song-categories", middleware.OpenRateLimit(), songHandler.GetCategories)
		v1.GET("/song-categories/:id/songs", middleware.OpenRateLimit(), songHandler.GetSongsByCategory)
		v1.GET("/songs/:id", middleware.OpenRateLimit(), songHandler.GetSong)

		// Public playlists
		v1.GET("/playlists/public", middleware.OpenRateLimit(), playlistHandler.GetPublicPlaylists)

		// Playlists (protected) - RELAXED rate limiting
		playlists := v1.Group("/playlists")
		playlists.Use(middleware.AuthMiddleware())
		playlists.Use(middleware.RelaxedRateLimit())
		{
			playlists.GET("", playlistHandler.GetMyPlaylists)
			playlists.POST("", playlistHandler.CreatePlaylist)
			playlists.GET("/:id", playlistHandler.GetPlaylist)
			playlists.PUT("/:id", playlistHandler.UpdatePlaylist)
			playlists.DELETE("/:id", playlistHandler.DeletePlaylist)
			playlists.POST("/:id/songs", playlistHandler.AddSongToPlaylist)
			playlists.POST("/:id/songs/batch", playlistHandler.AddSongsToPlaylist)
			playlists.DELETE("/:id/songs/:songId", playlistHandler.RemoveSongFromPlaylist)
			playlists.DELETE("/:id/items/:itemId", playlistHandler.RemoveItemFromPlaylist)
			playlists.PUT("/:id/reorder", playlistHandler.ReorderPlaylistItems)
		}

		// Chat (protected) - MODERATE rate limiting for messages
		chat := v1.Group("/chat-sessions")
		chat.Use(middleware.AuthMiddleware())
		chat.Use(middleware.ModerateRateLimit())
		{
			chat.GET("", chatHandler.GetSessions)
			chat.POST("", chatHandler.CreateSession)
			chat.GET("/:id", chatHandler.GetSession) // Changed from GetSession to GetSessionByID
			chat.POST("/:id/messages", chatHandler.SendMessage)
			chat.PUT("/:id/trash", chatHandler.ToggleTrash)
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

		// Level configs (public)
		v1.GET("/level-configs", levelConfigHandler.GetAllConfigs)

		// EXP History (protected)
		expHistory := v1.Group("/exp-history")
		expHistory.Use(middleware.AuthMiddleware())
		{
			expHistory.GET("", expHistoryHandler.GetHistory)
			expHistory.GET("/activity-types", expHistoryHandler.GetActivityTypes)
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
			admin.PUT("/song-categories/:id", adminHandler.UpdateSongCategory)
			admin.DELETE("/song-categories/:id", adminHandler.DeleteSongCategory)

			// Song management
			admin.GET("/songs", adminHandler.GetAllSongs)
			admin.POST("/songs", adminHandler.CreateSong)
			admin.PUT("/songs/:id", adminHandler.UpdateSong)
			admin.DELETE("/songs/:id", adminHandler.DeleteSong)

			// Forum category management
			admin.GET("/forum-categories", forumCategoryHandler.AdminGetAllCategories)
			admin.POST("/forum-categories", forumCategoryHandler.CreateCategory)
			admin.PUT("/forum-categories/:id", forumCategoryHandler.UpdateCategory)
			admin.DELETE("/forum-categories/:id", forumCategoryHandler.DeleteCategory)

			// Level config management
			admin.GET("/level-configs", levelConfigHandler.AdminGetAllConfigs)
			admin.POST("/level-configs", levelConfigHandler.CreateConfig)
			admin.PUT("/level-configs/:id", levelConfigHandler.UpdateConfig)
			admin.DELETE("/level-configs/:id", levelConfigHandler.DeleteConfig)
		}

		// Public Forum Categories
		v1.GET("/forum-categories", forumCategoryHandler.GetAllCategories)

		// Forum (protected)
		forum := v1.Group("/forums")
		forum.Use(middleware.AuthMiddleware())
		{
			forum.POST("", forumHandler.CreateForum)
			forum.GET("", forumHandler.GetForums)
			forum.GET("/:id", forumHandler.GetForumByID)
			forum.DELETE("/:id", forumHandler.DeleteForum)
			forum.POST("/:id/posts", forumHandler.CreateForumPost)
			forum.GET("/:id/posts", forumHandler.GetForumPosts)
			forum.PUT("/:id/like", forumHandler.ToggleLike)
		}

		// Forum Posts (protected)
		posts := v1.Group("/posts")
		posts.Use(middleware.AuthMiddleware())
		{
			posts.DELETE("/:id", forumHandler.DeleteForumPost)
		}

		// Moderation routes (protected, moderator or admin only)
		moderation := v1.Group("/moderation")
		moderation.Use(middleware.AuthMiddleware())
		moderation.Use(middleware.ModeratorMiddleware())
		{
			// Dashboard & Stats
			moderation.GET("/stats", moderationHandler.GetModerationStats)
			moderation.GET("/queue", moderationHandler.GetModerationQueue)

			// Article moderation
			moderation.PUT("/articles/:id", moderationHandler.ModerateArticle)

			// Report management
			moderation.GET("/reports", moderationHandler.GetReports)
			moderation.PUT("/reports/:id", moderationHandler.HandleReport)

			// User strikes
			moderation.GET("/users/:id/strikes", moderationHandler.GetUserStrikes)

			// Trigger warnings
			moderation.POST("/trigger-warnings", moderationHandler.AddTriggerWarnings)

			// Moderator action logs
			moderation.GET("/actions", moderationHandler.GetModeratorActions)

			// Crisis keywords management
			moderation.GET("/crisis-keywords", moderationHandler.GetCrisisKeywords)
			moderation.POST("/crisis-keywords", moderationHandler.CreateCrisisKeyword)
			moderation.DELETE("/crisis-keywords/:id", moderationHandler.DeleteCrisisKeyword)
		}

		// User reports (protected, any authenticated user)
		reports := v1.Group("/reports")
		reports.Use(middleware.AuthMiddleware())
		{
			reports.POST("", moderationHandler.CreateReport)
		}

		// User blocking (protected, any authenticated user)
		blocks := v1.Group("/blocks")
		blocks.Use(middleware.AuthMiddleware())
		{
			blocks.GET("", moderationHandler.GetBlockedUsers)
			blocks.POST("", moderationHandler.BlockUser)
			blocks.DELETE("/:id", moderationHandler.UnblockUser)
		}

		// User settings for moderation/safety (protected)
		userSettings := v1.Group("/user")
		userSettings.Use(middleware.AuthMiddleware())
		{
			userSettings.POST("/accept-ai-disclaimer", moderationHandler.AcceptAIDisclaimer)
			userSettings.PUT("/content-warning-preference", moderationHandler.UpdateContentWarningPreference)
		}

		// Search
		v1.GET("/search", searchHandler.Search)

		// ==========================================
		// Community Progress Routes
		// ==========================================

		// Public community routes
		v1.GET("/community/stats", middleware.OpenRateLimit(), communityProgressHandler.GetCommunityStats)
		v1.GET("/community/hall-of-fame/level/:level", middleware.OpenRateLimit(), communityProgressHandler.GetLevelHallOfFame)
		v1.GET("/community/hall-of-fame/monthly", middleware.OpenRateLimit(), communityProgressHandler.GetMonthlyHallOfFame)
		v1.GET("/community/hall-of-fame/categories", middleware.OpenRateLimit(), communityProgressHandler.GetHallOfFameCategories)

		// Protected community routes
		community := v1.Group("/community")
		community.Use(middleware.AuthMiddleware())
		community.Use(middleware.RelaxedRateLimit())
		{
			community.GET("/my-journey", communityProgressHandler.GetPersonalJourney)
			community.GET("/my-progress/weekly", communityProgressHandler.GetWeeklyProgress)
			community.GET("/my-progress/monthly", communityProgressHandler.GetMonthlyProgress)
			community.GET("/my-stats", communityProgressHandler.GetAllTimeStats)
			community.GET("/celebrate/:level", communityProgressHandler.GetLevelUpCelebration)
		}

		// ==========================================
		// Feature Unlock Routes
		// ==========================================

		// Public features routes
		v1.GET("/features", middleware.OpenRateLimit(), featureUnlockHandler.GetAllFeatures)
		v1.GET("/features/categories", middleware.OpenRateLimit(), featureUnlockHandler.GetFeatureCategories)
		v1.GET("/features/category/:category", middleware.OpenRateLimit(), featureUnlockHandler.GetFeaturesByCategory)

		// Protected features routes
		features := v1.Group("/features")
		features.Use(middleware.AuthMiddleware())
		features.Use(middleware.RelaxedRateLimit())
		{
			features.GET("/my-features", featureUnlockHandler.GetUserFeatures)
			features.GET("/check/:featureKey", featureUnlockHandler.CheckFeatureAccess)
			features.GET("/upcoming", featureUnlockHandler.GetUpcomingFeatures)
		}

		// ==========================================
		// Badge Routes
		// ==========================================

		// Public badge routes
		v1.GET("/badges", middleware.OpenRateLimit(), badgeHandler.GetAllBadges)
		v1.GET("/badges/categories", middleware.OpenRateLimit(), badgeHandler.GetBadgeCategories)
		v1.GET("/badges/category/:category", middleware.OpenRateLimit(), badgeHandler.GetBadgesByCategory)

		// Protected badge routes
		badges := v1.Group("/badges")
		badges.Use(middleware.AuthMiddleware())
		badges.Use(middleware.RelaxedRateLimit())
		{
			badges.GET("/my-badges", badgeHandler.GetUserBadges)
			badges.GET("/progress", badgeHandler.GetBadgeProgress)
			badges.GET("/recent", badgeHandler.GetRecentlyEarnedBadges)
			badges.GET("/display", badgeHandler.GetDisplayBadges)
			badges.POST("/check", badgeHandler.CheckNewBadges)
		}

		// ==========================================
		// Inspiring Stories Routes
		// ==========================================

		// Public stories routes
		v1.GET("/stories/categories", middleware.OpenRateLimit(), inspiringStoryHandler.GetCategories)
		v1.GET("/stories/featured", middleware.OpenRateLimit(), inspiringStoryHandler.GetFeaturedStories)
		v1.GET("/stories/most-appreciated", middleware.OpenRateLimit(), inspiringStoryHandler.GetMostAppreciated)
		v1.GET("/stories", middleware.OpenRateLimit(), inspiringStoryHandler.GetStories)
		v1.GET("/stories/:id", middleware.OpenRateLimit(), inspiringStoryHandler.GetStory)
		v1.GET("/stories/:id/comments", middleware.OpenRateLimit(), inspiringStoryHandler.GetComments)

		// Protected stories routes
		stories := v1.Group("/stories")
		stories.Use(middleware.AuthMiddleware())
		stories.Use(middleware.ModerateRateLimit())
		{
			stories.POST("", inspiringStoryHandler.CreateStory)
			stories.PUT("/:id", inspiringStoryHandler.UpdateStory)
			stories.DELETE("/:id", inspiringStoryHandler.DeleteStory)
			stories.GET("/my-stories", inspiringStoryHandler.GetMyStories)
			stories.GET("/my-stats", inspiringStoryHandler.GetMyStats)
			stories.POST("/:id/heart", inspiringStoryHandler.ToggleHeart)
			stories.POST("/:id/comments", inspiringStoryHandler.CreateComment)
			stories.DELETE("/:id/comments/:commentId", inspiringStoryHandler.DeleteComment)
			stories.POST("/:id/comments/:commentId/heart", inspiringStoryHandler.ToggleCommentHeart)
		}

		// Admin story moderation routes
		adminStories := v1.Group("/admin/stories")
		adminStories.Use(middleware.AuthMiddleware())
		adminStories.Use(middleware.ModeratorMiddleware())
		{
			adminStories.GET("/pending", inspiringStoryHandler.GetPendingStories)
			adminStories.POST("/:id/moderate", inspiringStoryHandler.ModerateStory)
			adminStories.POST("/:id/featured", inspiringStoryHandler.SetFeatured)
			adminStories.POST("/:id/comments/:commentId/hide", inspiringStoryHandler.HideComment)
		}

		// ==========================================
		// Breathing Exercise Routes
		// ==========================================

		// Protected breathing routes
		breathing := v1.Group("/breathing")
		breathing.Use(middleware.AuthMiddleware())
		breathing.Use(middleware.RelaxedRateLimit())
		{
			// Techniques
			breathing.GET("/techniques", breathingHandler.GetTechniques)
			breathing.GET("/techniques/:id", breathingHandler.GetTechniqueByID)
			breathing.GET("/techniques/slug/:slug", breathingHandler.GetTechniqueBySlug)
			breathing.POST("/techniques", breathingHandler.CreateTechnique)
			breathing.PUT("/techniques/:id", breathingHandler.UpdateTechnique)
			breathing.DELETE("/techniques/:id", breathingHandler.DeleteTechnique)

			// Sessions
			breathing.GET("/sessions", breathingHandler.GetSessionHistory)
			breathing.POST("/sessions", breathingHandler.StartSession)
			breathing.GET("/sessions/:id", breathingHandler.GetSessionByID)
			breathing.POST("/sessions/:id/complete", breathingHandler.CompleteSession)

			// Preferences
			breathing.GET("/preferences", breathingHandler.GetPreferences)
			breathing.PUT("/preferences", breathingHandler.UpdatePreferences)

			// Favorites
			breathing.GET("/favorites", breathingHandler.GetFavorites)
			breathing.POST("/favorites/:id", breathingHandler.AddFavorite)
			breathing.DELETE("/favorites/:id", breathingHandler.RemoveFavorite)
			breathing.PUT("/favorites/reorder", breathingHandler.ReorderFavorites)

			// Stats & Calendar
			breathing.GET("/stats", breathingHandler.GetStats)
			breathing.GET("/stats/usage", breathingHandler.GetTechniqueUsage)
			breathing.GET("/calendar", breathingHandler.GetCalendar)

			// Widget & Recommendations
			breathing.GET("/widget", breathingHandler.GetWidgetData)
			breathing.GET("/recommendations", breathingHandler.GetRecommendations)
		}
	}

	return r
}
