package models

import (
	"time"
)

type ForumLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ForumID   uint      `gorm:"not null" json:"forum_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (ForumLike) TableName() string {
	return "forum_likes"
}
