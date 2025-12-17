package models

import (
	"time"

	"gorm.io/gorm"
)

type Forum struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null" json:"user_id"`
	CategoryID *uint          `json:"category_id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Content    string         `gorm:"type:text" json:"content"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category     *ForumCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Posts        []ForumPost    `gorm:"foreignKey:ForumID" json:"posts,omitempty"`
	Likes        []ForumLike    `gorm:"foreignKey:ForumID" json:"likes,omitempty"`
	LikesCount   int64          `gorm:"-" json:"likes_count"`
	RepliesCount int64          `gorm:"-" json:"replies_count"`
	IsLiked      bool           `gorm:"-" json:"is_liked"`
}

func (Forum) TableName() string {
	return "forums"
}

type ForumPost struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ForumID   uint           `gorm:"not null" json:"forum_id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Forum Forum `gorm:"foreignKey:ForumID" json:"forum,omitempty"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ForumPost) TableName() string {
	return "forum_posts"
}
