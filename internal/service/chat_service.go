package service

import (
	"context"
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ChatService struct {
	sessionRepo           *repository.ChatSessionRepository
	messageRepo           *repository.ChatMessageRepository
	folderRepo            *repository.ChatFolderRepository
	moderationRepo        *repository.ModerationRepository
	journalRepo           *repository.JournalRepository
	journalSettingsRepo   *repository.JournalSettingsRepository
	journalAccessLogRepo  *repository.JournalAIAccessLogRepository
	genaiClient           *genai.Client
	genaiModel            *genai.GenerativeModel
	gamificationService   *GamificationService
	contentContextService *ContentContextService
}

func NewChatService(sessionRepo *repository.ChatSessionRepository, messageRepo *repository.ChatMessageRepository, cfg *config.Config, gamificationService *GamificationService, contentContextService *ContentContextService) *ChatService {
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
	}
}

func (s *ChatService) SetFolderRepo(repo *repository.ChatFolderRepository) {
	s.folderRepo = repo
}

func (s *ChatService) SetModerationRepo(repo *repository.ModerationRepository) {
	s.moderationRepo = repo
}

func (s *ChatService) GetGenAIClient() *genai.Client {
	return s.genaiClient
}

func (s *ChatService) SetJournalRepos(journalRepo *repository.JournalRepository, settingsRepo *repository.JournalSettingsRepository, accessLogRepo *repository.JournalAIAccessLogRepository) {
	s.journalRepo = journalRepo
	s.journalSettingsRepo = settingsRepo
	s.journalAccessLogRepo = accessLogRepo
}
