package model

import (
	"time"

	"github.com/google/uuid"
)

type BroadcastStatus string

const (
	BroadcastStatusDraft     BroadcastStatus = "draft"
	BroadcastStatusScheduled BroadcastStatus = "scheduled"
	BroadcastStatusSending   BroadcastStatus = "sending"
	BroadcastStatusSent      BroadcastStatus = "sent"
	BroadcastStatusCancelled BroadcastStatus = "cancelled"
)

type BroadcastNotification struct {
	ID          uuid.UUID       `gorm:"type:char(36);primaryKey" json:"id"`
	Title       string          `gorm:"type:varchar(255);not null" json:"title"`
	Body        string          `gorm:"type:text;not null" json:"body"`
	Icon        string          `gorm:"type:varchar(500)" json:"icon"`
	URL         string          `gorm:"type:varchar(500);column:url" json:"url"`
	Status      BroadcastStatus `gorm:"type:varchar(20);not null;default:draft" json:"status"`
	ScheduledAt *time.Time      `json:"scheduled_at"`
	SentAt      *time.Time      `json:"sent_at"`
	SentCount   int             `gorm:"not null;default:0" json:"sent_count"`
	FailedCount int             `gorm:"not null;default:0" json:"failed_count"`
	CreatedBy   uint            `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (BroadcastNotification) TableName() string {
	return "broadcast_notifications"
}
