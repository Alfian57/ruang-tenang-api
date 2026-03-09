package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/handler"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
)

type routeDependencies struct {
	cacheService *service.CacheService

	authHandler              *handler.AuthHandler
	userHandler              *handler.UserHandler
	articleHandler           *handler.ArticleHandler
	chatHandler              *handler.ChatHandler
	uploadHandler            *handler.UploadHandler
	songHandler              *handler.SongHandler
	moodHandler              *handler.MoodHandler
	adminHandler             *handler.AdminHandler
	searchHandler            *handler.SearchHandler
	forumHandler             *handler.ForumHandler
	forumCategoryHandler     *handler.ForumCategoryHandler
	levelConfigHandler       *handler.LevelConfigHandler
	expHistoryHandler        *handler.ExpHistoryHandler
	moderationHandler        *handler.ModerationHandler
	communityProgressHandler *handler.CommunityProgressHandler
	featureUnlockHandler     *handler.FeatureUnlockHandler
	badgeHandler             *handler.BadgeHandler
	inspiringStoryHandler    *handler.InspiringStoryHandler
	breathingHandler         *handler.BreathingHandler
	playlistHandler          *handler.PlaylistHandler
	journalHandler           *handler.JournalHandler
	dailyTaskHandler         *handler.DailyTaskHandler
	notificationHandler      *handler.NotificationHandler
	rewardHandler            *handler.RewardHandler
	guildHandler             *handler.GuildHandler
	progressMapHandler       *handler.ProgressMapHandler
	weeklyLeagueHandler      *handler.WeeklyLeagueHandler
	xpBoostComboHandler      *handler.XPBoostComboHandler
	mysteryChestHandler      *handler.MysteryChestHandler
	friendQuestHandler       *handler.FriendQuestHandler
	dailySpinHandler         *handler.DailySpinHandler
	streakSocietyHandler     *handler.StreakSocietyHandler
	timedChallengeHandler    *handler.TimedChallengeHandler
	pushHandler              *handler.PushHandler
	broadcastHandler         *handler.BroadcastHandler
	broadcastService         *service.BroadcastService
}

