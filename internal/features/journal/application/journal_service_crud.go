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

	// Journals are private by default; only public when explicitly requested.
	isPrivate := true
	if req.IsPrivate != nil {
		isPrivate = *req.IsPrivate
	}

	if settings.IsBlocked {
		return nil, errors.New("you are blocked from creating journal entries")
	}

	// Gate public journals through AI moderation: only clearly-approved content
	// may go public. Otherwise it is forced private and the user is informed.
	moderationNotice := ""
	if !isPrivate {
		if approved, notice := s.moderatePublicContent(ctx, req.Title, req.Content); !approved {
			isPrivate = true
			moderationNotice = notice
		}
	}

	wordCount := len(strings.Fields(req.Content))

	journal := &model.Journal{
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		MoodID:      req.MoodID,
		Tags:        pq.StringArray(req.Tags),
		IsPrivate:   isPrivate,
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

	response := s.toJournalResponse(ctx, journal)
	response.ModerationNotice = moderationNotice
	return response, nil
}

// moderatePublicContent runs AI moderation on content a user wants to publish
// publicly. Returns (approved, notice). When the moderator is unavailable it
// fails safe by NOT approving, so sensitive content never auto-publishes.
func (s *JournalService) moderatePublicContent(ctx context.Context, title, content string) (bool, string) {
	const declineNotice = "Jurnal disimpan sebagai privat karena moderasi otomatis belum dapat menyetujuinya untuk dibagikan ke komunitas. Kamu bisa mencoba menyuntingnya atau menjadikannya publik lagi nanti."

	if s.moderator == nil {
		return false, declineNotice
	}

	result, err := s.moderator.ModerateArticle(ctx, title, content)
	if err != nil || result == nil {
		return false, declineNotice
	}

	if result.Status == model.ArticleModerationApproved {
		return true, ""
	}

	return false, declineNotice
}

// gatePublicOnUpdate runs AI moderation when an update explicitly switches a
// journal to public. If not approved, the journal is reverted to private and a
// notice is returned. Returns "" when no gating was needed.
func (s *JournalService) gatePublicOnUpdate(ctx context.Context, journal *model.Journal, req dto.UpdateJournalRequest) string {
	// Only gate when the request explicitly requests public (is_private=false).
	if req.IsPrivate == nil || *req.IsPrivate {
		return ""
	}

	if approved, notice := s.moderatePublicContent(ctx, journal.Title, journal.Content); !approved {
		journal.IsPrivate = true
		return notice
	}
	return ""
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

	moderationNotice := s.gatePublicOnUpdate(ctx, journal, req)

	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	s.scheduleSummaryRegeneration(journal.ID, req.Content)

	response := s.toJournalResponse(ctx, journal)
	response.ModerationNotice = moderationNotice
	return response, nil
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

	moderationNotice := s.gatePublicOnUpdate(ctx, journal, req)

	if err := s.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	s.scheduleSummaryRegeneration(journal.ID, req.Content)

	response := s.toJournalResponse(ctx, journal)
	response.ModerationNotice = moderationNotice
	return response, nil
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

// ListPublicJournals returns public journals from all users for the community feed.
func (s *JournalService) ListPublicJournals(ctx context.Context, page, limit int, tags []string, search string) ([]dto.PublicJournalListResponse, int64, error) {
	journals, total, err := s.journalRepo.FindPublic(ctx, page, limit, tags, search)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.PublicJournalListResponse, len(journals))
	for i := range journals {
		responses[i] = s.toPublicJournalListResponse(ctx, &journals[i])
	}

	return responses, total, nil
}

// GetPublicJournal returns a single public journal by UUID for the community detail view.
func (s *JournalService) GetPublicJournal(ctx context.Context, journalUUID string) (*dto.PublicJournalResponse, error) {
	id, err := uuid.Parse(journalUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	journal, err := s.journalRepo.FindPublicByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toPublicJournalResponse(ctx, journal), nil
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
	if req.IsPrivate != nil {
		journal.IsPrivate = *req.IsPrivate
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
