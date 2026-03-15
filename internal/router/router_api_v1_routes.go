package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAPIV1Routes(r *gin.Engine, deps *routeDependencies) {
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middleware.StrictRateLimit())
		{
			auth.POST("/register", deps.authHandler.Register)
			auth.POST("/login", deps.authHandler.Login)
			auth.POST("/forgot-password", deps.authHandler.ForgotPassword)
			auth.POST("/reset-password", deps.authHandler.ResetPassword)
		}

		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware())
		authProtected.Use(middleware.RelaxedRateLimit())
		{
			authProtected.GET("/me", deps.authHandler.GetProfile)
			authProtected.PUT("/profile", deps.authHandler.UpdateProfile)
			authProtected.PUT("/password", deps.authHandler.UpdatePassword)
		}

		upload := v1.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		upload.Use(middleware.ModerateRateLimit())
		{
			upload.POST("/image", deps.uploadHandler.UploadImage)
			upload.POST("/audio", deps.uploadHandler.UploadAudio)
		}

		articles := v1.Group("/articles")
		articles.Use(middleware.OpenRateLimit())
		{
			articles.GET("", deps.articleHandler.GetArticles)
			articles.GET("/:slug", middleware.OptionalAuthMiddleware(), deps.articleHandler.GetArticle)
		}
		v1.GET("/article-categories", middleware.OpenRateLimit(), deps.articleHandler.GetCategories)

		myArticles := v1.Group("/my-articles")
		myArticles.Use(middleware.AuthMiddleware())
		myArticles.Use(middleware.ModerateRateLimit())
		{
			myArticles.GET("", deps.articleHandler.GetMyArticles)
			myArticles.POST("", deps.articleHandler.CreateMyArticle)
			myArticles.GET("/:slug", deps.articleHandler.GetArticleByIDForUser)
			myArticles.PUT("/:slug", deps.articleHandler.UpdateMyArticle)
			myArticles.DELETE("/:slug", deps.articleHandler.DeleteMyArticle)
		}

		v1.GET("/song-categories", middleware.OpenRateLimit(), deps.songHandler.GetCategories)
		v1.GET("/song-categories/:slug/songs", middleware.OpenRateLimit(), deps.songHandler.GetSongsByCategory)
		v1.GET("/songs/:id", middleware.OpenRateLimit(), deps.songHandler.GetSong)
		v1.GET("/playlists/public", middleware.OpenRateLimit(), deps.playlistHandler.GetPublicPlaylists)

		playlists := v1.Group("/playlists")
		playlists.Use(middleware.AuthMiddleware())
		playlists.Use(middleware.RelaxedRateLimit())
		{
			playlists.GET("", deps.playlistHandler.GetMyPlaylists)
			playlists.POST("", deps.playlistHandler.CreatePlaylist)
			playlists.GET("/:uuid", deps.playlistHandler.GetPlaylist)
			playlists.PUT("/:uuid", deps.playlistHandler.UpdatePlaylist)
			playlists.DELETE("/:uuid", deps.playlistHandler.DeletePlaylist)
			playlists.POST("/:uuid/songs", deps.playlistHandler.AddSongToPlaylist)
			playlists.POST("/:uuid/songs/batch", deps.playlistHandler.AddSongsToPlaylist)
			playlists.DELETE("/:uuid/songs/:songId", deps.playlistHandler.RemoveSongFromPlaylist)
			playlists.DELETE("/:uuid/items/:itemId", deps.playlistHandler.RemoveItemFromPlaylist)
			playlists.PUT("/:uuid/reorder", deps.playlistHandler.ReorderPlaylistItems)
		}

		chat := v1.Group("/chat-sessions")
		chat.Use(middleware.AuthMiddleware())
		chat.Use(middleware.ModerateRateLimit())
		{
			chat.GET("", deps.chatHandler.GetSessions)
			chat.POST("", deps.chatHandler.CreateSession)
			chat.GET("/:uuid", deps.chatHandler.GetSession)
			chat.POST("/:uuid/messages", deps.chatHandler.SendMessage)
			chat.PUT("/:uuid/trash", deps.chatHandler.ToggleTrash)
			chat.PUT("/:uuid/favorite", deps.chatHandler.ToggleFavorite)
			chat.DELETE("/:uuid", deps.chatHandler.DeleteSession)
			chat.PUT("/:uuid/folder", deps.chatHandler.MoveToFolder)
			chat.POST("/:uuid/export", deps.chatHandler.ExportChat)
			chat.GET("/:uuid/summary", deps.chatHandler.GetSummary)
			chat.POST("/:uuid/summary", deps.chatHandler.GenerateSummary)
			chat.GET("/:uuid/pinned", deps.chatHandler.GetPinnedMessages)
		}

		chatMessages := v1.Group("/chat-messages")
		chatMessages.Use(middleware.AuthMiddleware())
		{
			chatMessages.PUT("/:id/like", deps.chatHandler.ToggleMessageLike)
			chatMessages.PUT("/:id/dislike", deps.chatHandler.ToggleMessageDislike)
			chatMessages.PUT("/:id/pin", deps.chatHandler.ToggleMessagePin)
		}

		chatFolders := v1.Group("/chat-folders")
		chatFolders.Use(middleware.AuthMiddleware())
		{
			chatFolders.GET("", deps.chatHandler.GetFolders)
			chatFolders.POST("", deps.chatHandler.CreateFolder)
			chatFolders.PUT("/:id", deps.chatHandler.UpdateFolder)
			chatFolders.DELETE("/:id", deps.chatHandler.DeleteFolder)
			chatFolders.PUT("/reorder", deps.chatHandler.ReorderFolders)
		}

		suggestedPrompts := v1.Group("/chat-prompts")
		suggestedPrompts.Use(middleware.AuthMiddleware())
		{
			suggestedPrompts.GET("", deps.chatHandler.GetSuggestedPrompts)
		}

		journals := v1.Group("/journals")
		journals.Use(middleware.AuthMiddleware())
		journals.Use(middleware.RelaxedRateLimit())
		{
			journals.GET("", deps.journalHandler.ListJournals)
			journals.POST("", deps.journalHandler.CreateJournal)
			journals.GET("/search", deps.journalHandler.SearchJournals)
			journals.GET("/settings", deps.journalHandler.GetSettings)
			journals.PUT("/settings", deps.journalHandler.UpdateSettings)
			journals.GET("/analytics", deps.journalHandler.GetAnalytics)
			journals.GET("/prompt", deps.journalHandler.GetWritingPrompt)
			journals.GET("/weekly-summary", deps.journalHandler.GetWeeklySummary)
			journals.POST("/export", deps.journalHandler.ExportJournals)
			journals.GET("/ai-context", deps.journalHandler.GetAIContext)
			journals.GET("/ai-access-logs", deps.journalHandler.GetAIAccessLogs)
			journals.GET("/:uuid", deps.journalHandler.GetJournal)
			journals.PUT("/:uuid", deps.journalHandler.UpdateJournal)
			journals.DELETE("/:uuid", deps.journalHandler.DeleteJournal)
			journals.POST("/:uuid/toggle-ai-share", deps.journalHandler.ToggleAIShare)
		}

		mood := v1.Group("/user-moods")
		mood.Use(middleware.AuthMiddleware())
		{
			mood.GET("", deps.moodHandler.GetMoodHistory)
			mood.POST("", deps.moodHandler.RecordMood)
			mood.GET("/today", deps.moodHandler.CheckTodayMood)
			mood.GET("/latest", deps.moodHandler.GetLatestMood)
			mood.GET("/stats", deps.moodHandler.GetMoodStats)
		}

		v1.GET("/level-configs", deps.levelConfigHandler.GetAllConfigs)

		expHistory := v1.Group("/exp-history")
		expHistory.Use(middleware.AuthMiddleware())
		{
			expHistory.GET("", deps.expHistoryHandler.GetHistory)
			expHistory.GET("/activity-types", deps.expHistoryHandler.GetActivityTypes)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/stats", deps.adminHandler.GetDashboardStats)
			admin.GET("/articles", deps.adminHandler.GetAllArticles)
			admin.POST("/articles", deps.adminHandler.CreateArticle)
			admin.PUT("/articles/:id", deps.adminHandler.UpdateArticle)
			admin.DELETE("/articles/:id", deps.adminHandler.DeleteArticle)
			admin.PUT("/articles/:id/block", deps.adminHandler.BlockArticle)
			admin.PUT("/articles/:id/unblock", deps.adminHandler.UnblockArticle)
			admin.GET("/article-categories", deps.adminHandler.GetArticleCategories)
			admin.POST("/article-categories", deps.adminHandler.CreateArticleCategory)
			admin.PUT("/article-categories/:id", deps.adminHandler.UpdateArticleCategory)
			admin.DELETE("/article-categories/:id", deps.adminHandler.DeleteArticleCategory)
			admin.GET("/forum-categories", deps.forumCategoryHandler.AdminGetAllCategories)
			admin.POST("/forum-categories", deps.forumCategoryHandler.CreateCategory)
			admin.PUT("/forum-categories/:id", deps.forumCategoryHandler.UpdateCategory)
			admin.DELETE("/forum-categories/:id", deps.forumCategoryHandler.DeleteCategory)
			admin.GET("/forums", deps.adminHandler.GetForums)
			admin.POST("/forums/:id/toggle-flag", deps.adminHandler.ToggleForumFlag)

			admin.GET("/users", deps.adminHandler.GetUsers)
			admin.DELETE("/users/:id", deps.adminHandler.DeleteUser)
			admin.PUT("/users/:id/block", deps.adminHandler.BlockUser)
			admin.PUT("/users/:id/unblock", deps.adminHandler.UnblockUser)
			admin.PUT("/users/:id/block-journal", deps.adminHandler.ToggleJournalBlock)
			admin.PUT("/users/:id/block-forum", deps.adminHandler.ToggleForumBlock)
			admin.POST("/song-categories", deps.adminHandler.CreateSongCategory)
			admin.PUT("/song-categories/:id", deps.adminHandler.UpdateSongCategory)
			admin.DELETE("/song-categories/:id", deps.adminHandler.DeleteSongCategory)
			admin.GET("/songs", deps.adminHandler.GetAllSongs)
			admin.POST("/songs", deps.adminHandler.CreateSong)
			admin.PUT("/songs/:id", deps.adminHandler.UpdateSong)
			admin.DELETE("/songs/:id", deps.adminHandler.DeleteSong)
			admin.GET("/level-configs", deps.levelConfigHandler.AdminGetAllConfigs)
			admin.POST("/level-configs", deps.levelConfigHandler.CreateConfig)
			admin.PUT("/level-configs/:id", deps.levelConfigHandler.UpdateConfig)
			admin.DELETE("/level-configs/:id", deps.levelConfigHandler.DeleteConfig)
			admin.GET("/map-landmarks", deps.progressMapHandler.AdminGetAllLandmarks)
			admin.POST("/map-landmarks", deps.progressMapHandler.AdminCreateLandmark)
			admin.PUT("/map-landmarks/:id", deps.progressMapHandler.AdminUpdateLandmark)
			admin.DELETE("/map-landmarks/:id", deps.progressMapHandler.AdminDeleteLandmark)
			admin.POST("/cache/clear", deps.adminHandler.ClearCache)

			// Reward management
			admin.GET("/rewards", deps.rewardHandler.AdminGetAllRewards)
			admin.POST("/rewards", deps.rewardHandler.AdminCreateReward)
			admin.PUT("/rewards/:id", deps.rewardHandler.AdminUpdateReward)
			admin.DELETE("/rewards/:id", deps.rewardHandler.AdminDeleteReward)
			admin.GET("/rewards/claims", deps.rewardHandler.AdminGetAllClaims)

			// Broadcast notifications
			admin.GET("/broadcasts", deps.broadcastHandler.GetAll)
			admin.POST("/broadcasts", deps.broadcastHandler.Create)
			admin.GET("/broadcasts/:id", deps.broadcastHandler.GetByID)
			admin.PUT("/broadcasts/:id", deps.broadcastHandler.Update)
			admin.DELETE("/broadcasts/:id", deps.broadcastHandler.Delete)
			admin.POST("/broadcasts/:id/send", deps.broadcastHandler.SendNow)
			admin.POST("/broadcasts/:id/cancel", deps.broadcastHandler.Cancel)
		}

		v1.GET("/forum-categories", deps.forumCategoryHandler.GetAllCategories)
		v1.GET("/forums/report-reasons", deps.forumHandler.GetReportReasons)
		v1.GET("/forums/sort-options", deps.forumHandler.GetSortOptions)

		forum := v1.Group("/forums")
		forum.Use(middleware.AuthMiddleware())
		{
			forum.POST("", deps.forumHandler.CreateForum)
			forum.GET("", deps.forumHandler.GetForums)
			forum.GET("/:slug", deps.forumHandler.GetForumByID)
			forum.DELETE("/:slug", deps.forumHandler.DeleteForum)
			forum.POST("/:slug/posts", deps.forumHandler.CreateForumPost)
			forum.GET("/:slug/posts", deps.forumHandler.GetForumPosts)
			forum.PUT("/:slug/like", deps.forumHandler.ToggleLike)
			forum.GET("/:slug/accepted-answer", deps.forumHandler.GetAcceptedAnswer)
			forum.DELETE("/:slug/accepted-answer", deps.forumHandler.UnmarkAcceptedAnswer)
		}

		posts := v1.Group("/posts")
		posts.Use(middleware.AuthMiddleware())
		{
			posts.DELETE("/:id", deps.forumHandler.DeleteForumPost)
			posts.PUT("/:id/upvote", deps.forumHandler.UpvotePost)
			posts.PUT("/:id/downvote", deps.forumHandler.DownvotePost)
			posts.DELETE("/:id/vote", deps.forumHandler.RemovePostVote)
			posts.PUT("/:id/accept", deps.forumHandler.MarkAcceptedAnswer)
			posts.POST("/:id/report", deps.forumHandler.ReportPost)
		}

		moderation := v1.Group("/moderation")
		moderation.Use(middleware.AuthMiddleware())
		moderation.Use(middleware.AdminMiddleware())
		{
			moderation.GET("/stats", deps.moderationHandler.GetModerationStats)
			moderation.GET("/queue", deps.moderationHandler.GetModerationQueue)
			moderation.PUT("/articles/:id", deps.moderationHandler.ModerateArticle)
			moderation.GET("/post-reports", deps.forumHandler.GetPendingPostReports)
			moderation.PUT("/post-reports/:id", deps.forumHandler.ReviewPostReport)
			moderation.GET("/reports", deps.moderationHandler.GetReports)
			moderation.PUT("/reports/:id", deps.moderationHandler.HandleReport)
			moderation.GET("/users/:id/strikes", deps.moderationHandler.GetUserStrikes)
			moderation.POST("/trigger-warnings", deps.moderationHandler.AddTriggerWarnings)
			moderation.GET("/actions", deps.moderationHandler.GetModeratorActions)
			moderation.GET("/crisis-keywords", deps.moderationHandler.GetCrisisKeywords)
			moderation.POST("/crisis-keywords", deps.moderationHandler.CreateCrisisKeyword)
			moderation.DELETE("/crisis-keywords/:id", deps.moderationHandler.DeleteCrisisKeyword)
			moderation.GET("/appeals", deps.moderationHandler.GetAppeals)
			moderation.PUT("/appeals/:id", deps.moderationHandler.ReviewAppeal)
		}

		reports := v1.Group("/reports")
		reports.Use(middleware.AuthMiddleware())
		{
			reports.POST("", deps.moderationHandler.CreateReport)
		}

		appeals := v1.Group("/appeals")
		appeals.Use(middleware.AuthMiddleware())
		{
			appeals.POST("", deps.moderationHandler.CreateAppeal)
		}

		blocks := v1.Group("/blocks")
		blocks.Use(middleware.AuthMiddleware())
		{
			blocks.GET("", deps.moderationHandler.GetBlockedUsers)
			blocks.POST("", deps.moderationHandler.BlockUser)
			blocks.DELETE("/:id", deps.moderationHandler.UnblockUser)
		}

		userSettings := v1.Group("/user")
		userSettings.Use(middleware.AuthMiddleware())
		{
			userSettings.POST("/accept-ai-disclaimer", deps.moderationHandler.AcceptAIDisclaimer)
			userSettings.PUT("/content-warning-preference", deps.moderationHandler.UpdateContentWarningPreference)
		}

		v1.GET("/search", deps.searchHandler.Search)

		v1.GET("/community/stats", middleware.OpenRateLimit(), deps.communityProgressHandler.GetCommunityStats)
		v1.GET("/community/hall-of-fame/level/:level", middleware.OpenRateLimit(), deps.communityProgressHandler.GetLevelHallOfFame)
		v1.GET("/community/hall-of-fame/monthly", middleware.OpenRateLimit(), deps.communityProgressHandler.GetMonthlyHallOfFame)
		v1.GET("/community/hall-of-fame/categories", middleware.OpenRateLimit(), deps.communityProgressHandler.GetHallOfFameCategories)

		community := v1.Group("/community")
		community.Use(middleware.AuthMiddleware())
		community.Use(middleware.RelaxedRateLimit())
		{
			community.GET("/my-journey", deps.communityProgressHandler.GetPersonalJourney)
			community.GET("/my-progress/weekly", deps.communityProgressHandler.GetWeeklyProgress)
			community.GET("/my-progress/monthly", deps.communityProgressHandler.GetMonthlyProgress)
			community.GET("/my-stats", deps.communityProgressHandler.GetAllTimeStats)
			community.GET("/celebrate/:level", deps.communityProgressHandler.GetLevelUpCelebration)
		}

		v1.GET("/features", middleware.OpenRateLimit(), deps.featureUnlockHandler.GetAllFeatures)
		v1.GET("/features/categories", middleware.OpenRateLimit(), deps.featureUnlockHandler.GetFeatureCategories)
		v1.GET("/features/category/:category", middleware.OpenRateLimit(), deps.featureUnlockHandler.GetFeaturesByCategory)

		features := v1.Group("/features")
		features.Use(middleware.AuthMiddleware())
		features.Use(middleware.RelaxedRateLimit())
		{
			features.GET("/my-features", deps.featureUnlockHandler.GetUserFeatures)
			features.GET("/check/:featureKey", deps.featureUnlockHandler.CheckFeatureAccess)
			features.GET("/upcoming", deps.featureUnlockHandler.GetUpcomingFeatures)
		}

		v1.GET("/badges", middleware.OpenRateLimit(), deps.badgeHandler.GetAllBadges)
		v1.GET("/badges/categories", middleware.OpenRateLimit(), deps.badgeHandler.GetBadgeCategories)
		v1.GET("/badges/category/:category", middleware.OpenRateLimit(), deps.badgeHandler.GetBadgesByCategory)

		badges := v1.Group("/badges")
		badges.Use(middleware.AuthMiddleware())
		badges.Use(middleware.RelaxedRateLimit())
		{
			badges.GET("/my-badges", deps.badgeHandler.GetUserBadges)
			badges.GET("/progress", deps.badgeHandler.GetBadgeProgress)
			badges.GET("/recent", deps.badgeHandler.GetRecentlyEarnedBadges)
			badges.GET("/display", deps.badgeHandler.GetDisplayBadges)
			badges.POST("/check", deps.badgeHandler.CheckNewBadges)
		}

		v1.GET("/stories/categories", middleware.OpenRateLimit(), deps.inspiringStoryHandler.GetCategories)
		v1.GET("/stories/featured", middleware.OpenRateLimit(), deps.inspiringStoryHandler.GetFeaturedStories)
		v1.GET("/stories/most-appreciated", middleware.OpenRateLimit(), deps.inspiringStoryHandler.GetMostAppreciated)
		v1.GET("/stories", middleware.OpenRateLimit(), middleware.OptionalAuthMiddleware(), deps.inspiringStoryHandler.GetStories)
		v1.GET("/stories/:id", middleware.OpenRateLimit(), middleware.OptionalAuthMiddleware(), deps.inspiringStoryHandler.GetStory)
		v1.GET("/stories/:id/comments", middleware.OpenRateLimit(), middleware.OptionalAuthMiddleware(), deps.inspiringStoryHandler.GetComments)

		stories := v1.Group("/stories")
		stories.Use(middleware.AuthMiddleware())
		stories.Use(middleware.ModerateRateLimit())
		{
			stories.POST("", deps.inspiringStoryHandler.CreateStory)
			stories.PUT("/:id", deps.inspiringStoryHandler.UpdateStory)
			stories.DELETE("/:id", deps.inspiringStoryHandler.DeleteStory)
			stories.GET("/my-stories", deps.inspiringStoryHandler.GetMyStories)
			stories.GET("/my-stats", deps.inspiringStoryHandler.GetMyStats)
			stories.POST("/:id/heart", deps.inspiringStoryHandler.ToggleHeart)
			stories.POST("/:id/comments", deps.inspiringStoryHandler.CreateComment)
			stories.DELETE("/:id/comments/:commentId", deps.inspiringStoryHandler.DeleteComment)
			stories.POST("/:id/comments/:commentId/heart", deps.inspiringStoryHandler.ToggleCommentHeart)
		}

		notifications := v1.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		notifications.Use(middleware.ModerateRateLimit())
		{
			notifications.GET("", deps.notificationHandler.GetNotifications)
			notifications.GET("/unread-count", deps.notificationHandler.GetUnreadCount)
			notifications.POST("/:id/read", deps.notificationHandler.MarkAsRead)
			notifications.POST("/read-all", deps.notificationHandler.MarkAllAsRead)
		}

		push := v1.Group("/push")
		push.Use(middleware.ModerateRateLimit())
		{
			push.GET("/vapid-key", deps.pushHandler.GetVAPIDKey)
			push.POST("/subscribe", middleware.AuthMiddleware(), deps.pushHandler.Subscribe)
			push.POST("/unsubscribe", middleware.AuthMiddleware(), deps.pushHandler.Unsubscribe)
		}

		adminStories := v1.Group("/admin/stories")
		adminStories.Use(middleware.AuthMiddleware())
		adminStories.Use(middleware.AdminMiddleware())
		{
			adminStories.GET("/pending", deps.inspiringStoryHandler.GetPendingStories)
			adminStories.POST("/:id/moderate", deps.inspiringStoryHandler.ModerateStory)
			adminStories.POST("/:id/featured", deps.inspiringStoryHandler.SetFeatured)
			adminStories.POST("/:id/comments/:commentId/hide", deps.inspiringStoryHandler.HideComment)
		}

		breathing := v1.Group("/breathing")
		breathing.Use(middleware.AuthMiddleware())
		breathing.Use(middleware.RelaxedRateLimit())
		{
			breathing.GET("/techniques", deps.breathingHandler.GetTechniques)
			breathing.GET("/techniques/:id", deps.breathingHandler.GetTechniqueByID)
			breathing.GET("/techniques/slug/:slug", deps.breathingHandler.GetTechniqueBySlug)
			breathing.POST("/techniques", deps.breathingHandler.CreateTechnique)
			breathing.PUT("/techniques/:id", deps.breathingHandler.UpdateTechnique)
			breathing.DELETE("/techniques/:id", deps.breathingHandler.DeleteTechnique)
			breathing.GET("/sessions", deps.breathingHandler.GetSessionHistory)
			breathing.POST("/sessions", deps.breathingHandler.StartSession)
			breathing.GET("/sessions/:id", deps.breathingHandler.GetSessionByID)
			breathing.POST("/sessions/:id/complete", deps.breathingHandler.CompleteSession)
			breathing.GET("/preferences", deps.breathingHandler.GetPreferences)
			breathing.PUT("/preferences", deps.breathingHandler.UpdatePreferences)
			breathing.GET("/favorites", deps.breathingHandler.GetFavorites)
			breathing.POST("/favorites/:id", deps.breathingHandler.AddFavorite)
			breathing.DELETE("/favorites/:id", deps.breathingHandler.RemoveFavorite)
			breathing.PUT("/favorites/reorder", deps.breathingHandler.ReorderFavorites)
			breathing.GET("/stats", deps.breathingHandler.GetStats)
			breathing.GET("/stats/usage", deps.breathingHandler.GetTechniqueUsage)
			breathing.GET("/calendar", deps.breathingHandler.GetCalendar)
			breathing.GET("/widget", deps.breathingHandler.GetWidgetData)
			breathing.GET("/recommendations", deps.breathingHandler.GetRecommendations)
		}

		dailyTasks := v1.Group("/daily-tasks")
		dailyTasks.Use(middleware.AuthMiddleware())
		dailyTasks.Use(middleware.RelaxedRateLimit())
		{
			dailyTasks.GET("", deps.dailyTaskHandler.GetDailyTasks)
			dailyTasks.POST("/login", deps.dailyTaskHandler.ClaimDailyLogin)
			dailyTasks.POST("/:id/claim", deps.dailyTaskHandler.ClaimTaskReward)
			dailyTasks.POST("/claim-all", deps.dailyTaskHandler.ClaimAllRewards)
			dailyTasks.GET("/history", deps.dailyTaskHandler.GetTaskHistory)
		}

		rewards := v1.Group("/rewards")
		rewards.Use(middleware.AuthMiddleware())
		rewards.Use(middleware.RelaxedRateLimit())
		{
			rewards.GET("", deps.rewardHandler.GetAvailableRewards)
			rewards.GET("/balance", deps.rewardHandler.GetCoinBalance)
			rewards.GET("/my-claims", deps.rewardHandler.GetMyClaims)
			rewards.GET("/themes", deps.rewardHandler.GetOwnedThemes)
			rewards.PUT("/themes/activate", deps.rewardHandler.ActivateTheme)
			rewards.GET("/:id", deps.rewardHandler.GetRewardDetail)
			rewards.POST("/:id/claim", deps.rewardHandler.ClaimReward)
		}

		// Public guild endpoints
		v1.GET("/guilds", middleware.OpenRateLimit(), deps.guildHandler.GetPublicGuilds)
		v1.GET("/guilds/leaderboard", middleware.OpenRateLimit(), deps.guildHandler.GetGuildLeaderboard)

		guilds := v1.Group("/guilds")
		guilds.Use(middleware.AuthMiddleware())
		guilds.Use(middleware.RelaxedRateLimit())
		{
			guilds.POST("", deps.guildHandler.CreateGuild)
			guilds.GET("/my-guild", deps.guildHandler.GetMyGuild)
			guilds.POST("/join/:code", deps.guildHandler.JoinByInviteCode)
			guilds.GET("/:id", deps.guildHandler.GetGuild)
			guilds.PUT("/:id", deps.guildHandler.UpdateGuild)
			guilds.DELETE("/:id", deps.guildHandler.DeleteGuild)
			guilds.POST("/:id/join", deps.guildHandler.JoinGuild)
			guilds.POST("/:id/leave", deps.guildHandler.LeaveGuild)
			guilds.POST("/:id/kick/:userId", deps.guildHandler.KickMember)
			guilds.POST("/:id/promote/:userId", deps.guildHandler.PromoteMember)
			guilds.POST("/:id/transfer/:userId", deps.guildHandler.TransferLeadership)
			guilds.POST("/:id/challenges", deps.guildHandler.CreateChallenge)
			guilds.GET("/:id/challenges", deps.guildHandler.GetActiveChallenges)
			guilds.GET("/:id/challenges/history", deps.guildHandler.GetChallengeHistory)
			guilds.GET("/:id/activities", deps.guildHandler.GetRecentActivities)
		}

		progressMap := v1.Group("/map")
		progressMap.Use(middleware.AuthMiddleware())
		progressMap.Use(middleware.RelaxedRateLimit())
		{
			progressMap.GET("", deps.progressMapHandler.GetFullMap)
			progressMap.GET("/summary", deps.progressMapHandler.GetProgressSummary)
			progressMap.GET("/regions/:key", deps.progressMapHandler.GetRegionDetail)
			progressMap.POST("/landmarks/:id/claim", deps.progressMapHandler.ClaimLandmarkReward)
		}

		// Weekly Leagues
		leagues := v1.Group("/leagues")
		leagues.Use(middleware.AuthMiddleware())
		leagues.Use(middleware.RelaxedRateLimit())
		{
			leagues.GET("/overview", deps.weeklyLeagueHandler.GetOverview)
			leagues.GET("/divisions", deps.weeklyLeagueHandler.GetDivisions)
		}

		// XP Boost & Combo
		xpBoost := v1.Group("/xp-boost")
		xpBoost.Use(middleware.AuthMiddleware())
		xpBoost.Use(middleware.RelaxedRateLimit())
		{
			xpBoost.GET("/active", deps.xpBoostComboHandler.GetActiveBoost)
			xpBoost.GET("/multiplier", deps.xpBoostComboHandler.GetEffectiveMultiplier)
		}

		combo := v1.Group("/combo")
		combo.Use(middleware.AuthMiddleware())
		combo.Use(middleware.RelaxedRateLimit())
		{
			combo.GET("/status", deps.xpBoostComboHandler.GetComboStatus)
		}

		// Mystery Chests
		chests := v1.Group("/chests")
		chests.Use(middleware.AuthMiddleware())
		chests.Use(middleware.RelaxedRateLimit())
		{
			chests.GET("", deps.mysteryChestHandler.GetMyChests)
			chests.POST("/:id/open", deps.mysteryChestHandler.OpenChest)
		}

		// Friend Quests
		friendQuests := v1.Group("/friend-quests")
		friendQuests.Use(middleware.AuthMiddleware())
		friendQuests.Use(middleware.RelaxedRateLimit())
		{
			friendQuests.POST("", deps.friendQuestHandler.CreateQuest)
			friendQuests.GET("", deps.friendQuestHandler.GetMyQuests)
			friendQuests.GET("/:id", deps.friendQuestHandler.GetQuest)
			friendQuests.POST("/:id/accept", deps.friendQuestHandler.AcceptQuest)
			friendQuests.POST("/:id/decline", deps.friendQuestHandler.DeclineQuest)
		}

		// Daily Spin (Roulette)
		dailySpin := v1.Group("/daily-spin")
		dailySpin.Use(middleware.AuthMiddleware())
		dailySpin.Use(middleware.RelaxedRateLimit())
		{
			dailySpin.GET("/wheel", deps.dailySpinHandler.GetWheel)
			dailySpin.POST("/spin", deps.dailySpinHandler.Spin)
		}

		// Streak Society
		streakSociety := v1.Group("/streak-society")
		streakSociety.Use(middleware.AuthMiddleware())
		streakSociety.Use(middleware.RelaxedRateLimit())
		{
			streakSociety.GET("/overview", deps.streakSocietyHandler.GetOverview)
			streakSociety.POST("/join", deps.streakSocietyHandler.JoinSociety)
			streakSociety.GET("/:id/members", deps.streakSocietyHandler.GetMembers)
		}

		// Timed Challenges (Quest Kilat)
		challenges := v1.Group("/challenges")
		challenges.Use(middleware.AuthMiddleware())
		challenges.Use(middleware.RelaxedRateLimit())
		{
			challenges.GET("/templates", deps.timedChallengeHandler.GetTemplates)
			challenges.POST("/start", deps.timedChallengeHandler.StartChallenge)
			challenges.GET("/active", deps.timedChallengeHandler.GetActiveChallenge)
			challenges.POST("/:id/complete", deps.timedChallengeHandler.CompleteChallenge)
			challenges.GET("/history", deps.timedChallengeHandler.GetHistory)
		}
	}
}
