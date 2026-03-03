package service

import (
	"context"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"github.com/google/uuid"
)

type InspiringStoryService struct {
	storyRepo           *repository.InspiringStoryRepository
	userRepo            *repository.UserRepository
	levelConfigRepo     *repository.LevelConfigRepository
	badgeService        *BadgeService
	gamificationService *GamificationService
	notificationService *NotificationService
}

func NewInspiringStoryService(
	storyRepo *repository.InspiringStoryRepository,
	userRepo *repository.UserRepository,
	levelConfigRepo *repository.LevelConfigRepository,
	badgeService *BadgeService,
	gamificationService *GamificationService,
	notificationService *NotificationService,
) *InspiringStoryService {
	return &InspiringStoryService{
		storyRepo:           storyRepo,
		userRepo:            userRepo,
		levelConfigRepo:     levelConfigRepo,
		badgeService:        badgeService,
		gamificationService: gamificationService,
		notificationService: notificationService,
	}
}

// ==========================================
// Story Categories
// ==========================================

// GetCategories returns all story categories with story counts
func (s *InspiringStoryService) GetCategories(ctx context.Context) ([]dto.StoryCategoryResponse, error) {
	categoriesWithCount, err := s.storyRepo.GetCategoriesWithCount(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.StoryCategoryResponse, len(categoriesWithCount))
	for i, c := range categoriesWithCount {
		result[i] = dto.StoryCategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			Icon:        c.Icon,
			StoryCount:  c.StoryCount,
		}
	}

	return result, nil
}

// ==========================================
// Story CRUD
// ==========================================

