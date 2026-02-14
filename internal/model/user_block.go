package model

import (
	"time"
)

// UserBlock represents a blocking relationship between users
type UserBlock struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BlockerID uint      `gorm:"not null" json:"blocker_id"`
	BlockedID uint      `gorm:"not null" json:"blocked_id"`
	Reason    string    `gorm:"size:255" json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Blocker *User `gorm:"foreignKey:BlockerID" json:"blocker,omitempty"`
	Blocked *User `gorm:"foreignKey:BlockedID" json:"blocked,omitempty"`
}

func (UserBlock) TableName() string {
	return "user_blocks"
}
