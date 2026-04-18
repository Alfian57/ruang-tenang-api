package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRole string

type ChatSessionIntent string

const (
	ChatRoleUser ChatRole = "user"
	ChatRoleAI   ChatRole = "ai"

	ChatSessionIntentGeneral    ChatSessionIntent = "general"
	ChatSessionIntentGrounding  ChatSessionIntent = "grounding"
	ChatSessionIntentPlanning   ChatSessionIntent = "planning"
	ChatSessionIntentReflection ChatSessionIntent = "reflection"
	ChatSessionIntentCoping     ChatSessionIntent = "coping"
)

// ChatFolder represents a folder for organizing chat sessions (1 level hierarchy only)
type ChatFolder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UUID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()" json:"uuid"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Color     string    `gorm:"size:7;default:'#6366f1'" json:"color"` // Hex color code
	Icon      string    `gorm:"size:50;default:'folder'" json:"icon"`
	Position  int       `gorm:"default:0" json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User     User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Sessions []ChatSession `gorm:"foreignKey:FolderID" json:"sessions,omitempty"`
}

func (ChatFolder) TableName() string {
	return "chat_folders"
}

type ChatSession struct {
	ID                       uint              `gorm:"primaryKey" json:"id"`
	UUID                     uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()" json:"uuid"`
	UserID                   uint              `gorm:"not null" json:"user_id"`
	FolderID                 *uint             `gorm:"index" json:"folder_id,omitempty"` // Optional folder assignment
	Title                    string            `gorm:"size:255;not null" json:"title"`
	Summary                  *string           `gorm:"type:text" json:"summary,omitempty"` // AI-generated summary
	SummaryGeneratedAt       *time.Time        `json:"summary_generated_at,omitempty"`
	IsFavorite               bool              `gorm:"default:false" json:"is_favorite"`
	IsTrash                  bool              `gorm:"default:false" json:"is_trash"`
	EnableMoodContext        bool              `gorm:"default:true" json:"enable_mood_context"`
	EnableJournalContext     bool              `gorm:"default:false" json:"enable_journal_context"`
	EnableDailyTaskContext   bool              `gorm:"default:true" json:"enable_daily_task_context"`
	EnableXPLevelContext     bool              `gorm:"default:true" json:"enable_xp_level_context"`
	EnableBreathingContext   bool              `gorm:"default:true" json:"enable_breathing_context"`
	EnablePlaylistContext    bool              `gorm:"default:false" json:"enable_playlist_context"`
	EnableRewardsContext     bool              `gorm:"default:false" json:"enable_rewards_context"`
	EnableProgressMapContext bool              `gorm:"default:false" json:"enable_progress_map_context"`
	EnableSocialContext      bool              `gorm:"default:false" json:"enable_social_context"`
	SessionIntent            ChatSessionIntent `gorm:"size:30;not null;default:'general'" json:"session_intent"`
	CreatedAt                time.Time         `json:"created_at"`
	UpdatedAt                time.Time         `json:"updated_at"`
	DeletedAt                gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relations
	User     User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Folder   *ChatFolder   `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	Messages []ChatMessage `gorm:"foreignKey:ChatSessionID" json:"messages,omitempty"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

// GetPinnedMessages returns only pinned messages from the session
func (s *ChatSession) GetPinnedMessages() []ChatMessage {
	var pinned []ChatMessage
	for _, msg := range s.Messages {
		if msg.IsPinned {
			pinned = append(pinned, msg)
		}
	}
	return pinned
}

type ChatMessage struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UUID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()" json:"uuid"`
	ChatSessionID uint      `gorm:"not null" json:"chat_session_id"`
	Role          ChatRole  `gorm:"type:varchar(10);not null" json:"role"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	Type          string    `gorm:"type:varchar(20);default:'text'" json:"type"` // text, audio
	IsLiked       bool      `gorm:"default:false" json:"is_liked"`
	IsDisliked    bool      `gorm:"default:false" json:"is_disliked"`
	IsPinned      bool      `gorm:"default:false" json:"is_pinned"` // Pin important AI messages
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations
	ChatSession ChatSession `gorm:"foreignKey:ChatSessionID" json:"chat_session,omitempty"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

func (m *ChatMessage) IsAI() bool {
	return m.Role == ChatRoleAI
}

func (m *ChatMessage) IsUser() bool {
	return m.Role == ChatRoleUser
}