// CreateStory creates a new inspiring story
func (s *InspiringStoryService) CreateStory(ctx context.Context, userID uint, req *dto.CreateStoryRequest) (*dto.StoryResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user level for monthly limit calculation
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)

	// Check monthly submission limit (3 per month for regular users, 5 for level 7+)
	now := time.Now()
	storiesThisMonth, _ := s.storyRepo.GetAuthorStoriesCount(ctx, userID, int(now.Month()), now.Year())
	maxStories := 3
	if currentLevel != nil && currentLevel.Level >= 7 {
		maxStories = 5 // Higher level users can submit more
	}

	if int(storiesThisMonth) >= maxStories {
		return nil, ErrMonthlyStoryLimitReached
	}

	// Create story
	story := &model.InspiringStory{
		AuthorID:           userID,
		Title:              req.Title,
		Content:            req.Content,
		CoverImage:         req.CoverImage,
		IsAnonymous:        req.IsAnonymous,
		HasTriggerWarning:  req.HasTriggerWarning,
		TriggerWarningText: req.TriggerWarningText,
		Status:             model.StoryStatusPending,
	}

	if err := s.storyRepo.Create(ctx, story); err != nil {
		return nil, err
	}

	// Set categories
	if len(req.CategoryIDs) > 0 {
		if err := s.storyRepo.SetStoryCategories(ctx, story.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Set tags
	if len(req.Tags) > 0 {
		if err := s.storyRepo.SetStoryTags(ctx, story.ID, req.Tags); err != nil {
			return nil, err
		}
	}

	// Fetch the complete story with relations
	createdStory, err := s.storyRepo.GetByIDWithRelations(ctx, story.ID)
	if err != nil {
		return nil, err
	}

	return s.buildStoryResponseDTO(ctx, createdStory, userID)
}

// UpdateStory updates an existing story
func (s *InspiringStoryService) UpdateStory(ctx context.Context, storyID uuid.UUID, userID uint, req *dto.UpdateStoryRequest) (*dto.StoryResponse, error) {
	story, err := s.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return nil, ErrStoryNotFound
	}

	// Check ownership
	if story.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	// Only allow editing pending or revision_requested stories
	if story.Status != model.StoryStatusPending && story.Status != model.StoryStatusRevisionRequested {
		return nil, ErrCannotEditPublishedStory
	}

	// Update fields
	if req.Title != "" {
		story.Title = req.Title
	}
	if req.Content != "" {
		story.Content = req.Content
	}
	if req.CoverImage != "" {
		story.CoverImage = req.CoverImage
	}
	story.IsAnonymous = req.IsAnonymous
	story.HasTriggerWarning = req.HasTriggerWarning
	story.TriggerWarningText = req.TriggerWarningText
	story.Status = model.StoryStatusPending // Reset to pending for re-review

	if err := s.storyRepo.Update(ctx, story); err != nil {
		return nil, err
	}

	// Update categories if provided
	if len(req.CategoryIDs) > 0 {
		if err := s.storyRepo.SetStoryCategories(ctx, story.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Update tags if provided
	if req.Tags != nil {
		if err := s.storyRepo.SetStoryTags(ctx, story.ID, req.Tags); err != nil {
			return nil, err
		}
	}

	// Fetch the complete story with relations
	updatedStory, err := s.storyRepo.GetByIDWithRelations(ctx, story.ID)
	if err != nil {
		return nil, err
	}

	return s.buildStoryResponseDTO(ctx, updatedStory, userID)
}

// DeleteStory deletes a story
func (s *InspiringStoryService) DeleteStory(ctx context.Context, storyID uuid.UUID, userID uint) error {
	story, err := s.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return ErrStoryNotFound
	}

	// Check ownership
	if story.AuthorID != userID {
		return ErrUnauthorized
	}

	return s.storyRepo.Delete(ctx, storyID)
}

// GetStory returns a single story with full details
func (s *InspiringStoryService) GetStory(ctx context.Context, storyID uuid.UUID, viewerID uint) (*dto.StoryResponse, error) {
	story, err := s.storyRepo.GetByIDWithRelations(ctx, storyID)
	if err != nil {
		return nil, ErrStoryNotFound
	}

	// Only show approved stories to non-owners
	if story.Status != "approved" && story.AuthorID != viewerID {
		return nil, ErrStoryNotFound
	}

	// Increment view count
	s.storyRepo.IncrementViewCount(ctx, storyID)

	return s.buildStoryResponseDTO(ctx, story, viewerID)
}

func (s *InspiringStoryService) buildStoryResponseDTO(ctx context.Context, story *model.InspiringStory, viewerID uint) (*dto.StoryResponse, error) {
	// Check if viewer has hearted
	hasHearted := s.storyRepo.HasHearted(ctx, story.ID, viewerID)

	// Get tags
	tags, _ := s.storyRepo.GetStoryTags(ctx, story.ID)

	// Build author info
	var author *dto.StoryAuthorResponse
	if !story.IsAnonymous {
		author = s.buildAuthorResponse(ctx, story.AuthorID)
	} else {
		author = &dto.StoryAuthorResponse{
			Name: "Anonim",
		}
	}

	// Build categories
	categories := make([]dto.StoryCategoryResponse, len(story.Categories))
	for i, cat := range story.Categories {
		categories[i] = dto.StoryCategoryResponse{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
			Icon: cat.Icon,
		}
	}

	return &dto.StoryResponse{
		ID:                 story.ID,
		Title:              story.Title,
		Content:            story.Content,
		CoverImage:         story.CoverImage,
		IsAnonymous:        story.IsAnonymous,
		HasTriggerWarning:  story.HasTriggerWarning,
		TriggerWarningText: story.TriggerWarningText,
		Status:             string(story.Status),
		ViewCount:          story.ViewCount,
		HeartCount:         story.HeartCount,
		CommentCount:       story.CommentCount,
		IsFeatured:         story.IsFeatured,
		Author:             author,
		Categories:         categories,
		Tags:               tags,
		HasHearted:         hasHearted,
		CreatedAt:          story.CreatedAt,
		PublishedAt:        story.PublishedAt,
	}, nil
}

// GetStories returns paginated approved stories
func (s *InspiringStoryService) GetStories(ctx context.Context, filter *dto.StoryFilterRequest, viewerID uint) (*dto.StoriesListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 10
	}

	var stories []model.InspiringStory
	var total int64
	var err error

	if filter.CategoryID != "" {
		catID, _ := uuid.Parse(filter.CategoryID)
		stories, total, err = s.storyRepo.GetStoriesByCategory(ctx, catID, filter.Page, filter.Limit)
	} else if filter.Search != "" {
		stories, total, err = s.storyRepo.SearchStories(ctx, filter.Search, filter.Page, filter.Limit)
	} else if filter.AuthorID > 0 {
		stories, total, err = s.storyRepo.GetStoriesByAuthor(ctx, filter.AuthorID, "approved", filter.Page, filter.Limit)
	} else {
		stories, total, err = s.storyRepo.GetApprovedStories(ctx, filter.Page, filter.Limit, filter.SortBy)
	}

	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(ctx, story)
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	return &dto.StoriesListResponse{
		Stories:    storyCards,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetUserStories returns stories by a specific user
func (s *InspiringStoryService) GetUserStories(ctx context.Context, userID uint, status string, page, limit int) (*dto.StoriesListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	stories, total, err := s.storyRepo.GetStoriesByAuthor(ctx, userID, status, page, limit)
	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(ctx, story)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.StoriesListResponse{
		Stories:    storyCards,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetFeaturedStories returns featured stories
func (s *InspiringStoryService) GetFeaturedStories(ctx context.Context, limit int) ([]dto.StoryCardResponse, error) {
	if limit <= 0 {
		limit = 5
	}

	stories, err := s.storyRepo.GetFeaturedStories(ctx, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		result[i] = s.toStoryCard(ctx, story)
	}

	return result, nil
}

// ==========================================
// Story Moderation
// ==========================================

// GetPendingStories returns stories pending moderation
func (s *InspiringStoryService) GetPendingStories(ctx context.Context, page, limit int) (*dto.StoriesListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	stories, total, err := s.storyRepo.GetPendingStories(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(ctx, story)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.StoriesListResponse{
		Stories:    storyCards,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ModerateStory moderates a story (approve, reject, request revision)
func (s *InspiringStoryService) ModerateStory(ctx context.Context, storyID uuid.UUID, moderatorID uint, req *dto.ModerateStoryRequest) error {
	story, err := s.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return ErrStoryNotFound
	}

	if err := s.storyRepo.UpdateStatus(ctx, storyID, req.Status, moderatorID, req.Feedback); err != nil {
		return err
	}

	// If approved, award badge and XP to the story author
	if req.Status == "approved" {
		if s.badgeService != nil {
			s.badgeService.AwardBadge(ctx, story.AuthorID, "first_story")
		}
		// Award XP for approved story (with daily cap)
		if s.gamificationService != nil {
			s.gamificationService.AwardExp(ctx, story.AuthorID, gamification.ActivityStoryApproved, gamification.ExpStoryApproved)
		}
		// Notify the author
		if s.notificationService != nil {
			s.notificationService.CreateStoryApprovedNotification(ctx, story.AuthorID, story.Title, story.ID.String())
		}
	}

	// If rejected or needs revision, notify the author
	if req.Status == "rejected" || req.Status == "revision_requested" {
		if s.notificationService != nil {
			s.notificationService.CreateStoryRejectedNotification(ctx, story.AuthorID, story.Title, story.ID.String(), req.Feedback)
		}
	}

	return nil
}

// SetFeatured sets or removes featured status
func (s *InspiringStoryService) SetFeatured(ctx context.Context, storyID uuid.UUID, featured bool, moderatorID uint) error {
	_, err := s.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return ErrStoryNotFound
	}

	return s.storyRepo.SetFeatured(ctx, storyID, featured, moderatorID)
}

// ==========================================
// Story Hearts
// ==========================================

// ToggleHeart toggles heart on a story
func (s *InspiringStoryService) ToggleHeart(ctx context.Context, storyID uuid.UUID, userID uint) (bool, int, error) {
	story, err := s.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return false, 0, ErrStoryNotFound
	}

	if story.Status != "approved" {
		return false, 0, ErrStoryNotApproved
	}

	hasHearted := s.storyRepo.HasHearted(ctx, storyID, userID)

	if hasHearted {
		if err := s.storyRepo.RemoveHeart(ctx, storyID, userID); err != nil {
			return false, 0, err
		}
	} else {
		if err := s.storyRepo.AddHeart(ctx, storyID, userID); err != nil {
			return false, 0, err
		}
		// Check for supportive heart badges
		if s.badgeService != nil {
			s.badgeService.AwardBadge(ctx, userID, "first_heart")
		}

		// Award XP to story author (not to self-heart)
		if story.AuthorID != userID {
			if s.gamificationService != nil {
				s.gamificationService.AwardExp(ctx, story.AuthorID, gamification.ActivityHeartReceived, gamification.ExpHeartReceived)
			}

			// Notify story author about the heart
			var heartGiver *model.User
			if s.userRepo != nil {
				heartGiver, _ = s.userRepo.FindByID(ctx, userID)
			}
			heartGiverName := "Seseorang"
			if heartGiver != nil {
				heartGiverName = heartGiver.Name
			}
			if s.notificationService != nil {
				s.notificationService.CreateHeartNotification(ctx, story.AuthorID, heartGiverName, story.Title, storyID.String())
			}
		}
	}

	// Get updated count
	count, _ := s.storyRepo.GetStoryHeartCount(ctx, storyID)

	return !hasHearted, count, nil
}

// ==========================================
// Story Comments
// ==========================================

// CreateComment creates a comment on a story
func (s *InspiringStoryService) CreateComment(ctx context.Context, storyID uuid.UUID, userID uint, req *dto.CreateStoryCommentRequest) (*dto.StoryCommentResponse, error) {
	story, err := s.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return nil, ErrStoryNotFound
	}

	if story.Status != "approved" {
		return nil, ErrStoryNotApproved
	}

	comment := &model.StoryComment{
		StoryID: storyID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := s.storyRepo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	// Award first comment badge
	if s.badgeService != nil {
		s.badgeService.AwardBadge(ctx, userID, "first_comment")
	}

	// Get comment with user info
	comment, _ = s.storyRepo.GetCommentByID(ctx, comment.ID)

	return s.toCommentResponse(ctx, comment, userID), nil
}

// GetComments returns comments for a story
func (s *InspiringStoryService) GetComments(ctx context.Context, storyID uuid.UUID, viewerID uint, page, limit int) (*dto.StoryCommentsListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	comments, total, err := s.storyRepo.GetStoryComments(ctx, storyID, page, limit, false)
	if err != nil {
		return nil, err
	}

	commentResponses := make([]dto.StoryCommentResponse, len(comments))
	for i, c := range comments {
		commentResponses[i] = *s.toCommentResponse(ctx, &c, viewerID)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.StoryCommentsListResponse{
		Comments:   commentResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// DeleteComment deletes a comment
func (s *InspiringStoryService) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uint) error {
	comment, err := s.storyRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return ErrCommentNotFound
	}

	// Check ownership
	if comment.UserID != userID {
		return ErrUnauthorized
	}

	return s.storyRepo.DeleteComment(ctx, commentID, comment.StoryID)
}

// HideComment hides a comment (moderator action)
func (s *InspiringStoryService) HideComment(ctx context.Context, commentID uuid.UUID, req *dto.HideStoryCommentRequest) error {
	_, err := s.storyRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return ErrCommentNotFound
	}

	return s.storyRepo.HideComment(ctx, commentID, true, req.Reason)
}

// ToggleCommentHeart toggles heart on a comment
func (s *InspiringStoryService) ToggleCommentHeart(ctx context.Context, commentID uuid.UUID, userID uint) (bool, int, error) {
	comment, err := s.storyRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return false, 0, ErrCommentNotFound
	}

	hasHearted := s.storyRepo.HasHeartedComment(ctx, commentID, userID)

	if hasHearted {
		if err := s.storyRepo.RemoveCommentHeart(ctx, commentID, userID); err != nil {
			return false, 0, err
		}
	} else {
		if err := s.storyRepo.AddCommentHeart(ctx, commentID, userID); err != nil {
			return false, 0, err
		}
	}

	// Get updated comment for count
	comment, _ = s.storyRepo.GetCommentByID(ctx, commentID)

	return !hasHearted, comment.HeartCount, nil
}

// ==========================================
// Story Stats
// ==========================================

// GetAuthorStats returns stats for an author
func (s *InspiringStoryService) GetAuthorStats(ctx context.Context, authorID uint) (*dto.StoryStatsResponse, error) {
	stats, err := s.storyRepo.GetAuthorStats(ctx, authorID)
	if err != nil {
		return nil, err
	}

	// Get stories this month
	now := time.Now()
	storiesThisMonth, _ := s.storyRepo.GetAuthorStoriesCount(ctx, authorID, int(now.Month()), now.Year())

	// Get user level for max stories
	maxStories := 3
	if s.userRepo != nil && s.levelConfigRepo != nil {
		user, userErr := s.userRepo.FindByID(ctx, authorID)
		if userErr == nil && user != nil {
			currentLevel, levelErr := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
			if levelErr == nil && currentLevel != nil && currentLevel.Level >= 7 {
				maxStories = 5
			}
		}
	}

	return &dto.StoryStatsResponse{
		TotalStories:       int(stats.TotalStories),
		ApprovedStories:    int(stats.ApprovedStories),
		PendingStories:     int(stats.PendingStories),
		TotalHearts:        stats.TotalHearts,
		TotalViews:         stats.TotalViews,
		TotalComments:      stats.TotalComments,
		StoriesThisMonth:   int(storiesThisMonth),
		MaxStoriesPerMonth: maxStories,
		CanSubmitMore:      int(storiesThisMonth) < maxStories,
	}, nil
}

// GetMostAppreciatedStories returns most hearted stories for a period
func (s *InspiringStoryService) GetMostAppreciatedStories(ctx context.Context, month, year, limit int) (*dto.MostAppreciatedStoriesResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	stories, err := s.storyRepo.GetMostAppreciatedStories(ctx, month, year, limit)
	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(ctx, story)
	}

	return &dto.MostAppreciatedStoriesResponse{
		Month:   month,
		Year:    year,
		Stories: storyCards,
	}, nil
}

// ==========================================
// Helper Methods
// ==========================================

func (s *InspiringStoryService) createExcerpt(ctx context.Context, content string, maxLength int) string {
	// Strip HTML tags if any
	content = strings.TrimSpace(content)

	if len(content) <= maxLength {
		return content
	}

	// Find last space before maxLength
	truncated := content[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

func (s *InspiringStoryService) buildAuthorResponse(ctx context.Context, userID uint) *dto.StoryAuthorResponse {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil
	}

	// Get tier info
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)

	author := &dto.StoryAuthorResponse{
		ID:     user.ID,
		Name:   user.Name,
		Avatar: user.Avatar,
	}

	if currentLevel != nil {
		author.TierName = currentLevel.TierName
		author.TierColor = currentLevel.TierColor
	}

	return author
}

func (s *InspiringStoryService) toStoryCard(ctx context.Context, story model.InspiringStory) dto.StoryCardResponse {
	var author *dto.StoryAuthorResponse
	if !story.IsAnonymous && story.Author != nil {
		author = s.buildAuthorResponse(ctx, story.AuthorID)
	} else if !story.IsAnonymous {
		author = s.buildAuthorResponse(ctx, story.AuthorID)
	} else {
		author = &dto.StoryAuthorResponse{
			Name: "Anonim",
		}
	}

	categories := make([]dto.StoryCategoryResponse, len(story.Categories))
	for i, cat := range story.Categories {
		categories[i] = dto.StoryCategoryResponse{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
			Icon: cat.Icon,
		}
	}

	return dto.StoryCardResponse{
		ID:                story.ID,
		Title:             story.Title,
		Excerpt:           s.createExcerpt(ctx, story.Content, 200),
		CoverImage:        story.CoverImage,
		IsAnonymous:       story.IsAnonymous,
		HasTriggerWarning: story.HasTriggerWarning,
		HeartCount:        story.HeartCount,
		CommentCount:      story.CommentCount,
		IsFeatured:        story.IsFeatured,
		Author:            author,
		Categories:        categories,
		PublishedAt:       story.PublishedAt,
	}
}

func (s *InspiringStoryService) toCommentResponse(ctx context.Context, comment *model.StoryComment, viewerID uint) *dto.StoryCommentResponse {
	var author *dto.StoryAuthorResponse
	if comment.User != nil {
		author = s.buildAuthorResponse(ctx, comment.UserID)
	}

	hasHearted := s.storyRepo.HasHeartedComment(ctx, comment.ID, viewerID)

	return &dto.StoryCommentResponse{
		ID:         comment.ID,
		Content:    comment.Content,
		HeartCount: comment.HeartCount,
		Author:     author,
		HasHearted: hasHearted,
		IsHidden:   comment.IsHidden,
		CreatedAt:  comment.CreatedAt,
	}
}

// Custom errors
var (
	ErrStoryNotFound            = &ServiceError{Code: "STORY_NOT_FOUND", Message: "Cerita tidak ditemukan"}
	ErrStoryNotApproved         = &ServiceError{Code: "STORY_NOT_APPROVED", Message: "Cerita belum disetujui"}
	ErrCommentNotFound          = &ServiceError{Code: "COMMENT_NOT_FOUND", Message: "Komentar tidak ditemukan"}
	ErrUnauthorized             = &ServiceError{Code: "UNAUTHORIZED", Message: "Tidak memiliki akses"}
	ErrCannotEditPublishedStory = &ServiceError{Code: "CANNOT_EDIT_PUBLISHED", Message: "Tidak dapat mengedit cerita yang sudah dipublikasukan"}
	ErrMonthlyStoryLimitReached = &ServiceError{Code: "MONTHLY_LIMIT_REACHED", Message: "Kamu sudah mencapai batas maksimal cerita bulan ini"}
)
