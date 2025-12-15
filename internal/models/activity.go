package models

import (
	"time"
)

type UserActivity struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	ActivityType string    `gorm:"size:50;not null" json:"activity_type"`
	Date         time.Time `gorm:"type:date;not null" json:"date"`
	Count        int       `gorm:"default:0" json:"count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (UserActivity) TableName() string {
	return "user_activities"
}
