package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"

	// Shared
	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/contentctx"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/userctx"

	// Features
	adminhandler "github.com/Alfian57/ruang-tenang-api/internal/features/admin/interface/http"
	articleapp "github.com/Alfian57/ruang-tenang-api/internal/features/article/application"
	articleinfra "github.com/Alfian57/ruang-tenang-api/internal/features/article/infrastructure"
	articlehandler "github.com/Alfian57/ruang-tenang-api/internal/features/article/interface/http"
	authapp "github.com/Alfian57/ruang-tenang-api/internal/features/auth/application"
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	authhandler "github.com/Alfian57/ruang-tenang-api/internal/features/auth/interface/http"
	badgeapp "github.com/Alfian57/ruang-tenang-api/internal/features/badge/application"
	badgeinfra "github.com/Alfian57/ruang-tenang-api/internal/features/badge/infrastructure"
	badgehandler "github.com/Alfian57/ruang-tenang-api/internal/features/badge/interface/http"
	billingapp "github.com/Alfian57/ruang-tenang-api/internal/features/billing/application"
	billinginfra "github.com/Alfian57/ruang-tenang-api/internal/features/billing/infrastructure"
	billinghandler "github.com/Alfian57/ruang-tenang-api/internal/features/billing/interface/http"
	breathingapp "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/application"
	breathinginfra "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/infrastructure"
	breathinghandler "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/interface/http"
	broadcastapp "github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/application"
	broadcastinfra "github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/infrastructure"
	broadcasthandler "github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/interface/http"
	chatapp "github.com/Alfian57/ruang-tenang-api/internal/features/chat/application"
	chatinfra "github.com/Alfian57/ruang-tenang-api/internal/features/chat/infrastructure"
	chathandler "github.com/Alfian57/ruang-tenang-api/internal/features/chat/interface/http"
	dailyspinapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/application"
	dailyspininfra "github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/infrastructure"
	dailyspinhandler "github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/interface/http"
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	dailytaskinfra "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/infrastructure"
	dailytaskhandler "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/interface/http"
	featureunlockapp "github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/application"
	featureunlockinfra "github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/infrastructure"
	featureunlockhandler "github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/interface/http"
	forumapp "github.com/Alfian57/ruang-tenang-api/internal/features/forum/application"
	foruminfra "github.com/Alfian57/ruang-tenang-api/internal/features/forum/infrastructure"
	forumhandler "github.com/Alfian57/ruang-tenang-api/internal/features/forum/interface/http"
	friendquestapp "github.com/Alfian57/ruang-tenang-api/internal/features/friend_quest/application"
	friendquestinfra "github.com/Alfian57/ruang-tenang-api/internal/features/friend_quest/infrastructure"
	friendquesthandler "github.com/Alfian57/ruang-tenang-api/internal/features/friend_quest/interface/http"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	gamificationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/infrastructure"
	gamificationhandler "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/interface/http"
	guildapp "github.com/Alfian57/ruang-tenang-api/internal/features/guild/application"
	guildinfra "github.com/Alfian57/ruang-tenang-api/internal/features/guild/infrastructure"
	guildhandler "github.com/Alfian57/ruang-tenang-api/internal/features/guild/interface/http"
	journalapp "github.com/Alfian57/ruang-tenang-api/internal/features/journal/application"
	journalinfra "github.com/Alfian57/ruang-tenang-api/internal/features/journal/infrastructure"
	journalhandler "github.com/Alfian57/ruang-tenang-api/internal/features/journal/interface/http"
	moderationapp "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/application"
	moderationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/infrastructure"
	moderationhandler "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/interface/http"
	moodapp "github.com/Alfian57/ruang-tenang-api/internal/features/mood/application"
	moodinfra "github.com/Alfian57/ruang-tenang-api/internal/features/mood/infrastructure"
	moodhandler "github.com/Alfian57/ruang-tenang-api/internal/features/mood/interface/http"
	mysterychestapp "github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/application"
	mysterychestinfra "github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/infrastructure"
	mysterychesthandler "github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/interface/http"
	notificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/notification/application"
	notificationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/notification/infrastructure"
	notificationhandler "github.com/Alfian57/ruang-tenang-api/internal/features/notification/interface/http"
	playlistapp "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/application"
	playlistinfra "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/infrastructure"
	playlisthandler "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/interface/http"
	progressmapapp "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/application"
	progressmapinfra "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/infrastructure"
	progressmaphandler "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/interface/http"
	pushapp "github.com/Alfian57/ruang-tenang-api/internal/features/push/application"
	pushinfra "github.com/Alfian57/ruang-tenang-api/internal/features/push/infrastructure"
	pushhandler "github.com/Alfian57/ruang-tenang-api/internal/features/push/interface/http"
	rewardapp "github.com/Alfian57/ruang-tenang-api/internal/features/reward/application"
	rewardinfra "github.com/Alfian57/ruang-tenang-api/internal/features/reward/infrastructure"
	rewardhandler "github.com/Alfian57/ruang-tenang-api/internal/features/reward/interface/http"
	searchhandler "github.com/Alfian57/ruang-tenang-api/internal/features/search/interface/http"
	songapp "github.com/Alfian57/ruang-tenang-api/internal/features/song/application"
	songinfra "github.com/Alfian57/ruang-tenang-api/internal/features/song/infrastructure"
	songhandler "github.com/Alfian57/ruang-tenang-api/internal/features/song/interface/http"
	storyapp "github.com/Alfian57/ruang-tenang-api/internal/features/story/application"
	storyinfra "github.com/Alfian57/ruang-tenang-api/internal/features/story/infrastructure"
	storyhandler "github.com/Alfian57/ruang-tenang-api/internal/features/story/interface/http"
	streaksocietyapp "github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/application"
	streaksocietyinfra "github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/infrastructure"
	streaksocietyhandler "github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/interface/http"
	timedchallengeapp "github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/application"
	timedchallengeinfra "github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/infrastructure"
	timedchallengehandler "github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/interface/http"
	uploadhandler "github.com/Alfian57/ruang-tenang-api/internal/features/upload/interface/http"
	weeklyleagueapp "github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/application"
	weeklyleagueinfra "github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/infrastructure"
	weeklyleaguehandler "github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/interface/http"
	xpboostapp "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/application"
	xpboostinfra "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/infrastructure"
	xpboosthandler "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/interface/http"
)

