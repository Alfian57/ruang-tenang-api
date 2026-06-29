package application

import (
	moodinfra "github.com/Alfian57/ruang-tenang-api/internal/features/mood/infrastructure"
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"

	"github.com/Alfian57/ruang-tenang-api/internal/features/journal/infrastructure")

// journalModerator gates public journals through AI moderation.
// Implemented by the moderation feature's AIModerationService; injected via a
// setter to avoid an import cycle between journal and moderation packages.
type journalModerator interface {
	ModerateArticle(ctx context.Context, title, content string) (*dto.AIModerationResult, error)
}

// JournalService handles business logic for journals
type JournalService struct {
	journalRepo       *infrastructure.JournalRepository
	settingsRepo      *infrastructure.JournalSettingsRepository
	accessLogRepo     *infrastructure.JournalAIAccessLogRepository
	userMoodRepo      *moodinfra.UserMoodRepository
	genaiClient       *genai.Client
	moderator         journalModerator
	generateContentFn func(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error)
	searchByContentFn func(ctx context.Context, userID uint, query string, limit int) ([]model.Journal, error)
}

// NewJournalService creates a new JournalService instance
func NewJournalService(
	journalRepo *infrastructure.JournalRepository,
	settingsRepo *infrastructure.JournalSettingsRepository,
	accessLogRepo *infrastructure.JournalAIAccessLogRepository,
	userMoodRepo *moodinfra.UserMoodRepository,
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

// SetModerator injects the AI moderation collaborator used to gate public journals.
func (s *JournalService) SetModerator(m journalModerator) {
	s.moderator = m
}
