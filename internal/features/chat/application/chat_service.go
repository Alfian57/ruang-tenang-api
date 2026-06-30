package application

import (
	"context"
	"errors"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	badgeinfra "github.com/Alfian57/ruang-tenang-api/internal/features/badge/infrastructure"
	breathinginfra "github.com/Alfian57/ruang-tenang-api/internal/features/breathing/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/features/chat/infrastructure"
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	guildinfra "github.com/Alfian57/ruang-tenang-api/internal/features/guild/infrastructure"
	journalinfra "github.com/Alfian57/ruang-tenang-api/internal/features/journal/infrastructure"
	moderationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/infrastructure"
	playlistinfra "github.com/Alfian57/ruang-tenang-api/internal/features/playlist/infrastructure"
	progressmapinfra "github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/infrastructure"
	rewardinfra "github.com/Alfian57/ruang-tenang-api/internal/features/reward/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/contentctx"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/entitlement"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/userctx"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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
	genaiClient           *genai.Client
	genaiModel            *genai.GenerativeModel
	generateContentFn     func(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error)
	generateChatReplyFn   func(ctx context.Context, systemPrompt string, history []model.ChatMessage, userInput string) (string, error)
	gamificationService   *gamificationapp.GamificationService
	levelConfigService    *gamificationapp.LevelConfigService
	contentContextService *contentctx.ContentContextService
	userContextCache      *userctx.UserContextCache
	dailyTaskService      dailytaskapp.DailyTaskService
	breathingRepo         breathinginfra.BreathingRepository
	userRepo              *authinfra.UserRepository
	playlistRepo          *playlistinfra.PlaylistRepository
	rewardRepo            *rewardinfra.RewardRepository
	progressMapRepo       *progressmapinfra.ProgressMapRepository
	badgeRepo             *badgeinfra.BadgeRepository
	guildRepo             *guildinfra.GuildRepository
	chatQuotaChecker      entitlement.ChatQuotaChecker
}

func NewChatService(sessionRepo *infrastructure.ChatSessionRepository, messageRepo *infrastructure.ChatMessageRepository, cfg *config.Config, gamificationService *gamificationapp.GamificationService, contentContextService *contentctx.ContentContextService, userContextCache *userctx.UserContextCache) *ChatService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.AI.APIKey))
	modelName := cfg.AI.ChatModel
	var model *genai.GenerativeModel
	if err == nil {
		model = client.GenerativeModel(modelName)
	} else {
		// Avoid logging the raw error here — it may echo API-key details.
		logger.Warn("failed to create Gemini client for chat")
	}

	return &ChatService{
		modelName:             modelName,
		sessionRepo:           sessionRepo,
		messageRepo:           messageRepo,
		genaiClient:           client,
		genaiModel:            model,
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

func (s *ChatService) GetGenAIClient() *genai.Client {
	return s.genaiClient
}

func (s *ChatService) SetJournalRepos(journalRepo *journalinfra.JournalRepository, settingsRepo *journalinfra.JournalSettingsRepository, accessLogRepo *journalinfra.JournalAIAccessLogRepository) {
	s.journalRepo = journalRepo
	s.journalSettingsRepo = settingsRepo
	s.journalAccessLogRepo = accessLogRepo
}

func (s *ChatService) SetContextDependencies(
	userRepo *authinfra.UserRepository,
	dailyTaskService dailytaskapp.DailyTaskService,
	breathingRepo breathinginfra.BreathingRepository,
	levelConfigService *gamificationapp.LevelConfigService,
	playlistRepo *playlistinfra.PlaylistRepository,
	rewardRepo *rewardinfra.RewardRepository,
	progressMapRepo *progressmapinfra.ProgressMapRepository,
	badgeRepo *badgeinfra.BadgeRepository,
	guildRepo *guildinfra.GuildRepository,
) {
	s.userRepo = userRepo
	s.dailyTaskService = dailyTaskService
	s.breathingRepo = breathingRepo
	s.levelConfigService = levelConfigService
	s.playlistRepo = playlistRepo
	s.rewardRepo = rewardRepo
	s.progressMapRepo = progressMapRepo
	s.badgeRepo = badgeRepo
	s.guildRepo = guildRepo
}

func (s *ChatService) generateContent(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error) {
	if s.generateContentFn != nil {
		return s.generateContentFn(ctx, prompt)
	}
	model := s.modelForRequest()
	if model == nil {
		return nil, errors.New("AI service is not configured")
	}
	return model.GenerateContent(ctx, genai.Text(prompt))
}

// modelForRequest returns a fresh GenerativeModel derived from the shared client.
//
// We must NOT mutate s.genaiModel per request: ChatService is a singleton and
// setting s.genaiModel.Tools on one request would leak into concurrent requests
// (a race that could attach RAG tools to unrelated chats, or nil them out
// mid-flight). Deriving a new model from the client is cheap and isolated.
func (s *ChatService) modelForRequest() *genai.GenerativeModel {
	if s.genaiClient == nil {
		return nil
	}
	return s.genaiClient.GenerativeModel(s.modelName)
}
