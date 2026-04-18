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
	UUID         string `json:"uuid"`
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
	UUID               string           `json:"uuid"`
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
	UUID        string `json:"uuid"`
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
	Content  string               `json:"content" binding:"required,min=1"`
	Type     string               `json:"type"` // "text" or "audio", defaults to "text"
	Context  *MessageContextHints `json:"context,omitempty"`
	Metadata *SendMessageMetadata `json:"metadata,omitempty"`
}

type MessageContextHints struct {
	CurrentMood              string `json:"current_mood,omitempty"`
	SessionIntent            string `json:"session_intent,omitempty"`
	EnableMoodContext        *bool  `json:"enable_mood_context,omitempty"`
	EnableJournalContext     *bool  `json:"enable_journal_context,omitempty"`
	EnableDailyTaskContext   *bool  `json:"enable_daily_task_context,omitempty"`
	EnableXPLevelContext     *bool  `json:"enable_xp_level_context,omitempty"`
	EnableBreathingContext   *bool  `json:"enable_breathing_context,omitempty"`
	EnablePlaylistContext    *bool  `json:"enable_playlist_context,omitempty"`
	EnableRewardsContext     *bool  `json:"enable_rewards_context,omitempty"`
	EnableProgressMapContext *bool  `json:"enable_progress_map_context,omitempty"`
	EnableSocialContext      *bool  `json:"enable_social_context,omitempty"`
}

type SendMessageMetadata struct {
	Source   string `json:"source,omitempty"`
	PromptID string `json:"prompt_id,omitempty"`
}

type UpdateChatContextPreferencesRequest struct {
	EnableMoodContext        *bool   `json:"enable_mood_context,omitempty"`
	EnableJournalContext     *bool   `json:"enable_journal_context,omitempty"`
	EnableDailyTaskContext   *bool   `json:"enable_daily_task_context,omitempty"`
	EnableXPLevelContext     *bool   `json:"enable_xp_level_context,omitempty"`
	EnableBreathingContext   *bool   `json:"enable_breathing_context,omitempty"`
	EnablePlaylistContext    *bool   `json:"enable_playlist_context,omitempty"`
	EnableRewardsContext     *bool   `json:"enable_rewards_context,omitempty"`
	EnableProgressMapContext *bool   `json:"enable_progress_map_context,omitempty"`
	EnableSocialContext      *bool   `json:"enable_social_context,omitempty"`
	SessionIntent            *string `json:"session_intent,omitempty"`
}

type ChatContextPreferencesDTO struct {
	EnableMoodContext        bool   `json:"enable_mood_context"`
	EnableJournalContext     bool   `json:"enable_journal_context"`
	EnableDailyTaskContext   bool   `json:"enable_daily_task_context"`
	EnableXPLevelContext     bool   `json:"enable_xp_level_context"`
	EnableBreathingContext   bool   `json:"enable_breathing_context"`
	EnablePlaylistContext    bool   `json:"enable_playlist_context"`
	EnableRewardsContext     bool   `json:"enable_rewards_context"`
	EnableProgressMapContext bool   `json:"enable_progress_map_context"`
	EnableSocialContext      bool   `json:"enable_social_context"`
	SessionIntent            string `json:"session_intent"`
}

type ChatContextMoodDTO struct {
	Mood  string `json:"mood"`
	Emoji string `json:"emoji"`
}

type ChatContextDailyTaskDTO struct {
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
}

type ChatContextXPLevelDTO struct {
	Exp           int64 `json:"exp"`
	CurrentStreak int   `json:"current_streak"`
	CurrentLevel  int   `json:"current_level,omitempty"`
	NextLevel     int   `json:"next_level,omitempty"`
}

type ChatContextBreathingDTO struct {
	SessionsToday     int    `json:"sessions_today"`
	SessionsLast7Days int    `json:"sessions_last_7_days"`
	MostUsedTechnique string `json:"most_used_technique,omitempty"`
}

type ChatContextPlaylistDTO struct {
	TotalPlaylists      int    `json:"total_playlists"`
	TotalSavedSongs     int    `json:"total_saved_songs"`
	LatestPlaylistTitle string `json:"latest_playlist_title,omitempty"`
}

type ChatContextRewardsDTO struct {
	GoldCoins        int64  `json:"gold_coins"`
	ClaimCount       int    `json:"claim_count"`
	LatestRewardName string `json:"latest_reward_name,omitempty"`
}

type ChatContextProgressMapDTO struct {
	UnlockedRegions   int    `json:"unlocked_regions"`
	UnlockedLandmarks int    `json:"unlocked_landmarks"`
	LatestUnlockName  string `json:"latest_unlock_name,omitempty"`
}

type ChatContextSocialDTO struct {
	BadgeCount       int    `json:"badge_count"`
	GuildName        string `json:"guild_name,omitempty"`
	GuildRole        string `json:"guild_role,omitempty"`
	GuildMemberCount int    `json:"guild_member_count,omitempty"`
}

type ChatContextRuntimeDTO struct {
	Mood               *ChatContextMoodDTO        `json:"mood,omitempty"`
	JournalSharedCount int                        `json:"journal_shared_count"`
	DailyTask          *ChatContextDailyTaskDTO   `json:"daily_task,omitempty"`
	XPLevel            *ChatContextXPLevelDTO     `json:"xp_level,omitempty"`
	Breathing          *ChatContextBreathingDTO   `json:"breathing,omitempty"`
	Playlist           *ChatContextPlaylistDTO    `json:"playlist,omitempty"`
	Rewards            *ChatContextRewardsDTO     `json:"rewards,omitempty"`
	ProgressMap        *ChatContextProgressMapDTO `json:"progress_map,omitempty"`
	Social             *ChatContextSocialDTO      `json:"social,omitempty"`
	EffectiveSources   []string                   `json:"effective_sources"`
}

type ChatContextStateDTO struct {
	SessionUUID string                    `json:"session_uuid"`
	Preferences ChatContextPreferencesDTO `json:"preferences"`
	Runtime     ChatContextRuntimeDTO     `json:"runtime"`
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
