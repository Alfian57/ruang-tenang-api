package application

import (
	"context"
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	badgeinfra "github.com/Alfian57/ruang-tenang-api/internal/features/badge/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/features/chat/infrastructure"
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	journalinfra "github.com/Alfian57/ruang-tenang-api/internal/features/journal/infrastructure"
	moderationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/infrastructure"
	playlistinfra "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/infrastructure"
	progressmapinfra "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/infrastructure"
	rewardinfra "github.com/Alfian57/ruang-tenang-api/internal/features/reward/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/ai"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/contentctx"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/entitlement"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/userctx"
)

var ErrDailyChatQuotaExceeded = errors.New("chat quota exceeded")

type ChatService struct {
	modelName             string
	sessionRepo           *infrastructure.ChatSessionRepository
	messageRepo           *infrastructure.ChatMessageRepository
	folderRepo            *infrastructure.ChatFolderRepository
	moderationRepo        *moderationinfra.ModerationRepository
	journalRepo           *journalinfra.JournalRepository
	journalSettingsRepo   *journalinfra.JournalSettingsRepository
	journalAccessLogRepo  *journalinfra.JournalAIAccessLogRepository
	aiClient              ai.Client
	generateContentFn     func(ctx context.Context, prompt string) (*ai.CompletionResponse, error)
	generateChatReplyFn   func(ctx context.Context, systemPrompt string, history []model.ChatMessage, userInput string) (string, error)
	gamificationService   *gamificationapp.GamificationService
	levelConfigService    *gamificationapp.LevelConfigService
	contentContextService *contentctx.ContentContextService
	userContextCache      *userctx.UserContextCache
	dailyTaskService      dailytaskapp.DailyTaskService
	userRepo              *authinfra.UserRepository
	playlistRepo          *playlistinfra.PlaylistRepository
	rewardRepo            *rewardinfra.RewardRepository
	progressMapRepo       *progressmapinfra.ProgressMapRepository
	badgeRepo             *badgeinfra.BadgeRepository
	chatQuotaChecker      entitlement.ChatQuotaChecker
}

func NewChatService(sessionRepo *infrastructure.ChatSessionRepository, messageRepo *infrastructure.ChatMessageRepository, cfg *config.Config, aiClient ai.Client, gamificationService *gamificationapp.GamificationService, contentContextService *contentctx.ContentContextService, userContextCache *userctx.UserContextCache) *ChatService {
	return &ChatService{
		modelName:             cfg.AI.ChatModel,
		sessionRepo:           sessionRepo,
		messageRepo:           messageRepo,
		aiClient:              aiClient,
		gamificationService:   gamificationService,
		contentContextService: contentContextService,
		userContextCache:      userContextCache,
	}
}

func (s *ChatService) SetFolderRepo(repo *infrastructure.ChatFolderRepository) {
	s.folderRepo = repo
}

func (s *ChatService) SetModerationRepo(repo *moderationinfra.ModerationRepository) {
	s.moderationRepo = repo
}

func (s *ChatService) SetChatQuotaChecker(checker entitlement.ChatQuotaChecker) {
	s.chatQuotaChecker = checker
}

func (s *ChatService) GetAIClient() ai.Client {
	return s.aiClient
}

func (s *ChatService) SetJournalRepos(journalRepo *journalinfra.JournalRepository, settingsRepo *journalinfra.JournalSettingsRepository, accessLogRepo *journalinfra.JournalAIAccessLogRepository) {
	s.journalRepo = journalRepo
	s.journalSettingsRepo = settingsRepo
	s.journalAccessLogRepo = accessLogRepo
}

func (s *ChatService) SetContextDependencies(
	userRepo *authinfra.UserRepository,
	dailyTaskService dailytaskapp.DailyTaskService,
	levelConfigService *gamificationapp.LevelConfigService,
	playlistRepo *playlistinfra.PlaylistRepository,
	rewardRepo *rewardinfra.RewardRepository,
	progressMapRepo *progressmapinfra.ProgressMapRepository,
	badgeRepo *badgeinfra.BadgeRepository,
) {
	s.userRepo = userRepo
	s.dailyTaskService = dailyTaskService
	s.levelConfigService = levelConfigService
	s.playlistRepo = playlistRepo
	s.rewardRepo = rewardRepo
	s.progressMapRepo = progressMapRepo
	s.badgeRepo = badgeRepo
}

func (s *ChatService) generateContent(ctx context.Context, prompt string) (*ai.CompletionResponse, error) {
	if s.generateContentFn != nil {
		return s.generateContentFn(ctx, prompt)
	}
	if s.aiClient == nil || !s.aiClient.IsConfigured() {
		return nil, ai.ErrNotConfigured
	}
	return s.aiClient.Complete(ctx, ai.CompletionRequest{
		Model: s.modelName,
		Messages: []ai.Message{
			{Role: "user", Content: prompt},
		},
	})
}

func (s *ChatService) modelAvailable() bool {
	return s.generateContentFn != nil || s.generateChatReplyFn != nil || (s.aiClient != nil && s.aiClient.IsConfigured())
}
