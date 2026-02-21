package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"
)

// GetSettings gets journal settings for a user
func (s *JournalService) GetSettings(ctx context.Context, userID uint) (*dto.JournalSettingsResponse, error) {
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalEntries, _ := s.journalRepo.CountByUserID(ctx, userID)
	sharedCount, _ := s.journalRepo.CountSharedWithAI(ctx, userID)

	return &dto.JournalSettingsResponse{
		AllowAIAccess:       settings.AllowAIAccess,
		AIContextDays:       settings.AIContextDays,
		AIContextMaxEntries: settings.AIContextMaxEntries,
		DefaultShareWithAI:  settings.DefaultShareWithAI,
		TotalEntries:        int(totalEntries),
		SharedWithAICount:   int(sharedCount),
		IsBlocked:           settings.IsBlocked,
	}, nil
}

// UpdateSettings updates journal settings
func (s *JournalService) UpdateSettings(ctx context.Context, userID uint, req dto.JournalSettingsRequest) (*dto.JournalSettingsResponse, error) {
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.AllowAIAccess != nil {
		settings.AllowAIAccess = *req.AllowAIAccess
	}
	if req.AIContextDays != nil {
		settings.AIContextDays = *req.AIContextDays
	}
	if req.AIContextMaxEntries != nil {
		settings.AIContextMaxEntries = *req.AIContextMaxEntries
	}
	if req.DefaultShareWithAI != nil {
		settings.DefaultShareWithAI = *req.DefaultShareWithAI
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return s.GetSettings(ctx, userID)
}

// ToggleJournalBlock blocks/unblocks a user from using journals
func (s *JournalService) ToggleJournalBlock(ctx context.Context, userID uint) (*dto.JournalSettingsResponse, error) {
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	settings.IsBlocked = !settings.IsBlocked
	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return s.GetSettings(ctx, userID)
}

// GetAIContext gets journal context for AI chatbot
func (s *JournalService) GetAIContext(ctx context.Context, userID uint, chatSessionID *uint, req dto.JournalAIContextRequest) (*dto.JournalAIContext, error) {
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !settings.AllowAIAccess {
		return &dto.JournalAIContext{HasAccess: false, EntriesCount: 0}, nil
	}

	maxEntries := settings.AIContextMaxEntries
	if req.MaxEntries > 0 {
		maxEntries = req.MaxEntries
	}

	daysBack := settings.AIContextDays
	if req.DaysBack > 0 {
		daysBack = req.DaysBack
	}

	var journals []model.Journal
	if req.Query != "" {
		journals, err = s.journalRepo.FindRelevantForAIContext(ctx, userID, req.Query, maxEntries)
	} else {
		journals, err = s.journalRepo.FindForAIContext(ctx, userID, daysBack, maxEntries)
	}
	if err != nil {
		return nil, err
	}

	aiContext := &dto.JournalAIContext{
		HasAccess:    true,
		EntriesCount: len(journals),
		Entries:      make([]dto.JournalAIContextEntry, 0, len(journals)),
		RecentMoods:  make([]string, 0),
		CommonTags:   make([]string, 0),
	}

	moodSet := make(map[string]bool)
	tagCount := make(map[string]int)

	for _, journal := range journals {
		s.logAIAccess(ctx, userID, journal.ID, chatSessionID, "full")
		s.journalRepo.UpdateAIAccessedAt(ctx, journal.ID)

		entry := dto.JournalAIContextEntry{
			ID:        journal.ID,
			Title:     journal.Title,
			Content:   s.truncateContent(ctx, journal.Content, 500),
			CreatedAt: journal.CreatedAt,
		}

		if journal.Mood != nil {
			entry.Mood = string(journal.Mood.Mood)
			moodSet[entry.Mood] = true
		}

		if len(journal.Tags) > 0 {
			entry.Tags = []string(journal.Tags)
			for _, tag := range journal.Tags {
				tagCount[tag]++
			}
		}

		aiContext.Entries = append(aiContext.Entries, entry)
	}

	for mood := range moodSet {
		aiContext.RecentMoods = append(aiContext.RecentMoods, mood)
	}

	type tagFreq struct {
		tag   string
		count int
	}

	tagFreqs := make([]tagFreq, 0, len(tagCount))
	for tag, count := range tagCount {
		tagFreqs = append(tagFreqs, tagFreq{tag: tag, count: count})
	}
	sort.Slice(tagFreqs, func(i, j int) bool {
		return tagFreqs[i].count > tagFreqs[j].count
	})
	for i := 0; i < len(tagFreqs) && i < 5; i++ {
		aiContext.CommonTags = append(aiContext.CommonTags, tagFreqs[i].tag)
	}

	if len(journals) > 0 {
		aiContext.LastEntryDate = &journals[0].CreatedAt
	}

	if req.IncludeSummary && len(journals) > 0 {
		summary, _ := s.generateJournalSummary(ctx, journals)
		aiContext.Summary = summary
	}

	return aiContext, nil
}

// GetAIAccessLogs gets AI access logs for a user
func (s *JournalService) GetAIAccessLogs(ctx context.Context, userID uint, limit int) ([]dto.JournalAIAccessLogResponse, error) {
	logs, err := s.accessLogRepo.FindByUserID(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.JournalAIAccessLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = dto.JournalAIAccessLogResponse{
			ID:            log.ID,
			JournalID:     log.JournalID,
			JournalTitle:  log.Journal.Title,
			ChatSessionID: log.ChatSessionID,
			ContextType:   log.ContextType,
			AccessedAt:    log.AccessedAt,
		}
	}

	return responses, nil
}

func (s *JournalService) truncateContent(_ context.Context, content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}

	return content[:maxLen] + "..."
}

func (s *JournalService) logAIAccess(ctx context.Context, userID, journalID uint, chatSessionID *uint, contextType string) {
	log := &model.JournalAIAccessLog{
		UserID:        userID,
		JournalID:     journalID,
		ChatSessionID: chatSessionID,
		ContextType:   contextType,
		AccessedAt:    time.Now(),
	}
	_ = s.accessLogRepo.Create(ctx, log)
}

func (s *JournalService) generateJournalSummary(ctx context.Context, journals []model.Journal) (string, error) {
	if s.genaiClient == nil || len(journals) == 0 {
		return "", nil
	}

	var contentBuilder strings.Builder
	for _, journal := range journals {
		contentBuilder.WriteString(fmt.Sprintf(
			"Entry [%s]: %s\n",
			journal.CreatedAt.Format("2006-01-02"),
			s.truncateContent(ctx, journal.Content, 300),
		))
	}

	model := s.genaiClient.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.7)

	prompt := fmt.Sprintf(`Kamu adalah asisten kesehatan mental. Berikan ringkasan singkat (2-3 kalimat) dari entri jurnal berikut dalam Bahasa Indonesia. Fokus pada tema utama dan pola emosional. Jangan menyebutkan hal sensitif secara eksplisit.

Entri Jurnal:
%s

Ringkasan:`, contentBuilder.String())

	var (
		resp *genai.GenerateContentResponse
		err  error
	)
	if s.generateContentFn != nil {
		resp, err = s.generateContentFn(context.Background(), prompt)
	} else {
		resp, err = model.GenerateContent(context.Background(), genai.Text(prompt))
	}
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(text), nil
		}
	}

	return "", nil
}
