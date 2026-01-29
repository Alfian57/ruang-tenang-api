package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// Story Category DTOs
// ==========================================

// StoryCategoryResponse represents a story category
type StoryCategoryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	StoryCount  int       `json:"story_count,omitempty"`
}

// ==========================================
// Inspiring Story DTOs
// ==========================================

// CreateStoryRequest for submitting a new inspiring story
type CreateStoryRequest struct {
	Title              string      `json:"title" binding:"required,min=5,max=200"`
	Content            string      `json:"content" binding:"required,min=200,max=50000"`
	CoverImage         string      `json:"cover_image" binding:"omitempty,max=500"`
	CategoryIDs        []uuid.UUID `json:"category_ids" binding:"required,min=1,max=3"`
	Tags               []string    `json:"tags" binding:"omitempty,max=5,dive,max=50"`
	IsAnonymous        bool        `json:"is_anonymous"`
	HasTriggerWarning  bool        `json:"has_trigger_warning"`
	TriggerWarningText string      `json:"trigger_warning_text" binding:"max=500"`
}

// UpdateStoryRequest for editing a story
type UpdateStoryRequest struct {
	Title              string      `json:"title" binding:"omitempty,min=5,max=200"`
	Content            string      `json:"content" binding:"omitempty,min=200,max=50000"`
	CoverImage         string      `json:"cover_image" binding:"omitempty,max=500"`
	CategoryIDs        []uuid.UUID `json:"category_ids" binding:"omitempty,min=1,max=3"`
	Tags               []string    `json:"tags" binding:"omitempty,max=5,dive,max=50"`
	IsAnonymous        bool        `json:"is_anonymous"`
	HasTriggerWarning  bool        `json:"has_trigger_warning"`
	TriggerWarningText string      `json:"trigger_warning_text" binding:"max=500"`
}

// StoryAuthorResponse represents story author info (respects anonymity)
type StoryAuthorResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar,omitempty"`
	TierName  string `json:"tier_name,omitempty"`
	TierColor string `json:"tier_color,omitempty"`
}

// StoryResponse represents a full inspiring story
type StoryResponse struct {
	ID                 uuid.UUID               `json:"id"`
	Title              string                  `json:"title"`
	Content            string                  `json:"content"`
	CoverImage         string                  `json:"cover_image"`
	IsAnonymous        bool                    `json:"is_anonymous"`
	HasTriggerWarning  bool                    `json:"has_trigger_warning"`
	TriggerWarningText string                  `json:"trigger_warning_text,omitempty"`
	Status             string                  `json:"status"`
	ViewCount          int                     `json:"view_count"`
	HeartCount         int                     `json:"heart_count"`
	CommentCount       int                     `json:"comment_count"`
	IsFeatured         bool                    `json:"is_featured"`
	Author             *StoryAuthorResponse    `json:"author,omitempty"`
	Categories         []StoryCategoryResponse `json:"categories"`
	Tags               []string                `json:"tags"`
	HasHearted         bool                    `json:"has_hearted"`
	CreatedAt          time.Time               `json:"created_at"`
	PublishedAt        *time.Time              `json:"published_at,omitempty"`
}

// StoryCardResponse represents a story card (list view)
type StoryCardResponse struct {
	ID                uuid.UUID               `json:"id"`
	Title             string                  `json:"title"`
	Excerpt           string                  `json:"excerpt"`
	CoverImage        string                  `json:"cover_image"`
	IsAnonymous       bool                    `json:"is_anonymous"`
	HasTriggerWarning bool                    `json:"has_trigger_warning"`
	HeartCount        int                     `json:"heart_count"`
	CommentCount      int                     `json:"comment_count"`
	IsFeatured        bool                    `json:"is_featured"`
	Author            *StoryAuthorResponse    `json:"author,omitempty"`
	Categories        []StoryCategoryResponse `json:"categories"`
	PublishedAt       *time.Time              `json:"published_at,omitempty"`
}