func initializeRouteDependencies(cfg *config.Config) *routeDependencies {
	db := database.GetDB()

	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	articleCategoryRepo := repository.NewArticleCategoryRepository(db)
	chatSessionRepo := repository.NewChatSessionRepository(db)
	chatMessageRepo := repository.NewChatMessageRepository(db)
	chatFolderRepo := repository.NewChatFolderRepository(db)
	songRepo := repository.NewSongRepository(db)
	songCategoryRepo := repository.NewSongCategoryRepository(db)
	moodRepo := repository.NewUserMoodRepository(db)
	forumRepo := repository.NewForumRepository(db)
	forumCategoryRepo := repository.NewForumCategoryRepository(db)
	levelConfigRepo := repository.NewLevelConfigRepository(db)
	expHistoryRepo := repository.NewExpHistoryRepository(db)
	moderationRepo := repository.NewModerationRepository(db)
	communityProgressRepo := repository.NewCommunityProgressRepository(db)
	featureUnlockRepo := repository.NewFeatureUnlockRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)
	inspiringStoryRepo := repository.NewInspiringStoryRepository(db)
	breathingRepo := repository.NewBreathingRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)
	playlistItemRepo := repository.NewPlaylistItemRepository(db)
	journalRepo := repository.NewJournalRepository(db)
	journalSettingsRepo := repository.NewJournalSettingsRepository(db)
	journalAccessLogRepo := repository.NewJournalAIAccessLogRepository(db)
	dailyTaskRepo := repository.NewDailyTaskRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	rewardRepo := repository.NewRewardRepository(db)
	pushSubRepo := repository.NewPushSubscriptionRepository(db)
	guildRepo := repository.NewGuildRepository(db)
	progressMapRepo := repository.NewProgressMapRepository(db)
	weeklyLeagueRepo := repository.NewWeeklyLeagueRepository(db)
	xpBoostComboRepo := repository.NewXPBoostComboRepository(db)
	mysteryChestRepo := repository.NewMysteryChestRepository(db)
	friendQuestRepo := repository.NewFriendQuestRepository(db)
	dailySpinRepo := repository.NewDailySpinRepository(db)
	streakSocietyRepo := repository.NewStreakSocietyRepository(db)
	timedChallengeRepo := repository.NewTimedChallengeRepository(db)
	broadcastRepo := repository.NewBroadcastNotificationRepository(db)

	cacheService := service.NewCacheService()
	gamificationService := service.NewGamificationService(db)
	contentContextService := service.NewContentContextService(articleRepo, songRepo, songCategoryRepo, forumRepo)
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	aiModerationService := service.NewAIModerationService(moderationRepo, cfg)
	moderationService := service.NewModerationService(moderationRepo, userRepo, articleRepo, forumRepo, aiModerationService)
	articleService := service.NewArticleService(articleRepo, articleCategoryRepo, gamificationService, contentContextService, cacheService, moderationService)
	songService := service.NewSongService(songRepo, songCategoryRepo, cacheService)
	moodService := service.NewMoodService(moodRepo)
	forumService := service.NewForumService(forumRepo, userRepo, gamificationService, contentContextService)
	forumCategoryService := service.NewForumCategoryService(forumCategoryRepo, cacheService)
	levelConfigService := service.NewLevelConfigService(levelConfigRepo, cacheService)
	expHistoryService := service.NewExpHistoryService(expHistoryRepo)
	chatService := service.NewChatService(chatSessionRepo, chatMessageRepo, cfg, gamificationService, contentContextService)
	communityProgressService := service.NewCommunityProgressService(communityProgressRepo, levelConfigRepo, featureUnlockRepo, badgeRepo, userRepo)
	featureUnlockService := service.NewFeatureUnlockService(featureUnlockRepo, levelConfigRepo, userRepo)
	badgeService := service.NewBadgeService(badgeRepo, userRepo, levelConfigRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	pushService := service.NewPushService(pushSubRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDContact)
	notificationService.SetPushService(pushService)
	inspiringStoryService := service.NewInspiringStoryService(inspiringStoryRepo, userRepo, levelConfigRepo, badgeService, gamificationService, notificationService)
	dailyTaskService := service.NewDailyTaskService(dailyTaskRepo, userRepo)
	breathingService := service.NewBreathingService(breathingRepo, gamificationService, dailyTaskService)
	playlistService := service.NewPlaylistService(playlistRepo, playlistItemRepo, songRepo)
	journalService := service.NewJournalService(journalRepo, journalSettingsRepo, journalAccessLogRepo, moodRepo, chatService.GetGenAIClient())

	chatService.SetModerationRepo(moderationRepo)
	chatService.SetFolderRepo(chatFolderRepo)
	chatService.SetJournalRepos(journalRepo, journalSettingsRepo, journalAccessLogRepo)

	authHandler := handler.NewAuthHandler(authService, levelConfigService)
	userHandler := handler.NewUserHandler(userService, levelConfigService)
	articleHandler := handler.NewArticleHandler(articleService)
	chatHandler := handler.NewChatHandler(chatService)
	uploadHandler := handler.NewUploadHandler()
	songHandler := handler.NewSongHandler(songService)
	moodHandler := handler.NewMoodHandler(moodService)
	adminHandler := handler.NewAdminHandler(db, userRepo, articleRepo, forumRepo, cacheService, journalService)
	searchHandler := handler.NewSearchHandler(articleRepo, songRepo)
	forumHandler := handler.NewForumHandler(forumService)
	forumCategoryHandler := handler.NewForumCategoryHandler(forumCategoryService)
	levelConfigHandler := handler.NewLevelConfigHandler(levelConfigService)
	expHistoryHandler := handler.NewExpHistoryHandler(expHistoryService, levelConfigService)
	moderationHandler := handler.NewModerationHandler(moderationService)
	communityProgressHandler := handler.NewCommunityProgressHandler(communityProgressService)
	featureUnlockHandler := handler.NewFeatureUnlockHandler(featureUnlockService)
	badgeHandler := handler.NewBadgeHandler(badgeService)
	inspiringStoryHandler := handler.NewInspiringStoryHandler(inspiringStoryService)
	breathingHandler := handler.NewBreathingHandler(breathingService)
	playlistHandler := handler.NewPlaylistHandler(playlistService)
	journalHandler := handler.NewJournalHandler(journalService)
	dailyTaskHandler := handler.NewDailyTaskHandler(dailyTaskService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	pushHandler := handler.NewPushHandler(pushService)
	broadcastService := service.NewBroadcastService(broadcastRepo, pushSubRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDContact)
	broadcastHandler := handler.NewBroadcastHandler(broadcastService)
	broadcastService.StartScheduler()
	rewardService := service.NewRewardService(rewardRepo, userRepo)
	rewardHandler := handler.NewRewardHandler(rewardService)
	guildService := service.NewGuildService(guildRepo, userRepo, levelConfigRepo)
	guildHandler := handler.NewGuildHandler(guildService)
	progressMapService := service.NewProgressMapService(progressMapRepo, userRepo, levelConfigRepo)
	progressMapHandler := handler.NewProgressMapHandler(progressMapService)
	weeklyLeagueService := service.NewWeeklyLeagueService(weeklyLeagueRepo, userRepo)
	weeklyLeagueHandler := handler.NewWeeklyLeagueHandler(weeklyLeagueService)
	xpBoostComboService := service.NewXPBoostComboService(xpBoostComboRepo)
	xpBoostComboHandler := handler.NewXPBoostComboHandler(xpBoostComboService)
	mysteryChestService := service.NewMysteryChestService(mysteryChestRepo, userRepo)
	mysteryChestHandler := handler.NewMysteryChestHandler(mysteryChestService)
	friendQuestService := service.NewFriendQuestService(friendQuestRepo, userRepo)
	friendQuestHandler := handler.NewFriendQuestHandler(friendQuestService)
	dailySpinService := service.NewDailySpinService(dailySpinRepo, userRepo)
	dailySpinHandler := handler.NewDailySpinHandler(dailySpinService)
	streakSocietyService := service.NewStreakSocietyService(streakSocietyRepo, userRepo)
	streakSocietyHandler := handler.NewStreakSocietyHandler(streakSocietyService)
	timedChallengeService := service.NewTimedChallengeService(timedChallengeRepo, userRepo)
	timedChallengeHandler := handler.NewTimedChallengeHandler(timedChallengeService)

	moodHandler.SetDailyTaskService(dailyTaskService)
	chatHandler.SetDailyTaskService(dailyTaskService)
	journalHandler.SetDailyTaskService(dailyTaskService)
	forumHandler.SetDailyTaskService(dailyTaskService)
	articleHandler.SetDailyTaskService(dailyTaskService)
	songHandler.SetDailyTaskService(dailyTaskService)

	return &routeDependencies{
		cacheService: cacheService,

		authHandler:              authHandler,
		userHandler:              userHandler,
		articleHandler:           articleHandler,
		chatHandler:              chatHandler,
		uploadHandler:            uploadHandler,
		songHandler:              songHandler,
		moodHandler:              moodHandler,
		adminHandler:             adminHandler,
		searchHandler:            searchHandler,
		forumHandler:             forumHandler,
		forumCategoryHandler:     forumCategoryHandler,
		levelConfigHandler:       levelConfigHandler,
		expHistoryHandler:        expHistoryHandler,
		moderationHandler:        moderationHandler,
		communityProgressHandler: communityProgressHandler,
		featureUnlockHandler:     featureUnlockHandler,
		badgeHandler:             badgeHandler,
		inspiringStoryHandler:    inspiringStoryHandler,
		breathingHandler:         breathingHandler,
		playlistHandler:          playlistHandler,
		journalHandler:           journalHandler,
		dailyTaskHandler:         dailyTaskHandler,
		notificationHandler:      notificationHandler,
		rewardHandler:            rewardHandler,
		guildHandler:             guildHandler,
		progressMapHandler:       progressMapHandler,
		weeklyLeagueHandler:      weeklyLeagueHandler,
		xpBoostComboHandler:      xpBoostComboHandler,
		mysteryChestHandler:      mysteryChestHandler,
		friendQuestHandler:       friendQuestHandler,
		dailySpinHandler:         dailySpinHandler,
		streakSocietyHandler:     streakSocietyHandler,
		timedChallengeHandler:    timedChallengeHandler,
		pushHandler:              pushHandler,
		broadcastHandler:         broadcastHandler,
		broadcastService:         broadcastService,
	}
}
