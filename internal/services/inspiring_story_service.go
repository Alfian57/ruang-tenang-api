package services

import (
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/google/uuid"
)

type InspiringStoryService struct {
	storyRepo       *repositories.InspiringStoryRepository
	userRepo        *repositories.UserRepository
	levelConfigRepo *repositories.LevelConfigRepository
	badgeService    *BadgeService
}

func NewInspiringStoryService(
	storyRepo *repositories.InspiringStoryRepository,
	userRepo *repositories.UserRepository,
	levelConfigRepo *repositories.LevelConfigRepository,
	badgeService *BadgeService,
) *InspiringStoryService {
	return &InspiringStoryService{
		storyRepo:       storyRepo,
		userRepo:        userRepo,
		levelConfigRepo: levelConfigRepo,
		badgeService:    badgeService,
	}
}

// ==========================================
// Story Categories
// ==========================================

// GetCategories returns all story categories with story counts
func (s *InspiringStoryService) GetCategories() ([]dto.StoryCategoryResponse, error) {
	categoriesWithCount, err := s.storyRepo.GetCategoriesWithCount()
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
func (s *InspiringStoryService) CreateStory(userID uint, req *dto.CreateStoryRequest) (*dto.StoryResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Get user level for monthly limit calculation
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(user.Exp)

	// Check monthly submission limit (3 per month for regular users, 5 for level 7+)
	now := time.Now()
	storiesThisMonth, _ := s.storyRepo.GetAuthorStoriesCount(userID, int(now.Month()), now.Year())
	maxStories := 3
	if currentLevel != nil && currentLevel.Level >= 7 {
		maxStories = 5 // Higher level users can submit more
	}

	if int(storiesThisMonth) >= maxStories {
		return nil, ErrMonthlyStoryLimitReached
	}

	// Create story
	story := &models.InspiringStory{
		AuthorID:           userID,
		Title:              req.Title,
		Content:            req.Content,
		CoverImage:         req.CoverImage,
		IsAnonymous:        req.IsAnonymous,
		HasTriggerWarning:  req.HasTriggerWarning,
		TriggerWarningText: req.TriggerWarningText,
		Status:             models.StoryStatusPending,
	}

	if err := s.storyRepo.Create(story); err != nil {
		return nil, err
	}

	// Set categories
	if len(req.CategoryIDs) > 0 {
		if err := s.storyRepo.SetStoryCategories(story.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Set tags
	if len(req.Tags) > 0 {
		if err := s.storyRepo.SetStoryTags(story.ID, req.Tags); err != nil {
			return nil, err
		}
	}

	// Fetch the complete story with relations
	return s.GetStory(story.ID, userID)
}

// UpdateStory updates an existing story
func (s *InspiringStoryService) UpdateStory(storyID uuid.UUID, userID uint, req *dto.UpdateStoryRequest) (*dto.StoryResponse, error) {
	story, err := s.storyRepo.GetByID(storyID)
	if err != nil {
		return nil, ErrStoryNotFound
	}

	// Check ownership
	if story.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	// Only allow editing pending or revision_requested stories
	if story.Status != models.StoryStatusPending && story.Status != models.StoryStatusRevisionRequested {
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
	story.Status = models.StoryStatusPending // Reset to pending for re-review

	if err := s.storyRepo.Update(story); err != nil {
		return nil, err
	}

	// Update categories if provided
	if len(req.CategoryIDs) > 0 {
		if err := s.storyRepo.SetStoryCategories(story.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Update tags if provided
	if req.Tags != nil {
		if err := s.storyRepo.SetStoryTags(story.ID, req.Tags); err != nil {
			return nil, err
		}
	}

	return s.GetStory(storyID, userID)
}

// DeleteStory deletes a story
func (s *InspiringStoryService) DeleteStory(storyID uuid.UUID, userID uint) error {
	story, err := s.storyRepo.GetByID(storyID)
	if err != nil {
		return ErrStoryNotFound
	}

	// Check ownership
	if story.AuthorID != userID {
		return ErrUnauthorized
	}

	return s.storyRepo.Delete(storyID)
}

// GetStory returns a single story with full details
func (s *InspiringStoryService) GetStory(storyID uuid.UUID, viewerID uint) (*dto.StoryResponse, error) {
	story, err := s.storyRepo.GetByIDWithRelations(storyID)
	if err != nil {
		return nil, ErrStoryNotFound
	}

	// Only show approved stories to non-owners
	if story.Status != "approved" && story.AuthorID != viewerID {
		return nil, ErrStoryNotFound
	}

	// Increment view count
	s.storyRepo.IncrementViewCount(storyID)

	// Check if viewer has hearted
	hasHearted := s.storyRepo.HasHearted(storyID, viewerID)

	// Get tags
	tags, _ := s.storyRepo.GetStoryTags(storyID)

	// Build author info
	var author *dto.StoryAuthorResponse
	if !story.IsAnonymous {
		author = s.buildAuthorResponse(story.AuthorID)
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
func (s *InspiringStoryService) GetStories(filter *dto.StoryFilterRequest, viewerID uint) (*dto.StoriesListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 10
	}

	var stories []models.InspiringStory
	var total int64
	var err error

	if filter.CategoryID != "" {
		catID, _ := uuid.Parse(filter.CategoryID)
		stories, total, err = s.storyRepo.GetStoriesByCategory(catID, filter.Page, filter.Limit)
	} else if filter.Search != "" {
		stories, total, err = s.storyRepo.SearchStories(filter.Search, filter.Page, filter.Limit)
	} else if filter.AuthorID > 0 {
		stories, total, err = s.storyRepo.GetStoriesByAuthor(filter.AuthorID, "approved", filter.Page, filter.Limit)
	} else {
		stories, total, err = s.storyRepo.GetApprovedStories(filter.Page, filter.Limit, filter.SortBy)
	}

	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(story)
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
func (s *InspiringStoryService) GetUserStories(userID uint, status string, page, limit int) (*dto.StoriesListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	stories, total, err := s.storyRepo.GetStoriesByAuthor(userID, status, page, limit)
	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(story)
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
func (s *InspiringStoryService) GetFeaturedStories(limit int) ([]dto.StoryCardResponse, error) {
	if limit <= 0 {
		limit = 5
	}

	stories, err := s.storyRepo.GetFeaturedStories(limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		result[i] = s.toStoryCard(story)
	}

	return result, nil
}

// ==========================================
// Story Moderation
// ==========================================

// GetPendingStories returns stories pending moderation
func (s *InspiringStoryService) GetPendingStories(page, limit int) (*dto.StoriesListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	stories, total, err := s.storyRepo.GetPendingStories(page, limit)
	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(story)
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
func (s *InspiringStoryService) ModerateStory(storyID uuid.UUID, moderatorID uint, req *dto.ModerateStoryRequest) error {
	story, err := s.storyRepo.GetByID(storyID)
	if err != nil {
		return ErrStoryNotFound
	}

	if err := s.storyRepo.UpdateStatus(storyID, req.Status, moderatorID, req.Feedback); err != nil {
		return err
	}

	// If approved, check for first story badge
	if req.Status == "approved" {
		s.badgeService.AwardBadge(story.AuthorID, "first_story")
	}

	return nil
}

// SetFeatured sets or removes featured status
func (s *InspiringStoryService) SetFeatured(storyID uuid.UUID, featured bool, moderatorID uint) error {
	_, err := s.storyRepo.GetByID(storyID)
	if err != nil {
		return ErrStoryNotFound
	}

	return s.storyRepo.SetFeatured(storyID, featured, moderatorID)
}

// ==========================================
// Story Hearts
// ==========================================

// ToggleHeart toggles heart on a story
func (s *InspiringStoryService) ToggleHeart(storyID uuid.UUID, userID uint) (bool, int, error) {
	story, err := s.storyRepo.GetByID(storyID)
	if err != nil {
		return false, 0, ErrStoryNotFound
	}

	if story.Status != "approved" {
		return false, 0, ErrStoryNotApproved
	}

	hasHearted := s.storyRepo.HasHearted(storyID, userID)

	if hasHearted {
		if err := s.storyRepo.RemoveHeart(storyID, userID); err != nil {
			return false, 0, err
		}
	} else {
		if err := s.storyRepo.AddHeart(storyID, userID); err != nil {
			return false, 0, err
		}
		// Check for supportive heart badges
		s.badgeService.AwardBadge(userID, "first_heart")
	}

	// Get updated count
	count, _ := s.storyRepo.GetStoryHeartCount(storyID)

	return !hasHearted, count, nil
}

// ==========================================
// Story Comments
// ==========================================

// CreateComment creates a comment on a story
func (s *InspiringStoryService) CreateComment(storyID uuid.UUID, userID uint, req *dto.CreateStoryCommentRequest) (*dto.StoryCommentResponse, error) {
	story, err := s.storyRepo.GetByID(storyID)
	if err != nil {
		return nil, ErrStoryNotFound
	}

	if story.Status != "approved" {
		return nil, ErrStoryNotApproved
	}

	comment := &models.StoryComment{
		StoryID: storyID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := s.storyRepo.CreateComment(comment); err != nil {
		return nil, err
	}

	// Award first comment badge
	s.badgeService.AwardBadge(userID, "first_comment")

	// Get comment with user info
	comment, _ = s.storyRepo.GetCommentByID(comment.ID)

	return s.toCommentResponse(comment, userID), nil
}

// GetComments returns comments for a story
func (s *InspiringStoryService) GetComments(storyID uuid.UUID, viewerID uint, page, limit int) (*dto.StoryCommentsListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	comments, total, err := s.storyRepo.GetStoryComments(storyID, page, limit, false)
	if err != nil {
		return nil, err
	}

	commentResponses := make([]dto.StoryCommentResponse, len(comments))
	for i, c := range comments {
		commentResponses[i] = *s.toCommentResponse(&c, viewerID)
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
func (s *InspiringStoryService) DeleteComment(commentID uuid.UUID, userID uint) error {
	comment, err := s.storyRepo.GetCommentByID(commentID)
	if err != nil {
		return ErrCommentNotFound
	}

	// Check ownership
	if comment.UserID != userID {
		return ErrUnauthorized
	}

	return s.storyRepo.DeleteComment(commentID, comment.StoryID)
}

// HideComment hides a comment (moderator action)
func (s *InspiringStoryService) HideComment(commentID uuid.UUID, req *dto.HideStoryCommentRequest) error {
	_, err := s.storyRepo.GetCommentByID(commentID)
	if err != nil {
		return ErrCommentNotFound
	}

	return s.storyRepo.HideComment(commentID, true, req.Reason)
}

// ToggleCommentHeart toggles heart on a comment
func (s *InspiringStoryService) ToggleCommentHeart(commentID uuid.UUID, userID uint) (bool, int, error) {
	comment, err := s.storyRepo.GetCommentByID(commentID)
	if err != nil {
		return false, 0, ErrCommentNotFound
	}

	hasHearted := s.storyRepo.HasHeartedComment(commentID, userID)

	if hasHearted {
		if err := s.storyRepo.RemoveCommentHeart(commentID, userID); err != nil {
			return false, 0, err
		}
	} else {
		if err := s.storyRepo.AddCommentHeart(commentID, userID); err != nil {
			return false, 0, err
		}
	}

	// Get updated comment for count
	comment, _ = s.storyRepo.GetCommentByID(commentID)

	return !hasHearted, comment.HeartCount, nil
}

// ==========================================
// Story Stats
// ==========================================

// GetAuthorStats returns stats for an author
func (s *InspiringStoryService) GetAuthorStats(authorID uint) (*dto.StoryStatsResponse, error) {
	stats, err := s.storyRepo.GetAuthorStats(authorID)
	if err != nil {
		return nil, err
	}

	// Get stories this month
	now := time.Now()
	storiesThisMonth, _ := s.storyRepo.GetAuthorStoriesCount(authorID, int(now.Month()), now.Year())

	// Get user level for max stories
	user, _ := s.userRepo.FindByID(authorID)
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(user.Exp)
	maxStories := 3
	if currentLevel != nil && currentLevel.Level >= 7 {
		maxStories = 5
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
func (s *InspiringStoryService) GetMostAppreciatedStories(month, year, limit int) (*dto.MostAppreciatedStoriesResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	stories, err := s.storyRepo.GetMostAppreciatedStories(month, year, limit)
	if err != nil {
		return nil, err
	}

	storyCards := make([]dto.StoryCardResponse, len(stories))
	for i, story := range stories {
		storyCards[i] = s.toStoryCard(story)
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

func (s *InspiringStoryService) createExcerpt(content string, maxLength int) string {
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

func (s *InspiringStoryService) buildAuthorResponse(userID uint) *dto.StoryAuthorResponse {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil
	}

	// Get tier info
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(user.Exp)

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

func (s *InspiringStoryService) toStoryCard(story models.InspiringStory) dto.StoryCardResponse {
	var author *dto.StoryAuthorResponse
	if !story.IsAnonymous && story.Author != nil {
		author = s.buildAuthorResponse(story.AuthorID)
	} else if !story.IsAnonymous {
		author = s.buildAuthorResponse(story.AuthorID)
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
		Excerpt:           s.createExcerpt(story.Content, 200),
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

func (s *InspiringStoryService) toCommentResponse(comment *models.StoryComment, viewerID uint) *dto.StoryCommentResponse {
	var author *dto.StoryAuthorResponse
	if comment.User != nil {
		author = s.buildAuthorResponse(comment.UserID)
	}

	hasHearted := s.storyRepo.HasHeartedComment(comment.ID, viewerID)

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
