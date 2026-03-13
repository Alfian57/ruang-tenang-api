package router

import (
	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"

	// Shared
	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/contentctx"

	// Features
	adminhandler "github.com/Alfian57/ruang-tenang-api/internal/features/admin/interface/http"
	articleapp "github.com/Alfian57/ruang-tenang-api/internal/features/article/application"
	articlehandler "github.com/Alfian57/ruang-tenang-api/internal/features/article/interface/http"
	articleinfra "github.com/Alfian57/ruang-tenang-api/internal/features/article/infrastructure"
	authapp "github.com/Alfian57/ruang-tenang-api/internal/features/auth/application"
	authhandler "github.com/Alfian57/ruang-tenang-api/internal/features/auth/interface/http"
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	badgeapp "github.com/Alfian57/ruang-tenang-api/internal/features/badge/application"
	badgehandler "github.com/Alfian57/ruang-tenang-api/internal/features/badge/interface/http"
	badgeinfra "github.com/Alfian57/ruang-tenang-api/internal/features/badge/infrastructure"
	breathingapp "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/application"
	breathinghandler "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/interface/http"
	breathinginfra "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/infrastructure"
	broadcastapp "github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/application"
	broadcasthandler "github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/interface/http"
	broadcastinfra "github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/infrastructure"
	chatapp "github.com/Alfian57/ruang-tenang-api/internal/features/chat/application"
	chathandler "github.com/Alfian57/ruang-tenang-api/internal/features/chat/interface/http"
	chatinfra "github.com/Alfian57/ruang-tenang-api/internal/features/chat/infrastructure"
	dailyspinapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/application"
	dailyspinhandler "github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/interface/http"
	dailyspininfra "github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/infrastructure"
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	dailytaskhandler "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/interface/http"
	dailytaskinfra "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/infrastructure"
	featureunlockapp "github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/application"
	featureunlockhandler "github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/interface/http"
	featureunlockinfra "github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/infrastructure"
	forumapp "github.com/Alfian57/ruang-tenang-api/internal/features/forum/application"
	forumhandler "github.com/Alfian57/ruang-tenang-api/internal/features/forum/interface/http"
	foruminfra "github.com/Alfian57/ruang-tenang-api/internal/features/forum/infrastructure"
	friendquestapp "github.com/Alfian57/ruang-tenang-api/internal/features/friend_quest/application"
	friendquesthandler "github.com/Alfian57/ruang-tenang-api/internal/features/friend_quest/interface/http"
	friendquestinfra "github.com/Alfian57/ruang-tenang-api/internal/features/friend_quest/infrastructure"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	gamificationhandler "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/interface/http"
	gamificationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/infrastructure"
	guildapp "github.com/Alfian57/ruang-tenang-api/internal/features/guild/application"
	guildhandler "github.com/Alfian57/ruang-tenang-api/internal/features/guild/interface/http"
	guildinfra "github.com/Alfian57/ruang-tenang-api/internal/features/guild/infrastructure"
	journalapp "github.com/Alfian57/ruang-tenang-api/internal/features/journal/application"
	journalhandler "github.com/Alfian57/ruang-tenang-api/internal/features/journal/interface/http"
	journalinfra "github.com/Alfian57/ruang-tenang-api/internal/features/journal/infrastructure"
	moderationapp "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/application"
	moderationhandler "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/interface/http"
	moderationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/infrastructure"
	moodapp "github.com/Alfian57/ruang-tenang-api/internal/features/mood/application"
	moodhandler "github.com/Alfian57/ruang-tenang-api/internal/features/mood/interface/http"
	moodinfra "github.com/Alfian57/ruang-tenang-api/internal/features/mood/infrastructure"
	mysterychestapp "github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/application"
	mysterychesthandler "github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/interface/http"
	mysterychestinfra "github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/infrastructure"
	notificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/notification/application"
	notificationhandler "github.com/Alfian57/ruang-tenang-api/internal/features/notification/interface/http"
	notificationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/notification/infrastructure"
	playlistapp "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/application"
	playlisthandler "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/interface/http"
	playlistinfra "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/infrastructure"
	progressmapapp "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/application"
	progressmaphandler "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/interface/http"
	progressmapinfra "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/infrastructure"
	pushapp "github.com/Alfian57/ruang-tenang-api/internal/features/push/application"
	pushhandler "github.com/Alfian57/ruang-tenang-api/internal/features/push/interface/http"
	pushinfra "github.com/Alfian57/ruang-tenang-api/internal/features/push/infrastructure"
	rewardapp "github.com/Alfian57/ruang-tenang-api/internal/features/reward/application"
	rewardhandler "github.com/Alfian57/ruang-tenang-api/internal/features/reward/interface/http"
	rewardinfra "github.com/Alfian57/ruang-tenang-api/internal/features/reward/infrastructure"
	searchhandler "github.com/Alfian57/ruang-tenang-api/internal/features/search/interface/http"
	songapp "github.com/Alfian57/ruang-tenang-api/internal/features/song/application"
	songhandler "github.com/Alfian57/ruang-tenang-api/internal/features/song/interface/http"
	songinfra "github.com/Alfian57/ruang-tenang-api/internal/features/song/infrastructure"
	storyapp "github.com/Alfian57/ruang-tenang-api/internal/features/story/application"
	storyhandler "github.com/Alfian57/ruang-tenang-api/internal/features/story/interface/http"
	storyinfra "github.com/Alfian57/ruang-tenang-api/internal/features/story/infrastructure"
	streaksocietyapp "github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/application"
	streaksocietyhandler "github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/interface/http"
	streaksocietyinfra "github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/infrastructure"
	timedchallengeapp "github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/application"
	timedchallengehandler "github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/interface/http"
	timedchallengeinfra "github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/infrastructure"
	uploadhandler "github.com/Alfian57/ruang-tenang-api/internal/features/upload/interface/http"
	weeklyleagueapp "github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/application"
	weeklyleaguehandler "github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/interface/http"
	weeklyleagueinfra "github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/infrastructure"
	xpboostapp "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/application"
	xpboosthandler "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/interface/http"
	xpboostinfra "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/infrastructure"
)

