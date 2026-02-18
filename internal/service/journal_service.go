package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/lib/pq"
)

// JournalService handles business logic for journals
type JournalService struct {
	journalRepo   *repository.JournalRepository
	settingsRepo  *repository.JournalSettingsRepository
	accessLogRepo *repository.JournalAIAccessLogRepository
	userMoodRepo  *repository.UserMoodRepository
	genaiClient   *genai.Client
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

// ===== Journal CRUD =====

// CreateJournal creates a new journal entry
func (s *JournalService) CreateJournal(ctx context.Context, userID uint, req dto.CreateJournalRequest) (*dto.JournalResponse, error) {
	// Get user settings for default share_with_ai
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	shareWithAI := settings.DefaultShareWithAI
	if req.ShareWithAI != nil {
		shareWithAI = *req.ShareWithAI
	}

	// Check if user is blocked from creating journals
	if settings.IsBlocked {
		return nil, errors.New("you are blocked from creating journal entries")
	}

	// Calculate word count
	wordCount := len(strings.Fields(req.Content))

	journal := &model.Journal{
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		MoodID:      req.MoodID,
		Tags:        pq.StringArray(req.Tags),
		IsPrivate:   true, // Always private by default
		ShareWithAI: shareWithAI,
		WordCount:   wordCount,
	}

	if err := s.journalRepo.Create(ctx, journal); err != nil {
		return nil, fmt.Errorf("failed to create journal: %w", err)
	}

	// Generate summary if content is long enough
	if wordCount > 100 {
		go func(jID uint, content string) {
			// Create a background context
			bgCtx := context.Background()
			summary, err := s.generateSingleEntrySummary(bgCtx, content)
			if err == nil && summary != "" {
				_ = s.journalRepo.UpdateSummary(bgCtx, jID, summary)
			}
		}(journal.ID, journal.Content)
	}

	// Reload with mood
	journal, err = s.journalRepo.FindByID(ctx, journal.ID)
	if err != nil {
		return nil, err
	}

	return s.toJournalResponse(ctx, journal), nil
}

// GetJournal gets a journal by ID
func (s *JournalService) GetJournal(ctx context.Context, userID, journalID uint) (*dto.JournalResponse, error) {
	journal, err := s.journalRepo.FindByIDAndUserID(ctx, journalID, userID)
	if err != nil {
		return nil, fmt.Errorf("journal not found: %w", err)
	}
	return s.toJournalResponse(ctx, journal), nil
}

// GetJournalByUUID gets a journal by UUID
func (s *JournalService) GetJournalByUUID(ctx context.Context, userID uint, journalUUID string) (*dto.JournalResponse, error) {
	// Parse UUID
	id, err := uuid.Parse(journalUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	journal, err := s.journalRepo.FindByUUIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return s.toJournalResponse(ctx, journal), nil
}

// UpdateJournal updates a journal entry
func (s *JournalService) UpdateJournal(ctx context.Context, userID, journalID uint, req dto.UpdateJournalRequest) (*dto.JournalResponse, error) {
	// Find journal
	journal, err := s.journalRepo.FindByIDAndUserID(ctx, journalID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Title != nil {
		journal.Title = *req.Title
	}
	if req.Content != nil {
		journal.Content = *req.Content
		// Recalculate word count
		journal.WordCount = len(strings.Fields(journal.Content))
		// Analyze sentiment (mock for now, or use GenAI if implemented)
		// journal.SentimentScore = ...
	}
	if req.MoodID != nil {
		journal.MoodID = req.MoodID
	}
	if len(req.Tags) > 0 {
		journal.Tags = pq.StringArray(req.Tags)
	}
	if req.ShareWithAI != nil {
		journal.ShareWithAI = *req.ShareWithAI
	}

	// Check if user is blocked (prevent updates too)
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err == nil && settings.IsBlocked {
		return nil, errors.New("you are blocked from editing journal entries")
	}

	// Save
	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	// Regenerate summary if content changed and is long enough
	if req.Content != nil && len(strings.Fields(*req.Content)) > 100 {
		go func(jID uint, content string) {
			bgCtx := context.Background()
			summary, err := s.generateSingleEntrySummary(bgCtx, content)
			if err == nil && summary != "" {
				_ = s.journalRepo.UpdateSummary(bgCtx, jID, summary)
			}
		}(journal.ID, *req.Content)
	}

	return s.toJournalResponse(ctx, journal), nil
}

// UpdateJournalByUUID updates a journal entry by UUID
func (s *JournalService) UpdateJournalByUUID(ctx context.Context, userID uint, journalUUID string, req dto.UpdateJournalRequest) (*dto.JournalResponse, error) {
	// Parse UUID
	id, err := uuid.Parse(journalUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	// Find journal
	journal, err := s.journalRepo.FindByUUIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Title != nil {
		journal.Title = *req.Title
	}
	if req.Content != nil {
		journal.Content = *req.Content
		// Recalculate word count
		journal.WordCount = len(strings.Fields(journal.Content))
	}
	if req.MoodID != nil {
		journal.MoodID = req.MoodID
	}
	if len(req.Tags) > 0 {
		journal.Tags = pq.StringArray(req.Tags)
	}
	if req.ShareWithAI != nil {
		journal.ShareWithAI = *req.ShareWithAI
	}

	// Check if user is blocked (prevent updates too)
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err == nil && settings.IsBlocked {
		return nil, errors.New("you are blocked from editing journal entries")
	}

	// Save
	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	// Regenerate summary if content changed and is long enough
	if req.Content != nil && len(strings.Fields(*req.Content)) > 100 {
		go func(jID uint, content string) {
			bgCtx := context.Background()
			summary, err := s.generateSingleEntrySummary(bgCtx, content)
			if err == nil && summary != "" {
				_ = s.journalRepo.UpdateSummary(bgCtx, jID, summary)
			}
		}(journal.ID, *req.Content)
	}

	return s.toJournalResponse(ctx, journal), nil
}

// DeleteJournal deletes a journal entry
func (s *JournalService) DeleteJournal(ctx context.Context, userID, journalID uint) error {
	// Check existence and ownership
	_, err := s.journalRepo.FindByIDAndUserID(ctx, journalID, userID)
	if err != nil {
		return err
	}

	return s.journalRepo.Delete(ctx, journalID, userID)
}

// DeleteJournalByUUID deletes a journal entry by UUID
func (s *JournalService) DeleteJournalByUUID(ctx context.Context, userID uint, journalUUID string) error {
	// Parse UUID
	id, err := uuid.Parse(journalUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	// Check existence and ownership
	journal, err := s.journalRepo.FindByUUIDAndUserID(ctx, id, userID)
	if err != nil {
		return err
	}

	return s.journalRepo.Delete(ctx, journal.ID, userID)
}

// ListJournals lists journals for a user
func (s *JournalService) ListJournals(ctx context.Context, userID uint, page, limit int, tags []string, moodID *uint, startDate, endDate *time.Time) ([]dto.JournalListResponse, int64, error) {
	journals, total, err := s.journalRepo.FindByUserID(ctx, userID, page, limit, tags, moodID, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.JournalListResponse, len(journals))
	for i, j := range journals {
		responses[i] = s.toJournalListResponse(ctx, &j)
	}

	return responses, total, nil
}

// SearchJournals searches journals by content
func (s *JournalService) SearchJournals(ctx context.Context, userID uint, query string, limit int) ([]dto.JournalListResponse, error) {
	journals, err := s.journalRepo.SearchByContent(ctx, userID, query, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.JournalListResponse, len(journals))
	for i, j := range journals {
		responses[i] = s.toJournalListResponse(ctx, &j)
	}

	return responses, nil
}

// ===== Settings =====

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

// ===== AI Context Integration =====

// GetAIContext gets journal context for AI chatbot
func (s *JournalService) GetAIContext(ctx context.Context, userID uint, chatSessionID *uint, req dto.JournalAIContextRequest) (*dto.JournalAIContext, error) {
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user allows AI access globally
	if !settings.AllowAIAccess {
		return &dto.JournalAIContext{
			HasAccess:    false,
			EntriesCount: 0,
		}, nil
	}

	// Determine parameters
	maxEntries := settings.AIContextMaxEntries
	if req.MaxEntries > 0 {
		maxEntries = req.MaxEntries
	}

	daysBack := settings.AIContextDays
	if req.DaysBack > 0 {
		daysBack = req.DaysBack
	}

	var journals []model.Journal

	// If query provided, search for relevant entries
	if req.Query != "" {
		journals, err = s.journalRepo.FindRelevantForAIContext(ctx, userID, req.Query, maxEntries)
	} else {
		journals, err = s.journalRepo.FindForAIContext(ctx, userID, daysBack, maxEntries)
	}

	if err != nil {
		return nil, err
	}

	// Build context
	context := &dto.JournalAIContext{
		HasAccess:    true,
		EntriesCount: len(journals),
		Entries:      make([]dto.JournalAIContextEntry, 0, len(journals)),
		RecentMoods:  make([]string, 0),
		CommonTags:   make([]string, 0),
	}

	moodSet := make(map[string]bool)
	tagCount := make(map[string]int)

	for _, j := range journals {
		// Log AI access for transparency
		s.logAIAccess(ctx, userID, j.ID, chatSessionID, "full")

		// Update AI accessed timestamp
		s.journalRepo.UpdateAIAccessedAt(ctx, j.ID)

		entry := dto.JournalAIContextEntry{
			ID:        j.ID,
			Title:     j.Title,
			Content:   s.truncateContent(ctx, j.Content, 500), // Truncate for context window
			CreatedAt: j.CreatedAt,
		}

		if j.Mood != nil {
			entry.Mood = string(j.Mood.Mood)
			moodSet[entry.Mood] = true
		}

		if len(j.Tags) > 0 {
			entry.Tags = []string(j.Tags)
			for _, tag := range j.Tags {
				tagCount[tag]++
			}
		}

		context.Entries = append(context.Entries, entry)
	}

	// Extract recent moods
	for mood := range moodSet {
		context.RecentMoods = append(context.RecentMoods, mood)
	}

	// Extract common tags (top 5)
	type tagFreq struct {
		tag   string
		count int
	}
	var tagFreqs []tagFreq
	for tag, count := range tagCount {
		tagFreqs = append(tagFreqs, tagFreq{tag, count})
	}
	sort.Slice(tagFreqs, func(i, j int) bool {
		return tagFreqs[i].count > tagFreqs[j].count
	})
	for i := 0; i < len(tagFreqs) && i < 5; i++ {
		context.CommonTags = append(context.CommonTags, tagFreqs[i].tag)
	}

	// Set last entry date
	if len(journals) > 0 {
		context.LastEntryDate = &journals[0].CreatedAt
	}

	// Generate summary if requested
	if req.IncludeSummary && len(journals) > 0 {
		summary, _ := s.generateJournalSummary(ctx, journals)
		context.Summary = summary
	}

	return context, nil
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

// ===== Analytics =====

// GetAnalytics gets journal analytics for a user
func (s *JournalService) GetAnalytics(ctx context.Context, userID uint) (*dto.JournalAnalytics, error) {
	totalEntries, _ := s.journalRepo.CountByUserID(ctx, userID)
	totalWordCount, _ := s.journalRepo.GetTotalWordCount(ctx, userID)
	moodDistribution, _ := s.journalRepo.GetMoodDistribution(ctx, userID)
	tagFrequency, _ := s.journalRepo.GetTagFrequency(ctx, userID)
	entriesByMonth, _ := s.journalRepo.GetEntriesByMonth(ctx, userID, 12)
	writingStreak, _ := s.journalRepo.GetWritingStreak(ctx, userID)

	avgWordCount := 0
	if totalEntries > 0 {
		avgWordCount = totalWordCount / int(totalEntries)
	}

	// Calculate entries this month
	entriesThisMonth := 0
	currentMonth := time.Now().Format("2006-01")
	for _, e := range entriesByMonth {
		if e.Month == currentMonth {
			entriesThisMonth = e.Count
			break
		}
	}

	monthlyEntries := make([]dto.MonthlyEntryCount, len(entriesByMonth))
	for i, e := range entriesByMonth {
		monthlyEntries[i] = dto.MonthlyEntryCount{
			Month: e.Month,
			Count: e.Count,
		}
	}

	return &dto.JournalAnalytics{
		TotalEntries:     int(totalEntries),
		EntriesThisMonth: entriesThisMonth,
		TotalWordCount:   totalWordCount,
		AvgWordCount:     avgWordCount,
		MoodDistribution: moodDistribution,
		TagFrequency:     tagFrequency,
		EntriesByMonth:   monthlyEntries,
		WritingStreak:    writingStreak,
		LongestStreak:    writingStreak, // TODO: Track longest streak separately
	}, nil
}

// ===== AI-Powered Features =====

// GetWritingPrompt generates an AI writing prompt
func (s *JournalService) GetWritingPrompt(ctx context.Context, userID uint) (*dto.JournalPromptResponse, error) {
	// Get user's recent mood
	latestMood, _ := s.userMoodRepo.GetLatestByUserID(ctx, userID)

	// Get recent tags
	tagFrequency, _ := s.journalRepo.GetTagFrequency(ctx, userID)

	// Build prompt based on context
	var moodContext string
	if latestMood != nil {
		moodContext = fmt.Sprintf("User's current mood is: %s (%s)", latestMood.Mood, latestMood.GetMoodEmoji())
	}

	var topTags []string
	for tag := range tagFrequency {
		topTags = append(topTags, tag)
		if len(topTags) >= 3 {
			break
		}
	}

	prompt := s.generateWritingPrompt(ctx, moodContext, topTags)

	return prompt, nil
}

// GetWeeklySummary generates a weekly summary of journals
func (s *JournalService) GetWeeklySummary(ctx context.Context, userID uint) (*dto.JournalWeeklySummary, error) {
	weekStart := time.Now().AddDate(0, 0, -7)
	weekEnd := time.Now()

	journals, _, err := s.journalRepo.FindByUserID(ctx, userID, 1, 100, nil, nil, &weekStart, &weekEnd)
	if err != nil {
		return nil, err
	}

	if len(journals) == 0 {
		return &dto.JournalWeeklySummary{
			WeekStart:    weekStart,
			WeekEnd:      weekEnd,
			EntriesCount: 0,
			Summary:      "Tidak ada entri jurnal minggu ini.",
		}, nil
	}

	// Generate AI summary
	summary, themes, insights, suggestions, moodTrend := s.generateWeeklySummary(ctx, journals)

	return &dto.JournalWeeklySummary{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		EntriesCount: len(journals),
		Summary:      summary,
		KeyThemes:    themes,
		MoodTrend:    moodTrend,
		Insights:     insights,
		Suggestions:  suggestions,
	}, nil
}

// ===== Export =====

// ExportJournals exports journals in the specified format
func (s *JournalService) ExportJournals(ctx context.Context, userID uint, req dto.JournalExportRequest) (*dto.JournalExportResponse, error) {
	var startDate, endDate *time.Time

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			// Set to end of day
			t = t.Add(24*time.Hour - time.Second)
			endDate = &t
		}
	}

	// Pass nil for moodID as export doesn't currently support mood filtering
	journals, _, err := s.journalRepo.FindByUserID(ctx, userID, 1, 1000, req.Tags, nil, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var content string
	var filename string

	switch req.Format {
	case "txt":
		content = s.exportToTXT(ctx, journals)
		filename = fmt.Sprintf("journal_export_%s.txt", time.Now().Format("2006-01-02"))
	case "pdf":
		pdfContent, err := s.exportToPDF(ctx, journals)
		if err != nil {
			return nil, err
		}
		content = base64.StdEncoding.EncodeToString(pdfContent)
		filename = fmt.Sprintf("journal_export_%s.pdf", time.Now().Format("2006-01-02"))
	default:
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	return &dto.JournalExportResponse{
		Format:   req.Format,
		Content:  content,
		Filename: filename,
	}, nil
}

// ===== Helper Methods =====

func (s *JournalService) toJournalResponse(ctx context.Context, j *model.Journal) *dto.JournalResponse {
	response := &dto.JournalResponse{
		ID:    j.ID,
		UUID:  j.UUID.String(),
		Title: j.Title,

		Content:        j.Content,
		Summary:        j.Summary,
		MoodID:         j.MoodID,
		Tags:           []string(j.Tags),
		IsPrivate:      j.IsPrivate,
		ShareWithAI:    j.ShareWithAI,
		AIAccessedAt:   j.AIAccessedAt,
		WordCount:      j.WordCount,
		SentimentScore: j.SentimentScore,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}

	if j.Mood != nil {
		response.MoodLabel = string(j.Mood.Mood)
		response.MoodEmoji = j.Mood.GetMoodEmoji()
	}

	if response.Tags == nil {
		response.Tags = []string{}
	}

	return response
}

func (s *JournalService) toJournalListResponse(ctx context.Context, j *model.Journal) dto.JournalListResponse {
	preview := j.Content
	if len(preview) > 150 {
		preview = preview[:150] + "..."
	}

	response := dto.JournalListResponse{
		ID:           j.ID,
		UUID:         j.UUID.String(),
		Title:        j.Title,
		Preview:      preview,
		MoodID:       j.MoodID,
		Tags:         []string(j.Tags),
		ShareWithAI:  j.ShareWithAI,
		AIAccessedAt: j.AIAccessedAt,
		WordCount:    j.WordCount,
		CreatedAt:    j.CreatedAt,
	}

	if j.Mood != nil {
		response.MoodLabel = string(j.Mood.Mood)
		response.MoodEmoji = j.Mood.GetMoodEmoji()
	}

	if response.Tags == nil {
		response.Tags = []string{}
	}

	return response
}

func (s *JournalService) truncateContent(ctx context.Context, content string, maxLen int) string {
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
	s.accessLogRepo.Create(ctx, log)
}

func (s *JournalService) generateJournalSummary(ctx context.Context, journals []model.Journal) (string, error) {
	if s.genaiClient == nil || len(journals) == 0 {
		return "", nil
	}

	var contentBuilder strings.Builder
	for _, j := range journals {
		contentBuilder.WriteString(fmt.Sprintf("Entry [%s]: %s\n", j.CreatedAt.Format("2006-01-02"), s.truncateContent(ctx, j.Content, 300)))
	}

	model := s.genaiClient.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.7)

	prompt := fmt.Sprintf(`Kamu adalah asisten kesehatan mental. Berikan ringkasan singkat (2-3 kalimat) dari entri jurnal berikut dalam Bahasa Indonesia. Fokus pada tema utama dan pola emosional. Jangan menyebutkan hal sensitif secara eksplisit.

Entri Jurnal:
%s

Ringkasan:`, contentBuilder.String())

	resp, err := model.GenerateContent(context.Background(), genai.Text(prompt))
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

func (s *JournalService) generateSingleEntrySummary(ctx context.Context, content string) (string, error) {
	if s.genaiClient == nil {
		return "", nil
	}

	model := s.genaiClient.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.5)

	prompt := fmt.Sprintf(`Buatlah ringkasan singkat (1 kalimat) dari entri jurnal berikut. Fokus pada inti kejadian atau perasaan utama.
Jurnal: "%s"`, s.truncateContent(ctx, content, 2000))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return strings.TrimSpace(string(text)), nil
		}
	}

	return "", nil
}

func (s *JournalService) generateWritingPrompt(ctx context.Context, moodContext string, topTags []string) *dto.JournalPromptResponse {
	categories := []string{"reflection", "gratitude", "goal", "emotion"}
	prompts := map[string][]string{
		"reflection": {
			"Apa satu hal yang kamu pelajari tentang dirimu hari ini?",
			"Bagaimana perasaanmu saat ini dan mengapa?",
			"Apa yang membuatmu bersyukur minggu ini?",
		},
		"gratitude": {
			"Sebutkan 3 hal kecil yang membuatmu tersenyum hari ini",
			"Siapa yang ingin kamu ucapkan terima kasih, dan mengapa?",
			"Apa momen sederhana yang membuat harimu lebih baik?",
		},
		"goal": {
			"Apa satu langkah kecil yang bisa kamu ambil besok untuk mencapai tujuanmu?",
			"Bagaimana kamu ingin perasaanmu di akhir minggu ini?",
			"Apa yang ingin kamu lepaskan untuk bisa maju?",
		},
		"emotion": {
			"Ceritakan tentang emosi yang paling kuat kamu rasakan hari ini",
			"Apa yang bisa kamu katakan pada dirimu sendiri untuk merasa lebih baik?",
			"Bagaimana tubuhmu merespons stres, dan apa yang bisa membantu?",
		},
	}

	// Simple random selection
	categoryIdx := time.Now().Unix() % int64(len(categories))
	category := categories[categoryIdx]
	promptList := prompts[category]
	promptIdx := time.Now().UnixNano() % int64(len(promptList))

	return &dto.JournalPromptResponse{
		Prompt:      promptList[promptIdx],
		Category:    category,
		RelatedTags: topTags,
	}
}

func (s *JournalService) generateWeeklySummary(ctx context.Context, journals []model.Journal) (string, []string, []string, []string, string) {
	if s.genaiClient == nil || len(journals) == 0 {
		return "Tidak cukup data untuk membuat ringkasan.", []string{}, []string{}, []string{}, "stable"
	}

	var contentBuilder strings.Builder
	for _, j := range journals {
		mood := ""
		if j.Mood != nil {
			mood = string(j.Mood.Mood)
		}
		contentBuilder.WriteString(fmt.Sprintf("[%s] Mood: %s | %s\n", j.CreatedAt.Format("2006-01-02"), mood, s.truncateContent(ctx, j.Content, 200)))
	}

	model := s.genaiClient.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.7)

	prompt := fmt.Sprintf(`Kamu adalah asisten kesehatan mental. Analisis entri jurnal minggu ini dan berikan:
1. Ringkasan singkat (2-3 kalimat)
2. 3 tema utama (pisahkan dengan koma)
3. 2 insight positif (pisahkan dengan |)
4. 2 saran untuk minggu depan (pisahkan dengan |)
5. Tren mood: "improving", "stable", atau "declining"

Format jawaban:
SUMMARY: [ringkasan]
THEMES: [tema1, tema2, tema3]
INSIGHTS: [insight1 | insight2]
SUGGESTIONS: [saran1 | saran2]
MOOD_TREND: [tren]

Entri Jurnal:
%s`, contentBuilder.String())

	resp, err := model.GenerateContent(context.Background(), genai.Text(prompt))
	if err != nil {
		return "Gagal membuat ringkasan.", []string{}, []string{}, []string{}, "stable"
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return s.parseWeeklySummaryResponse(ctx, string(text))
		}
	}

	return "Gagal membuat ringkasan.", []string{}, []string{}, []string{}, "stable"
}

func (s *JournalService) parseWeeklySummaryResponse(ctx context.Context, response string) (string, []string, []string, []string, string) {
	summary := ""
	themes := []string{}
	insights := []string{}
	suggestions := []string{}
	moodTrend := "stable"

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SUMMARY:") {
			summary = strings.TrimPrefix(line, "SUMMARY:")
			summary = strings.TrimSpace(summary)
		} else if strings.HasPrefix(line, "THEMES:") {
			themesStr := strings.TrimPrefix(line, "THEMES:")
			themesStr = strings.TrimSpace(themesStr)
			for _, t := range strings.Split(themesStr, ",") {
				themes = append(themes, strings.TrimSpace(t))
			}
		} else if strings.HasPrefix(line, "INSIGHTS:") {
			insightsStr := strings.TrimPrefix(line, "INSIGHTS:")
			insightsStr = strings.TrimSpace(insightsStr)
			for _, i := range strings.Split(insightsStr, "|") {
				insights = append(insights, strings.TrimSpace(i))
			}
		} else if strings.HasPrefix(line, "SUGGESTIONS:") {
			suggestionsStr := strings.TrimPrefix(line, "SUGGESTIONS:")
			suggestionsStr = strings.TrimSpace(suggestionsStr)
			for _, s := range strings.Split(suggestionsStr, "|") {
				suggestions = append(suggestions, strings.TrimSpace(s))
			}
		} else if strings.HasPrefix(line, "MOOD_TREND:") {
			moodTrend = strings.TrimPrefix(line, "MOOD_TREND:")
			moodTrend = strings.TrimSpace(moodTrend)
		}
	}

	return summary, themes, insights, suggestions, moodTrend
}

