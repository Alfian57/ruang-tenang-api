package dto

import "time"

// ================================
// Chat Folder DTOs
// ================================

// CreateChatFolderRequest for creating a new folder
type CreateChatFolderRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=100"`
	Color string `json:"color,omitempty"` // Hex color code, optional
	Icon  string `json:"icon,omitempty"`  // Icon name, optional
}

// UpdateChatFolderRequest for updating a folder
type UpdateChatFolderRequest struct {
	Name     string `json:"name,omitempty"`
	Color    string `json:"color,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Position *int   `json:"position,omitempty"`
}

// ChatFolderDTO for folder responses
type ChatFolderDTO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	Icon         string `json:"icon"`
	Position     int    `json:"position"`
	SessionCount int    `json:"session_count"`
	CreatedAt    string `json:"created_at"`
}

// ReorderFoldersRequest for batch reordering folders
type ReorderFoldersRequest struct {
	FolderIDs []uint `json:"folder_ids" binding:"required,min=1"`
}

// MoveSessionToFolderRequest for moving session to folder
type MoveSessionToFolderRequest struct {
	FolderID *uint `json:"folder_id"` // null to remove from folder
}

// ================================
// Chat Session DTOs
// ================================

type CreateChatSessionRequest struct {
	Title    string `json:"title" binding:"required,min=1,max=255"`
	FolderID *uint  `json:"folder_id,omitempty"` // Optional folder assignment
}

type ChatSessionDTO struct {
	ID                 uint             `json:"id"`
	Title              string           `json:"title"`
	FolderID           *uint            `json:"folder_id,omitempty"`
	FolderName         string           `json:"folder_name,omitempty"`
	Summary            *string          `json:"summary,omitempty"`
	SummaryGeneratedAt *time.Time       `json:"summary_generated_at,omitempty"`
	IsFavorite         bool             `json:"is_favorite"`
	IsTrash            bool             `json:"is_trash"`
	LastMessage        *ChatMessageDTO  `json:"last_message,omitempty"`
	Messages           []ChatMessageDTO `json:"messages,omitempty"`
	PinnedMessages     []ChatMessageDTO `json:"pinned_messages,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type ChatSessionListDTO struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	FolderID    *uint  `json:"folder_id,omitempty"`
	IsFavorite  bool   `json:"is_favorite"`
	IsTrash     bool   `json:"is_trash"`
	HasSummary  bool   `json:"has_summary"`
	LastMessage string `json:"last_message"`
	CreatedAt   string `json:"created_at"`
}

// ChatSessionSummaryDTO for summary generation response
type ChatSessionSummaryDTO struct {
	SessionID   uint      `json:"session_id"`
	Summary     string    `json:"summary"`
	MainTopics  []string  `json:"main_topics"`
	KeyInsights []string  `json:"key_insights"`
	ActionItems []string  `json:"action_items,omitempty"`
	Sentiment   string    `json:"sentiment"` // positive, neutral, negative, mixed
	GeneratedAt time.Time `json:"generated_at"`
}

// ================================
// Chat Message DTOs
// ================================

type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1"`
	Type    string `json:"type"` // "text" or "audio", defaults to "text"
}

type ChatMessageDTO struct {
	ID         uint      `json:"id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	IsLiked    bool      `json:"is_liked"`
	IsDisliked bool      `json:"is_disliked"`
	IsPinned   bool      `json:"is_pinned"`
	CreatedAt  time.Time `json:"created_at"`
}

// ================================
// Export DTOs
// ================================

// ExportFormat defines supported export formats
type ExportFormat string

const (
	ExportFormatPDF ExportFormat = "pdf"
	ExportFormatTXT ExportFormat = "txt"
)

// ExportChatRequest for exporting chat session
type ExportChatRequest struct {
	Format          ExportFormat `json:"format" binding:"required,oneof=pdf txt"`
	IncludePinned   bool         `json:"include_pinned"`   // Only export pinned messages
	IncludeMetadata bool         `json:"include_metadata"` // Include timestamps, session info
}

// ExportChatResponse returns export data
type ExportChatResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"` // Base64 encoded for PDF, plain text for TXT
	Size        int64  `json:"size"`    // File size in bytes
}

// ================================
// Suggested Prompts DTOs
// ================================

// SuggestedPromptDTO for suggested prompts
type SuggestedPromptDTO struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Category string `json:"category"` // mood, general, time_based, follow_up
	Icon     string `json:"icon,omitempty"`
}

// GetSuggestedPromptsRequest for getting suggested prompts
type GetSuggestedPromptsRequest struct {
	Mood          string `form:"mood,omitempty"`        // Current user mood
	TimeOfDay     string `form:"time_of_day,omitempty"` // morning, afternoon, evening, night
	HasMessages   bool   `form:"has_messages"`          // Whether session has existing messages
	LastAIMessage string `form:"-"`                     // Last AI message for context (internal use)
}

// SuggestedPromptsResponse returns suggested prompts
type SuggestedPromptsResponse struct {
	Prompts []SuggestedPromptDTO `json:"prompts"`
}

// ================================
// Query Params
// ================================

type ChatSessionQueryParams struct {
	Filter   string `form:"filter"` // all, bookmarked, favorites, trash
	Search   string `form:"search"`
	FolderID *uint  `form:"folder_id,omitempty"` // Filter by folder
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}
