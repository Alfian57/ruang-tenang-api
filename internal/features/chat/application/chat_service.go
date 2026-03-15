package application

import (
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	journalinfra "github.com/Alfian57/ruang-tenang-api/internal/features/journal/infrastructure"
	moderationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/moderation/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/contentctx"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/userctx"
	"context"
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/Alfian57/ruang-tenang-api/internal/features/chat/infrastructure")

type ChatService struct {
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
	contentContextService *contentctx.ContentContextService
	userContextCache      *userctx.UserContextCache
}

func NewChatService(sessionRepo *infrastructure.ChatSessionRepository, messageRepo *infrastructure.ChatMessageRepository, cfg *config.Config, gamificationService *gamificationapp.GamificationService, contentContextService *contentctx.ContentContextService, userContextCache *userctx.UserContextCache) *ChatService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	var model *genai.GenerativeModel
	if err == nil {
		model = client.GenerativeModel("gemini-flash-latest")
	} else {
		fmt.Printf("Failed to create Gemini client: %v\n", err)
	}

	return &ChatService{
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

func (s *ChatService) GetGenAIClient() *genai.Client {
	return s.genaiClient
}

func (s *ChatService) SetJournalRepos(journalRepo *journalinfra.JournalRepository, settingsRepo *journalinfra.JournalSettingsRepository, accessLogRepo *journalinfra.JournalAIAccessLogRepository) {
	s.journalRepo = journalRepo
	s.journalSettingsRepo = settingsRepo
	s.journalAccessLogRepo = accessLogRepo
}

func (s *ChatService) generateContent(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error) {
	if s.generateContentFn != nil {
		return s.generateContentFn(ctx, prompt)
	}
	return s.genaiModel.GenerateContent(ctx, genai.Text(prompt))
}

