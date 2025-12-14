package models

import (
	"time"

	"gorm.io/gorm"
)

// ArticleStatus represents the status of an article
type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusBlocked   ArticleStatus = "blocked"
)

type ArticleCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Articles []Article `gorm:"foreignKey:ArticleCategoryID" json:"articles,omitempty"`
}

func (ArticleCategory) TableName() string {
	return "article_categories"
}

type Article struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Title             string         `gorm:"size:255;not null" json:"title"`
	Thumbnail         string         `gorm:"size:500" json:"thumbnail"`
	Content           string         `gorm:"type:text;not null" json:"content"`
	ArticleCategoryID uint           `gorm:"not null" json:"article_category_id"`
	UserID            *uint          `gorm:"index" json:"user_id"`
	Status            ArticleStatus  `gorm:"size:20;default:'published'" json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category ArticleCategory `gorm:"foreignKey:ArticleCategoryID" json:"category,omitempty"`
	Author   *User           `gorm:"foreignKey:UserID" json:"author,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}
