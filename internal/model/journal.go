package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Journal represents a private journal entry
type Journal struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UUID           uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()" json:"uuid"`
	UserID         uint           `gorm:"not null" json:"user_id"`
	Title          string         `gorm:"size:255" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Summary        string         `gorm:"type:text" json:"summary"`
	MoodID         *uint          `json:"mood_id,omitempty"`
	Tags           pq.StringArray `gorm:"type:text[]" json:"tags"`
	IsPrivate      bool           `gorm:"default:true" json:"is_private"`
	ShareWithAI    bool           `gorm:"default:false" json:"share_with_ai"`
	AIAccessedAt   *time.Time     `json:"ai_accessed_at,omitempty"`
	WordCount      int            `gorm:"default:0" json:"word_count"`
	SentimentScore *float64       `gorm:"type:decimal(3,2)" json:"sentiment_score,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	// Relations
	User User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Mood *UserMood `gorm:"foreignKey:MoodID" json:"mood,omitempty"`
}

func (Journal) TableName() string {
	return "journals"
}

// JournalSettings represents user's global journal privacy settings
type JournalSettings struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	UserID              uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	AllowAIAccess       bool      `gorm:"default:false" json:"allow_ai_access"`
	AIContextDays       int       `gorm:"default:7" json:"ai_context_days"`
	AIContextMaxEntries int       `gorm:"default:5" json:"ai_context_max_entries"`
	DefaultShareWithAI  bool      `gorm:"default:false" json:"default_share_with_ai"`
	IsBlocked           bool      `gorm:"default:false" json:"is_blocked"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (JournalSettings) TableName() string {
	return "journal_settings"
}

// JournalAIAccessLog tracks when AI accessed journal entries for transparency
type JournalAIAccessLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	JournalID     uint      `gorm:"not null" json:"journal_id"`
	ChatSessionID *uint     `json:"chat_session_id,omitempty"`
	AccessedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"accessed_at"`
	ContextType   string    `gorm:"size:50" json:"context_type"` // 'full', 'summary', 'keyword_match'

	// Relations
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Journal     Journal      `gorm:"foreignKey:JournalID" json:"journal,omitempty"`
	ChatSession *ChatSession `gorm:"foreignKey:ChatSessionID" json:"chat_session,omitempty"`
}

func (JournalAIAccessLog) TableName() string {
	return "journal_ai_access_logs"
}

// GetMoodLabel returns a human-readable mood label
func (j *Journal) GetMoodLabel() string {
	if j.Mood == nil {
		return ""
	}
	return string(j.Mood.Mood)
}

// GetMoodEmoji returns the emoji for the journal's mood
func (j *Journal) GetMoodEmoji() string {
	if j.Mood == nil {
		return ""
	}
	return j.Mood.GetMoodEmoji()
}