func (s *JournalService) exportToTXT(ctx context.Context, journals []model.Journal) string {
	var builder strings.Builder
	builder.WriteString("=== RUANG TENANG - JOURNAL EXPORT ===\n")
	builder.WriteString(fmt.Sprintf("Exported: %s\n", time.Now().Format("2006-01-02 15:04")))
	builder.WriteString(fmt.Sprintf("Total Entries: %d\n", len(journals)))
	builder.WriteString("=====================================\n\n")

	for _, j := range journals {
		builder.WriteString(fmt.Sprintf("--- %s ---\n", j.CreatedAt.Format("Monday, 2 January 2006")))
		if j.Title != "" {
			builder.WriteString(fmt.Sprintf("Title: %s\n", j.Title))
		}
		if j.Mood != nil {
			builder.WriteString(fmt.Sprintf("Mood: %s %s\n", j.Mood.GetMoodEmoji(), j.Mood.Mood))
		}
		if len(j.Tags) > 0 {
			builder.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(j.Tags, ", ")))
		}
		builder.WriteString("\n")
		builder.WriteString(s.stripHTML(j.Content))
		builder.WriteString("\n\n")
	}

	return builder.String()
}

func (s *JournalService) exportToHTML(ctx context.Context, journals []model.Journal) string {
	var builder strings.Builder
	builder.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Journal Export - Ruang Tenang</title>
<style>
body { font-family: 'Segoe UI', sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
h1 { color: #E11D48; text-align: center; }
.entry { border: 1px solid #eee; padding: 20px; margin: 20px 0; border-radius: 8px; }
.entry-date { color: #666; font-size: 0.9em; }
.entry-title { font-size: 1.2em; font-weight: bold; margin: 10px 0; }
.entry-mood { display: inline-block; padding: 4px 12px; background: #f0f0f0; border-radius: 20px; margin: 5px 0; }
.entry-tags { color: #E11D48; font-size: 0.9em; }
.entry-content { margin-top: 15px; line-height: 1.6; white-space: pre-wrap; }
</style>
</head>
<body>
<h1>🧘 Ruang Tenang - Journal Export</h1>
<p style="text-align: center; color: #666;">`)
	builder.WriteString(fmt.Sprintf("Exported: %s | Total Entries: %d", time.Now().Format("2 January 2006"), len(journals)))
	builder.WriteString(`</p>`)

	for _, j := range journals {
		builder.WriteString(`<div class="entry">`)
		builder.WriteString(fmt.Sprintf(`<div class="entry-date">%s</div>`, j.CreatedAt.Format("Monday, 2 January 2006")))
		if j.Title != "" {
			builder.WriteString(fmt.Sprintf(`<div class="entry-title">%s</div>`, j.Title))
		}
		if j.Mood != nil {
			builder.WriteString(fmt.Sprintf(`<span class="entry-mood">%s %s</span>`, j.Mood.GetMoodEmoji(), j.Mood.Mood))
		}
		if len(j.Tags) > 0 {
			builder.WriteString(fmt.Sprintf(`<div class="entry-tags">Tags: %s</div>`, strings.Join(j.Tags, ", ")))
		}
		builder.WriteString(fmt.Sprintf(`<div class="entry-content">%s</div>`, j.Content))
		builder.WriteString(`</div>`)
	}

	builder.WriteString(`</body></html>`)
	return builder.String()
}

func (s *JournalService) exportToPDF(ctx context.Context, journals []model.Journal) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Ruang Tenang - Journal Export")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 10, fmt.Sprintf("Exported: %s | Total Entries: %d", time.Now().Format("2 Jan 2006"), len(journals)))
	pdf.Ln(20)

	for _, j := range journals {
		// Entry Header
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(240, 240, 240)
		header := fmt.Sprintf("%s", j.CreatedAt.Format("Monday, 2 January 2006"))
		pdf.CellFormat(0, 10, header, "0", 1, "L", true, 0, "")
		pdf.Ln(10)

		// Title
		if j.Title != "" {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(0, 8, j.Title)
			pdf.Ln(8)
		}

		// Meta (Mood & Tags)
		pdf.SetFont("Arial", "I", 10)
		meta := ""
		if j.Mood != nil {
			meta += fmt.Sprintf("Mood: %s | ", j.Mood.Mood)
		}
		if len(j.Tags) > 0 {
			meta += fmt.Sprintf("Tags: %s", strings.Join(j.Tags, ", "))
		}
		if meta != "" {
			pdf.Cell(0, 8, meta)
			pdf.Ln(8)
		}

		// Content
		pdf.SetFont("Arial", "", 11)
		content := s.stripHTML(j.Content)
		pdf.MultiCell(0, 6, content, "", "L", false)
		pdf.Ln(10)

		// Separator
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// stripHTML removes HTML tags and converts common block elements to newlines
func (s *JournalService) stripHTML(content string) string {
	// Replace block elements with newlines first
	content = strings.ReplaceAll(content, "</p>", "\n\n")
	content = strings.ReplaceAll(content, "</div>", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	// Strip all other tags
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// Decode HTML entities
	content = html.UnescapeString(content)

	return strings.TrimSpace(content)
}
