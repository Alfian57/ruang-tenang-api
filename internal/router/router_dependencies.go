package router

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"

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
	wellnessapp "github.com/Alfian57/ruang-tenang-api/internal/features/wellness/application"
	wellnessinfra "github.com/Alfian57/ruang-tenang-api/internal/features/wellness/infrastructure"
	wellnesshandler "github.com/Alfian57/ruang-tenang-api/internal/features/wellness/interface/http"
	xpboostapp "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/application"
	xpboostinfra "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/infrastructure"
	xpboosthandler "github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/interface/http"

	"gorm.io/gorm"
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
	wellnessHandler          *wellnesshandler.WellnessHandler
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
	wellnessRepo := wellnessinfra.NewWellnessRepository(db)
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
			ResetInterval:     cfg.ChatQuotaResetInterval,
		},
	)
	b2bService := billingapp.NewB2BService(b2bRepo)
	billingService.SetB2BService(b2bService)
	// Banning/suspending a user should release any B2B seat they occupy.
	moderationService.SetSeatReleaser(b2bService)
	communityProgressService := gamificationapp.NewCommunityProgressService(communityProgressRepo, levelConfigRepo, featureUnlockRepo, badgeRepo, userRepo)
	featureUnlockService := featureunlockapp.NewFeatureUnlockService(featureUnlockRepo, levelConfigRepo, userRepo)
	badgeService := badgeapp.NewBadgeService(badgeRepo, userRepo, levelConfigRepo)
	notificationService := notificationapp.NewNotificationService(notificationRepo)
	pushService := pushapp.NewPushService(pushSubRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDContact)
	notificationService.SetPushService(pushService)
	b2bService.SetNotificationService(notificationService)
	// Wire level-up side effects: when AwardExp pushes a user across a level
	// boundary, unlock the new level's features and send a level-up notification.
	gamificationService.SetLevelUpHandlers(levelConfigRepo, featureUnlockService, notificationService)
	inspiringStoryService := storyapp.NewInspiringStoryService(inspiringStoryRepo, userRepo, levelConfigRepo, badgeService, gamificationService, notificationService)
	dailyTaskService := dailytaskapp.NewDailyTaskService(dailyTaskRepo, userRepo)
	breathingService := breathingapp.NewBreathingService(breathingRepo, gamificationService, dailyTaskService)
	playlistService := playlistapp.NewPlaylistService(playlistRepo, playlistItemRepo, songRepo)
	journalService := journalapp.NewJournalService(journalRepo, journalSettingsRepo, journalAccessLogRepo, moodRepo, chatService.GetGenAIClient())
	// Gate public journals through AI moderation before they reach the community feed.
	journalService.SetModerator(aiModerationService)
	wellnessService := wellnessapp.NewWellnessService(wellnessRepo, chatService.GetGenAIClient())

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
	wellnessHandler := wellnesshandler.NewWellnessHandler(wellnessService)

	wireDailyTaskService(
		dailyTaskService,
		moodHandler,
		chatHandler,
		journalHandler,
		forumHandler,
		articleHandler,
		songHandler,
	)

	// Enforce ban/block/suspension on every authenticated request (not just at
	// login). Cached briefly to avoid a DB hit on each call.
	middleware.SetAccountStatusResolver(buildAccountStatusResolver(userRepo, cacheService))
	middleware.SetAccountStatusInvalidator(func(userID uint) {
		cacheService.Delete("account_status:" + strconv.FormatUint(uint64(userID), 10))
	})

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
		wellnessHandler:          wellnessHandler,
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

// buildAccountStatusResolver returns a per-request access checker used by the
// auth middleware. Results are cached briefly (per user) to avoid a DB query
// on every authenticated request while still reflecting bans within seconds.
func buildAccountStatusResolver(
	userRepo *authinfra.UserRepository,
	cacheService *cache.CacheService,
) middleware.AccountStatusResolver {
	const ttl = 15 * time.Second

	return func(userID uint) middleware.AccountStatus {
		cacheKey := "account_status:" + strconv.FormatUint(uint64(userID), 10)
		if cached := cacheService.Get(cacheKey); cached != nil {
			if status, ok := cached.(middleware.AccountStatus); ok {
				return status
			}
		}

		user, err := userRepo.FindByID(context.Background(), userID)
		if err != nil {
			// A not-found result means the account was deleted (soft-deleted) —
			// deny access. Other (transient) errors fail open so a DB blip
			// doesn't lock everyone out.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				denied := middleware.AccountStatus{Allowed: false, Reason: "Akun tidak ditemukan atau telah dihapus."}
				cacheService.SetWithTTL(cacheKey, denied, ttl)
				return denied
			}
			return middleware.AccountStatus{Allowed: true}
		}
		if user == nil {
			return middleware.AccountStatus{Allowed: false, Reason: "Akun tidak ditemukan atau telah dihapus."}
		}

		status := middleware.AccountStatus{Allowed: true}
		switch {
		case user.IsBanned:
			status = middleware.AccountStatus{Allowed: false, Reason: "Akun Anda telah dibanned. Silakan hubungi administrator atau ajukan banding."}
		case user.IsBlocked:
			status = middleware.AccountStatus{Allowed: false, Reason: "Akun Anda telah diblokir. Silakan hubungi administrator."}
		case user.IsSuspended():
			status = middleware.AccountStatus{Allowed: false, Reason: "Akun Anda sedang disuspend. Silakan coba lagi nanti."}
		}

		cacheService.SetWithTTL(cacheKey, status, ttl)
		return status
	}
}