// StoriesListResponse for paginated story list
type StoriesListResponse struct {
	Stories    []StoryCardResponse `json:"stories"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

// StoryFilterRequest for filtering stories
type StoryFilterRequest struct {
	CategoryID string `form:"category_id"`
	Search     string `form:"search"`
	SortBy     string `form:"sort_by"` // recent, hearts, featured
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=10"`
	AuthorID   uint   `form:"author_id"`
	IsFeatured *bool  `form:"is_featured"`
}

// ==========================================
// Story Moderation DTOs
// ==========================================

// ModerateStoryRequest for moderator actions on stories
type ModerateStoryRequest struct {
	Status   string `json:"status" binding:"required,oneof=approved rejected revision_requested"`
	Feedback string `json:"feedback" binding:"max=2000"`
}

// StoryModerationQueueResponse for moderation queue
type StoryModerationQueueResponse struct {
	ID                 uuid.UUID               `json:"id"`
	Title              string                  `json:"title"`
	Content            string                  `json:"content"`
	CoverImage         string                  `json:"cover_image"`
	IsAnonymous        bool                    `json:"is_anonymous"`
	HasTriggerWarning  bool                    `json:"has_trigger_warning"`
	TriggerWarningText string                  `json:"trigger_warning_text,omitempty"`
	Status             string                  `json:"status"`
	Author             *StoryAuthorResponse    `json:"author"`
	Categories         []StoryCategoryResponse `json:"categories"`
	Tags               []string                `json:"tags"`
	SubmittedAt        time.Time               `json:"submitted_at"`
}

// ==========================================
// Story Comment DTOs
// ==========================================

// CreateCommentRequest for adding a supportive comment
type CreateStoryCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

// StoryCommentResponse represents a comment on a story
type StoryCommentResponse struct {
	ID         uuid.UUID            `json:"id"`
	Content    string               `json:"content"`
	HeartCount int                  `json:"heart_count"`
	Author     *StoryAuthorResponse `json:"author"`
	HasHearted bool                 `json:"has_hearted"`
	IsHidden   bool                 `json:"is_hidden,omitempty"`
	CreatedAt  time.Time            `json:"created_at"`
}

// StoryCommentsListResponse for paginated comments
type StoryCommentsListResponse struct {
	Comments   []StoryCommentResponse `json:"comments"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

// HideCommentRequest for moderator to hide a comment
type HideStoryCommentRequest struct {
	Reason string `json:"reason" binding:"required,max=255"`
}

// ==========================================
// Story Stats DTOs
// ==========================================

// StoryStatsResponse for author's story stats
type StoryStatsResponse struct {
	TotalStories       int   `json:"total_stories"`
	ApprovedStories    int   `json:"approved_stories"`
	PendingStories     int   `json:"pending_stories"`
	TotalHearts        int64 `json:"total_hearts"`
	TotalViews         int64 `json:"total_views"`
	TotalComments      int64 `json:"total_comments"`
	StoriesThisMonth   int   `json:"stories_this_month"`
	MaxStoriesPerMonth int   `json:"max_stories_per_month"`
	CanSubmitMore      bool  `json:"can_submit_more"`
}

// FeaturedStoryResponse for featured story widget
type FeaturedStoryResponse struct {
	Story      *StoryCardResponse `json:"story"`
	FeaturedAt *time.Time         `json:"featured_at"`
	FeaturedBy string             `json:"featured_by,omitempty"`
}

// MostAppreciatedStoriesResponse for most appreciated section
type MostAppreciatedStoriesResponse struct {
	Month   int                 `json:"month"`
	Year    int                 `json:"year"`
	Stories []StoryCardResponse `json:"stories"`
}

// ==========================================
// Story Nomination DTO (Level 8+ feature)
// ==========================================

// NominateStoryRequest for nominating a story for featured
type NominateStoryRequest struct {
	StoryID uuid.UUID `json:"story_id" binding:"required"`
	Reason  string    `json:"reason" binding:"required,min=10,max=500"`
}
