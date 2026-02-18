package dto

import "time"

// ===== Journal Entry DTOs =====

// CreateJournalRequest represents request to create a journal entry
type CreateJournalRequest struct {
	Title       string   `json:"title"`
	Content     string   `json:"content" binding:"required,min=1"`
	MoodID      *uint    `json:"mood_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ShareWithAI *bool    `json:"share_with_ai,omitempty"` // Nil = use default from settings
}

// UpdateJournalRequest represents request to update a journal entry
type UpdateJournalRequest struct {
	Title       *string  `json:"title,omitempty"`
	Content     *string  `json:"content,omitempty"`
	MoodID      *uint    `json:"mood_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ShareWithAI *bool    `json:"share_with_ai,omitempty"`
}

// JournalResponse represents a journal entry response
type JournalResponse struct {
	ID             uint       `json:"id"`
	UUID           string     `json:"uuid"`
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Summary        string     `json:"summary,omitempty"`
	MoodID         *uint      `json:"mood_id,omitempty"`
	MoodLabel      string     `json:"mood_label,omitempty"`
	MoodEmoji      string     `json:"mood_emoji,omitempty"`
	Tags           []string   `json:"tags"`
	IsPrivate      bool       `json:"is_private"`
	ShareWithAI    bool       `json:"share_with_ai"`
	AIAccessedAt   *time.Time `json:"ai_accessed_at,omitempty"`
	WordCount      int        `json:"word_count"`
	SentimentScore *float64   `json:"sentiment_score,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// JournalListResponse represents a list item (without full content)
type JournalListResponse struct {
	ID           uint       `json:"id"`
	UUID         string     `json:"uuid"`
	Title        string     `json:"title"`
	Preview      string     `json:"preview"` // First 150 chars of content
	MoodID       *uint      `json:"mood_id,omitempty"`
	MoodLabel    string     `json:"mood_label,omitempty"`
	MoodEmoji    string     `json:"mood_emoji,omitempty"`
	Tags         []string   `json:"tags"`
	ShareWithAI  bool       `json:"share_with_ai"`
	AIAccessedAt *time.Time `json:"ai_accessed_at,omitempty"`
	WordCount    int        `json:"word_count"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ===== Journal Settings DTOs =====

// JournalSettingsRequest represents request to update journal settings
type JournalSettingsRequest struct {
	AllowAIAccess       *bool `json:"allow_ai_access,omitempty"`
	AIContextDays       *int  `json:"ai_context_days,omitempty"`
	AIContextMaxEntries *int  `json:"ai_context_max_entries,omitempty"`
	DefaultShareWithAI  *bool `json:"default_share_with_ai,omitempty"`
}

// JournalSettingsResponse represents journal settings response
type JournalSettingsResponse struct {
	AllowAIAccess       bool `json:"allow_ai_access"`
	AIContextDays       int  `json:"ai_context_days"`
	AIContextMaxEntries int  `json:"ai_context_max_entries"`
	DefaultShareWithAI  bool `json:"default_share_with_ai"`
	// Stats for UI
	TotalEntries      int  `json:"total_entries"`
	SharedWithAICount int  `json:"shared_with_ai_count"`
	IsBlocked         bool `json:"is_blocked"`
}

// ===== AI Integration DTOs =====

// JournalAIContextRequest represents request for AI context
type JournalAIContextRequest struct {
	Query          string `json:"query,omitempty"`           // Optional: search for relevant entries
	MaxEntries     int    `json:"max_entries,omitempty"`     // Override default max
	IncludeSummary bool   `json:"include_summary,omitempty"` // Whether to include summary
	DaysBack       int    `json:"days_back,omitempty"`       // Override default days
}

// JournalAIContext represents journal context for AI chatbot
type JournalAIContext struct {
	HasAccess     bool                    `json:"has_access"`        // Whether user allows AI access
	EntriesCount  int                     `json:"entries_count"`     // Number of entries included
	Entries       []JournalAIContextEntry `json:"entries"`           // The actual entries
	Summary       string                  `json:"summary,omitempty"` // AI-generated summary of patterns
	RecentMoods   []string                `json:"recent_moods"`
	CommonTags    []string                `json:"common_tags"`
	LastEntryDate *time.Time              `json:"last_entry_date,omitempty"`
}

// JournalAIContextEntry represents a single entry in AI context
type JournalAIContextEntry struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"` // May be truncated or summarized
	Mood      string    `json:"mood,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ===== AI Access Log DTOs =====

// JournalAIAccessLogResponse represents an AI access log entry
type JournalAIAccessLogResponse struct {
	ID            uint      `json:"id"`
	JournalID     uint      `json:"journal_id"`
	JournalTitle  string    `json:"journal_title"`
	ChatSessionID *uint     `json:"chat_session_id,omitempty"`
	ContextType   string    `json:"context_type"`
	AccessedAt    time.Time `json:"accessed_at"`
}

// ===== Journal Analytics DTOs =====

// JournalAnalytics represents analytics data for journals
type JournalAnalytics struct {
	TotalEntries     int                  `json:"total_entries"`
	EntriesThisMonth int                  `json:"entries_this_month"`
	TotalWordCount   int                  `json:"total_word_count"`
	AvgWordCount     int                  `json:"avg_word_count"`
	MoodDistribution map[string]int       `json:"mood_distribution"`
	TagFrequency     map[string]int       `json:"tag_frequency"`
	EntriesByMonth   []MonthlyEntryCount  `json:"entries_by_month"`
	SentimentTrend   []SentimentDataPoint `json:"sentiment_trend,omitempty"`
	WritingStreak    int                  `json:"writing_streak"` // Consecutive days
	LongestStreak    int                  `json:"longest_streak"`
}

// MonthlyEntryCount represents entry count for a month
type MonthlyEntryCount struct {
	Month string `json:"month"` // Format: "2025-01"
	Count int    `json:"count"`
}

// SentimentDataPoint represents sentiment score for a date
type SentimentDataPoint struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
}

// ===== AI-Powered Features DTOs =====

// JournalPromptResponse represents AI-generated writing prompt
type JournalPromptResponse struct {
	Prompt      string   `json:"prompt"`
	Category    string   `json:"category"` // 'reflection', 'gratitude', 'goal', 'emotion'
	RelatedTags []string `json:"related_tags,omitempty"`
}

// JournalWeeklySummary represents AI-generated weekly summary
type JournalWeeklySummary struct {
	WeekStart    time.Time `json:"week_start"`
	WeekEnd      time.Time `json:"week_end"`
	EntriesCount int       `json:"entries_count"`
	Summary      string    `json:"summary"`
	KeyThemes    []string  `json:"key_themes"`
	MoodTrend    string    `json:"mood_trend"` // 'improving', 'stable', 'declining'
	Insights     []string  `json:"insights"`
	Suggestions  []string  `json:"suggestions"`
}

// ===== Export DTOs =====

// JournalExportRequest represents request to export journals
type JournalExportRequest struct {
	Format    string   `json:"format" binding:"required,oneof=pdf txt"` // 'pdf' or 'txt'
	StartDate string   `json:"start_date,omitempty"`                    // "YYYY-MM-DD"
	EndDate   string   `json:"end_date,omitempty"`                      // "YYYY-MM-DD"
	Tags      []string `json:"tags,omitempty"`                          // Filter by tags
}

// JournalExportResponse represents export response
type JournalExportResponse struct {
	Format   string `json:"format"`
	Content  string `json:"content"` // Base64 encoded for PDF, plain text for TXT
	Filename string `json:"filename"`
}