type routeDependencies struct {
	cacheService *cache.CacheService

	authHandler              *authhandler.AuthHandler
	userHandler              *authhandler.UserHandler
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

	// === Shared Services ===
	cacheService := cache.NewCacheService()
	gamificationService := gamificationapp.NewGamificationService(db)
	contentContextService := contentctx.NewContentContextService(articleRepo, songRepo, songCategoryRepo, forumRepo)

	// === Feature Services ===
	authService := authapp.NewAuthService(userRepo)
	userService := authapp.NewUserService(userRepo)
	aiModerationService := moderationapp.NewAIModerationService(moderationRepo, cfg)
	moderationService := moderationapp.NewModerationService(moderationRepo, userRepo, articleRepo, forumRepo, aiModerationService)
	articleService := articleapp.NewArticleService(articleRepo, articleCategoryRepo, gamificationService, contentContextService, cacheService, moderationService)
	songService := songapp.NewSongService(songRepo, songCategoryRepo, cacheService)
	moodService := moodapp.NewMoodService(moodRepo)
	forumService := forumapp.NewForumService(forumRepo, userRepo, gamificationService, contentContextService)
	forumCategoryService := forumapp.NewForumCategoryService(forumCategoryRepo, cacheService)
	levelConfigService := gamificationapp.NewLevelConfigService(levelConfigRepo, cacheService)
	expHistoryService := gamificationapp.NewExpHistoryService(expHistoryRepo)
	chatService := chatapp.NewChatService(chatSessionRepo, chatMessageRepo, cfg, gamificationService, contentContextService)
	communityProgressService := gamificationapp.NewCommunityProgressService(communityProgressRepo, levelConfigRepo, featureUnlockRepo, badgeRepo, userRepo)
	featureUnlockService := featureunlockapp.NewFeatureUnlockService(featureUnlockRepo, levelConfigRepo, userRepo)
	badgeService := badgeapp.NewBadgeService(badgeRepo, userRepo, levelConfigRepo)
	notificationService := notificationapp.NewNotificationService(notificationRepo)
	pushService := pushapp.NewPushService(pushSubRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDContact)
	notificationService.SetPushService(pushService)
	inspiringStoryService := storyapp.NewInspiringStoryService(inspiringStoryRepo, userRepo, levelConfigRepo, badgeService, gamificationService, notificationService)
	dailyTaskService := dailytaskapp.NewDailyTaskService(dailyTaskRepo, userRepo)
	breathingService := breathingapp.NewBreathingService(breathingRepo, gamificationService, dailyTaskService)
	playlistService := playlistapp.NewPlaylistService(playlistRepo, playlistItemRepo, songRepo)
	journalService := journalapp.NewJournalService(journalRepo, journalSettingsRepo, journalAccessLogRepo, moodRepo, chatService.GetGenAIClient())

	chatService.SetModerationRepo(moderationRepo)
	chatService.SetFolderRepo(chatFolderRepo)
	chatService.SetJournalRepos(journalRepo, journalSettingsRepo, journalAccessLogRepo)

	// === Handlers ===
	authHandler := authhandler.NewAuthHandler(authService, levelConfigService)
	userHandler := authhandler.NewUserHandler(userService, levelConfigService)
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

	// === Daily task service injection ===
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