type routeDependencies struct {
	cacheService *cache.CacheService

	authHandler              *authhandler.AuthHandler
	userHandler              *authhandler.UserHandler
	billingHandler           *billinghandler.BillingHandler
	b2bHandler               *billinghandler.B2BHandler
	articleHandler           *articlehandler.ArticleHandler
	chatHandler              *chathandler.ChatHandler
	uploadHandler            *uploadhandler.UploadHandler
	songHandler              *songhandler.SongHandler
	moodHandler              *moodhandler.MoodHandler
	adminHandler             *adminhandler.AdminHandler
	searchHandler            *searchhandler.SearchHandler
	forumHandler             *forumhandler.ForumHandler
	forumCategoryHandler     *forumhandler.ForumCategoryHandler
	levelConfigHandler       *gamificationhandler.LevelConfigHandler
	expHistoryHandler        *gamificationhandler.ExpHistoryHandler
	moderationHandler        *moderationhandler.ModerationHandler
	communityProgressHandler *gamificationhandler.CommunityProgressHandler
	featureUnlockHandler     *featureunlockhandler.FeatureUnlockHandler
	badgeHandler             *badgehandler.BadgeHandler
	inspiringStoryHandler    *storyhandler.InspiringStoryHandler
	breathingHandler         *breathinghandler.BreathingHandler
	playlistHandler          *playlisthandler.PlaylistHandler
	journalHandler           *journalhandler.JournalHandler
	dailyTaskHandler         *dailytaskhandler.DailyTaskHandler
	notificationHandler      *notificationhandler.NotificationHandler
	rewardHandler            *rewardhandler.RewardHandler
	guildHandler             *guildhandler.GuildHandler
	progressMapHandler       *progressmaphandler.ProgressMapHandler
	weeklyLeagueHandler      *weeklyleaguehandler.WeeklyLeagueHandler
	xpBoostComboHandler      *xpboosthandler.XPBoostComboHandler
	mysteryChestHandler      *mysterychesthandler.MysteryChestHandler
	friendQuestHandler       *friendquesthandler.FriendQuestHandler
	dailySpinHandler         *dailyspinhandler.DailySpinHandler
	streakSocietyHandler     *streaksocietyhandler.StreakSocietyHandler
	timedChallengeHandler    *timedchallengehandler.TimedChallengeHandler
	pushHandler              *pushhandler.PushHandler
	broadcastHandler         *broadcasthandler.BroadcastHandler
	broadcastService         *broadcastapp.BroadcastService
}

