package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateJournal creates a new journal entry
func (s *JournalService) CreateJournal(ctx context.Context, userID uint, req dto.CreateJournalRequest) (*dto.JournalResponse, error) {
	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	shareWithAI := settings.DefaultShareWithAI
	if req.ShareWithAI != nil {
		shareWithAI = *req.ShareWithAI
	}

	if settings.IsBlocked {
		return nil, errors.New("you are blocked from creating journal entries")
	}

	wordCount := len(strings.Fields(req.Content))

	journal := &model.Journal{
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		MoodID:      req.MoodID,
		Tags:        pq.StringArray(req.Tags),
		IsPrivate:   true,
		ShareWithAI: shareWithAI,
		WordCount:   wordCount,
	}

	if err := s.journalRepo.Create(ctx, journal); err != nil {
		return nil, fmt.Errorf("failed to create journal: %w", err)
	}

	if wordCount > 100 {
		go func(journalID uint, content string) {
			bgCtx := context.Background()
			summary, summaryErr := s.generateSingleEntrySummary(bgCtx, content)
			if summaryErr == nil && summary != "" {
				_ = s.journalRepo.UpdateSummary(bgCtx, journalID, summary)
			}
		}(journal.ID, journal.Content)
	}

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
	journal, err := s.journalRepo.FindByIDAndUserID(ctx, journalID, userID)
	if err != nil {
		return nil, err
	}

	s.applyUpdateRequest(journal, req)

	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err == nil && settings.IsBlocked {
		return nil, errors.New("you are blocked from editing journal entries")
	}

	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	s.scheduleSummaryRegeneration(journal.ID, req.Content)

	return s.toJournalResponse(ctx, journal), nil
}

// UpdateJournalByUUID updates a journal entry by UUID
func (s *JournalService) UpdateJournalByUUID(ctx context.Context, userID uint, journalUUID string, req dto.UpdateJournalRequest) (*dto.JournalResponse, error) {
	id, err := uuid.Parse(journalUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	journal, err := s.journalRepo.FindByUUIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	s.applyUpdateRequest(journal, req)

	settings, err := s.settingsRepo.FindOrCreate(ctx, userID)
	if err == nil && settings.IsBlocked {
		return nil, errors.New("you are blocked from editing journal entries")
	}

	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	s.scheduleSummaryRegeneration(journal.ID, req.Content)

	return s.toJournalResponse(ctx, journal), nil
}

// DeleteJournal deletes a journal entry
func (s *JournalService) DeleteJournal(ctx context.Context, userID, journalID uint) error {
	if _, err := s.journalRepo.FindByIDAndUserID(ctx, journalID, userID); err != nil {
		return err
	}

	return s.journalRepo.Delete(ctx, journalID, userID)
}

// DeleteJournalByUUID deletes a journal entry by UUID
func (s *JournalService) DeleteJournalByUUID(ctx context.Context, userID uint, journalUUID string) error {
	id, err := uuid.Parse(journalUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

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
	for i, journal := range journals {
		responses[i] = s.toJournalListResponse(ctx, &journal)
	}

	return responses, total, nil
}

// SearchJournals searches journals by content
func (s *JournalService) SearchJournals(ctx context.Context, userID uint, query string, limit int) ([]dto.JournalListResponse, error) {
	var (
		journals []model.Journal
		err      error
	)
	if s.searchByContentFn != nil {
		journals, err = s.searchByContentFn(ctx, userID, query, limit)
	} else {
		journals, err = s.journalRepo.SearchByContent(ctx, userID, query, limit)
	}
	if err != nil {
		return nil, err
	}

	responses := make([]dto.JournalListResponse, len(journals))
	for i, journal := range journals {
		responses[i] = s.toJournalListResponse(ctx, &journal)
	}

	return responses, nil
}

func (s *JournalService) applyUpdateRequest(journal *model.Journal, req dto.UpdateJournalRequest) {
	if req.Title != nil {
		journal.Title = *req.Title
	}
	if req.Content != nil {
		journal.Content = *req.Content
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
}

func (s *JournalService) scheduleSummaryRegeneration(journalID uint, content *string) {
	if content == nil || len(strings.Fields(*content)) <= 100 {
		return
	}

	go func(entryID uint, entryContent string) {
		bgCtx := context.Background()
		summary, err := s.generateSingleEntrySummary(bgCtx, entryContent)
		if err == nil && summary != "" {
			_ = s.journalRepo.UpdateSummary(bgCtx, entryID, summary)
		}
	}(journalID, *content)
}
