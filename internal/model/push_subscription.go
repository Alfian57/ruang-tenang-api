package model

import (
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_push_sub_user" json:"user_id"`
	Endpoint  string    `gorm:"type:text;not null;uniqueIndex" json:"endpoint"`
	P256dh    string    `gorm:"type:text;not null" json:"p256dh"`
	Auth      string    `gorm:"type:text;not null" json:"auth"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PushSubscription) TableName() string {
	return "push_subscriptions"
}