func initializeRouteDependencies(cfg *config.Config) *routeDependencies {
	db := database.GetDB()

	// === Repositories ===
	userRepo := authinfra.NewUserRepository(db)
	articleRepo := articleinfra.NewArticleRepository(db)
	articleCategoryRepo := articleinfra.NewArticleCategoryRepository(db)
	chatSessionRepo := chatinfra.NewChatSessionRepository(db)
	chatMessageRepo := chatinfra.NewChatMessageRepository(db)
	chatFolderRepo := chatinfra.NewChatFolderRepository(db)
	songRepo := songinfra.NewSongRepository(db)
	songCategoryRepo := songinfra.NewSongCategoryRepository(db)
	moodRepo := moodinfra.NewUserMoodRepository(db)
	forumRepo := foruminfra.NewForumRepository(db)
	forumCategoryRepo := foruminfra.NewForumCategoryRepository(db)
	levelConfigRepo := gamificationinfra.NewLevelConfigRepository(db)
	expHistoryRepo := gamificationinfra.NewExpHistoryRepository(db)
	moderationRepo := moderationinfra.NewModerationRepository(db)
	communityProgressRepo := gamificationinfra.NewCommunityProgressRepository(db)
	featureUnlockRepo := featureunlockinfra.NewFeatureUnlockRepository(db)
	badgeRepo := badgeinfra.NewBadgeRepository(db)
	inspiringStoryRepo := storyinfra.NewInspiringStoryRepository(db)
	breathingRepo := breathinginfra.NewBreathingRepository(db)
	playlistRepo := playlistinfra.NewPlaylistRepository(db)
	playlistItemRepo := playlistinfra.NewPlaylistItemRepository(db)
	journalRepo := journalinfra.NewJournalRepository(db)
	journalSettingsRepo := journalinfra.NewJournalSettingsRepository(db)
	journalAccessLogRepo := journalinfra.NewJournalAIAccessLogRepository(db)
	dailyTaskRepo := dailytaskinfra.NewDailyTaskRepository(db)
	notificationRepo := notificationinfra.NewNotificationRepository(db)
	rewardRepo := rewardinfra.NewRewardRepository(db)
	pushSubRepo := pushinfra.NewPushSubscriptionRepository(db)
	guildRepo := guildinfra.NewGuildRepository(db)
	progressMapRepo := progressmapinfra.NewProgressMapRepository(db)
	weeklyLeagueRepo := weeklyleagueinfra.NewWeeklyLeagueRepository(db)
	xpBoostComboRepo := xpboostinfra.NewXPBoostComboRepository(db)
	mysteryChestRepo := mysterychestinfra.NewMysteryChestRepository(db)
	friendQuestRepo := friendquestinfra.NewFriendQuestRepository(db)
	dailySpinRepo := dailyspininfra.NewDailySpinRepository(db)
	streakSocietyRepo := streaksocietyinfra.NewStreakSocietyRepository(db)
	timedChallengeRepo := timedchallengeinfra.NewTimedChallengeRepository(db)
	broadcastRepo := broadcastinfra.NewBroadcastNotificationRepository(db)
	billingRepo := billinginfra.NewBillingRepository(db)
	b2bRepo := billinginfra.NewB2BRepository(db)

	// === Shared Services ===
	cacheService := cache.NewCacheService()
	gamificationService := gamificationapp.NewGamificationService(db)
	contentContextService := contentctx.NewContentContextService(articleRepo, songRepo, songCategoryRepo, forumRepo)
	userContextCache := userctx.NewUserContextCache(cacheService, moodRepo)

	// === Feature Services ===
	authService := authapp.NewAuthService(userRepo)
	userService := authapp.NewUserService(userRepo)
	aiModerationService := moderationapp.NewAIModerationService(moderationRepo, cfg)
	moderationService := moderationapp.NewModerationService(moderationRepo, userRepo, articleRepo, forumRepo, aiModerationService, gamificationService)
	articleService := articleapp.NewArticleService(articleRepo, articleCategoryRepo, gamificationService, contentContextService, cacheService, moderationService)
	songService := songapp.NewSongService(songRepo, songCategoryRepo, cacheService)
	moodService := moodapp.NewMoodService(moodRepo, userContextCache)
	forumService := forumapp.NewForumService(forumRepo, userRepo, gamificationService, contentContextService)
	forumCategoryService := forumapp.NewForumCategoryService(forumCategoryRepo, cacheService)
	levelConfigService := gamificationapp.NewLevelConfigService(levelConfigRepo, cacheService)
	expHistoryService := gamificationapp.NewExpHistoryService(expHistoryRepo)
	chatService := chatapp.NewChatService(chatSessionRepo, chatMessageRepo, cfg, gamificationService, contentContextService, userContextCache)
	midtransClient := billingapp.NewMidtransClient(cfg.MidtransBaseURL, cfg.MidtransServerKey)
	billingService := billingapp.NewService(
		billingRepo,
		midtransClient,
		billingapp.ServiceConfig{
			MidtransServerKey: cfg.MidtransServerKey,
			DefaultDailyLimit: cfg.ChatDailyMessageLimit,
		},
	)
	b2bService := billingapp.NewB2BService(b2bRepo)
	billingService.SetB2BService(b2bService)
	communityProgressService := gamificationapp.NewCommunityProgressService(communityProgressRepo, levelConfigRepo, featureUnlockRepo, badgeRepo, userRepo)
	featureUnlockService := featureunlockapp.NewFeatureUnlockService(featureUnlockRepo, levelConfigRepo, userRepo)
	badgeService := badgeapp.NewBadgeService(badgeRepo, userRepo, levelConfigRepo)
	notificationService := notificationapp.NewNotificationService(notificationRepo)
	pushService := pushapp.NewPushService(pushSubRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDContact)
	notificationService.SetPushService(pushService)
	b2bService.SetNotificationService(notificationService)
	inspiringStoryService := storyapp.NewInspiringStoryService(inspiringStoryRepo, userRepo, levelConfigRepo, badgeService, gamificationService, notificationService)
	dailyTaskService := dailytaskapp.NewDailyTaskService(dailyTaskRepo, userRepo)
	breathingService := breathingapp.NewBreathingService(breathingRepo, gamificationService, dailyTaskService)
	playlistService := playlistapp.NewPlaylistService(playlistRepo, playlistItemRepo, songRepo)
	journalService := journalapp.NewJournalService(journalRepo, journalSettingsRepo, journalAccessLogRepo, moodRepo, chatService.GetGenAIClient())

	chatService.SetModerationRepo(moderationRepo)
	chatService.SetFolderRepo(chatFolderRepo)
	chatService.SetChatQuotaChecker(billingService)
	chatService.SetJournalRepos(journalRepo, journalSettingsRepo, journalAccessLogRepo)
	chatService.SetContextDependencies(
		userRepo,
		dailyTaskService,
		breathingRepo,
		levelConfigService,
		playlistRepo,
		rewardRepo,
		progressMapRepo,
		badgeRepo,
		guildRepo,
	)

	// === Handlers ===
	authHandler := authhandler.NewAuthHandler(authService, levelConfigService)
	userHandler := authhandler.NewUserHandler(userService, levelConfigService)
	billingHandler := billinghandler.NewBillingHandler(billingService)
	b2bHandler := billinghandler.NewB2BHandler(b2bService)
	articleHandler := articlehandler.NewArticleHandler(articleService)
	chatHandler := chathandler.NewChatHandler(chatService)
	uploadHandler := uploadhandler.NewUploadHandler()
	songHandler := songhandler.NewSongHandler(songService)
	moodHandler := moodhandler.NewMoodHandler(moodService)
	adminHandler := adminhandler.NewAdminHandler(db, userRepo, articleRepo, forumRepo, cacheService, journalService)
	searchHandler := searchhandler.NewSearchHandler(articleRepo, songRepo)
	forumHandler := forumhandler.NewForumHandler(forumService)
	forumCategoryHandler := forumhandler.NewForumCategoryHandler(forumCategoryService)
	levelConfigHandler := gamificationhandler.NewLevelConfigHandler(levelConfigService)
	expHistoryHandler := gamificationhandler.NewExpHistoryHandler(expHistoryService, levelConfigService)
	moderationHandler := moderationhandler.NewModerationHandler(moderationService)
	communityProgressHandler := gamificationhandler.NewCommunityProgressHandler(communityProgressService)
	featureUnlockHandler := featureunlockhandler.NewFeatureUnlockHandler(featureUnlockService)
	badgeHandler := badgehandler.NewBadgeHandler(badgeService)
	inspiringStoryHandler := storyhandler.NewInspiringStoryHandler(inspiringStoryService)
	breathingHandler := breathinghandler.NewBreathingHandler(breathingService)
	playlistHandler := playlisthandler.NewPlaylistHandler(playlistService)
	journalHandler := journalhandler.NewJournalHandler(journalService)
	dailyTaskHandler := dailytaskhandler.NewDailyTaskHandler(dailyTaskService)
	notificationHandler := notificationhandler.NewNotificationHandler(notificationService)
	pushHandler := pushhandler.NewPushHandler(pushService)
	broadcastService := broadcastapp.NewBroadcastService(broadcastRepo, pushSubRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDContact)
	broadcastHandler := broadcasthandler.NewBroadcastHandler(broadcastService)
	broadcastService.StartScheduler()
	rewardService := rewardapp.NewRewardService(rewardRepo, userRepo)
	rewardHandler := rewardhandler.NewRewardHandler(rewardService)
	guildService := guildapp.NewGuildService(guildRepo, userRepo, levelConfigRepo)
	guildHandler := guildhandler.NewGuildHandler(guildService)
	progressMapService := progressmapapp.NewProgressMapService(progressMapRepo, userRepo, levelConfigRepo)
	progressMapHandler := progressmaphandler.NewProgressMapHandler(progressMapService)
	weeklyLeagueService := weeklyleagueapp.NewWeeklyLeagueService(weeklyLeagueRepo, userRepo)
	weeklyLeagueHandler := weeklyleaguehandler.NewWeeklyLeagueHandler(weeklyLeagueService)
	xpBoostComboService := xpboostapp.NewXPBoostComboService(xpBoostComboRepo)
	xpBoostComboHandler := xpboosthandler.NewXPBoostComboHandler(xpBoostComboService)
	mysteryChestService := mysterychestapp.NewMysteryChestService(mysteryChestRepo, userRepo)
	mysteryChestHandler := mysterychesthandler.NewMysteryChestHandler(mysteryChestService)
	friendQuestService := friendquestapp.NewFriendQuestService(friendQuestRepo, userRepo)
	friendQuestHandler := friendquesthandler.NewFriendQuestHandler(friendQuestService)
	dailySpinService := dailyspinapp.NewDailySpinService(dailySpinRepo, userRepo)
	dailySpinHandler := dailyspinhandler.NewDailySpinHandler(dailySpinService)
	streakSocietyService := streaksocietyapp.NewStreakSocietyService(streakSocietyRepo, userRepo)
	streakSocietyHandler := streaksocietyhandler.NewStreakSocietyHandler(streakSocietyService)
	timedChallengeService := timedchallengeapp.NewTimedChallengeService(timedChallengeRepo, userRepo)
	timedChallengeHandler := timedchallengehandler.NewTimedChallengeHandler(timedChallengeService)

	wireDailyTaskService(
		dailyTaskService,
		moodHandler,
		chatHandler,
		journalHandler,
		forumHandler,
		articleHandler,
		songHandler,
	)

	return &routeDependencies{
		cacheService: cacheService,

		authHandler:              authHandler,
		userHandler:              userHandler,
		billingHandler:           billingHandler,
		b2bHandler:               b2bHandler,
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

func wireDailyTaskService(
	dailyTaskService dailytaskapp.DailyTaskService,
	moodHandler *moodhandler.MoodHandler,
	chatHandler *chathandler.ChatHandler,
	journalHandler *journalhandler.JournalHandler,
	forumHandler *forumhandler.ForumHandler,
	articleHandler *articlehandler.ArticleHandler,
	songHandler *songhandler.SongHandler,
) {
	if dailyTaskService == nil {
		return
	}
	moodHandler.SetDailyTaskService(dailyTaskService)
	chatHandler.SetDailyTaskService(dailyTaskService)
	journalHandler.SetDailyTaskService(dailyTaskService)
	forumHandler.SetDailyTaskService(dailyTaskService)
	articleHandler.SetDailyTaskService(dailyTaskService)
	songHandler.SetDailyTaskService(dailyTaskService)
}
