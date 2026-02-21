package service

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
)

// JournalService handles business logic for journals
type JournalService struct {
	journalRepo       *repository.JournalRepository
	settingsRepo      *repository.JournalSettingsRepository
	accessLogRepo     *repository.JournalAIAccessLogRepository
	userMoodRepo      *repository.UserMoodRepository
	genaiClient       *genai.Client
	generateContentFn func(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error)
	searchByContentFn func(ctx context.Context, userID uint, query string, limit int) ([]model.Journal, error)
}

// NewJournalService creates a new JournalService instance
func NewJournalService(
	journalRepo *repository.JournalRepository,
	settingsRepo *repository.JournalSettingsRepository,
	accessLogRepo *repository.JournalAIAccessLogRepository,
	userMoodRepo *repository.UserMoodRepository,
	genaiClient *genai.Client,
) *JournalService {
	return &JournalService{
		journalRepo:   journalRepo,
		settingsRepo:  settingsRepo,
		accessLogRepo: accessLogRepo,
		userMoodRepo:  userMoodRepo,
		genaiClient:   genaiClient,
	}
}
