package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatRole string

const (
	ChatRoleUser ChatRole = "user"
	ChatRoleAI   ChatRole = "ai"
)

type ChatSession struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	Title        string         `gorm:"size:255;not null" json:"title"`
	IsBookmarked bool           `gorm:"default:false" json:"is_bookmarked"`
	IsFavorite   bool           `gorm:"default:false" json:"is_favorite"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Messages []ChatMessage `gorm:"foreignKey:ChatSessionID" json:"messages,omitempty"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

type ChatMessage struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ChatSessionID uint      `gorm:"not null" json:"chat_session_id"`
	Role          ChatRole  `gorm:"type:varchar(10);not null" json:"role"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	IsLiked       bool      `gorm:"default:false" json:"is_liked"`
	IsDisliked    bool      `gorm:"default:false" json:"is_disliked"`
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
