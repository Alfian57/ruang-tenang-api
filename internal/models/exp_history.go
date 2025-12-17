package models

import (
	"time"
)

type ExpHistory struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	ActivityType string    `gorm:"size:50;not null" json:"activity_type"`
	Points       int       `gorm:"not null" json:"points"`
	Description  string    `gorm:"size:255" json:"description"`
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ExpHistory) TableName() string {
	return "exp_histories"
}
