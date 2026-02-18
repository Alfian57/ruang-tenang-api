package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeHeart         NotificationType = "heart"
	NotificationTypeStoryApproved NotificationType = "story_approved"
	NotificationTypeStoryRejected NotificationType = "story_rejected"
	NotificationTypeBadgeEarned   NotificationType = "badge_earned"
	NotificationTypeLevelUp       NotificationType = "level_up"
)

type Notification struct {
	ID        uuid.UUID        `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uint             `gorm:"not null;index:idx_notification_user_read" json:"user_id"`
	Type      NotificationType `gorm:"size:50;not null" json:"type"`
	Title     string           `gorm:"size:255;not null" json:"title"`
	Message   string           `gorm:"size:500;not null" json:"message"`
	IsRead    bool             `gorm:"default:false;index:idx_notification_user_read" json:"is_read"`
	Data      string           `gorm:"type:text" json:"data,omitempty"` // JSON metadata (story_id, badge_key, etc.)
	CreatedAt time.Time        `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
