package model

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/pkg/slug"
	"gorm.io/gorm"
)

// ArticleStatus represents the status of an article
type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusBlocked   ArticleStatus = "blocked"
)

// ArticleModerationStatus represents the moderation status
type ArticleModerationStatus string

const (
	ArticleModerationPending        ArticleModerationStatus = "pending"
	ArticleModerationApproved       ArticleModerationStatus = "approved"
	ArticleModerationFlagged        ArticleModerationStatus = "flagged"
	ArticleModerationRejected       ArticleModerationStatus = "rejected"
	ArticleModerationRevisionNeeded ArticleModerationStatus = "revision_needed"
)

type ArticleCategory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Slug        string         `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Articles []Article `gorm:"foreignKey:ArticleCategoryID" json:"articles,omitempty"`
}

func (ArticleCategory) TableName() string {
	return "article_categories"
}

// BeforeCreate generates a slug from Name if not already set
func (c *ArticleCategory) BeforeCreate(tx *gorm.DB) error {
	if c.Slug == "" {
		c.Slug = slug.GenerateUnique(c.Name)
	}
	return nil
}

type Article struct {
	ID                uint          `gorm:"primaryKey" json:"id"`
	Title             string        `gorm:"size:255;not null" json:"title"`
	Slug              string        `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	Thumbnail         string        `gorm:"size:500" json:"thumbnail"`
	Content           string        `gorm:"type:text;not null" json:"content"`
	ArticleCategoryID uint          `gorm:"not null" json:"article_category_id"`
	UserID            uint          `gorm:"index;not null" json:"user_id"`
	Status            ArticleStatus `gorm:"size:20;default:'draft'" json:"status"`

	// Moderation fields
	ModerationStatus ArticleModerationStatus `gorm:"size:50;default:'pending'" json:"moderation_status"`
	ModerationNotes  string                  `gorm:"type:text" json:"moderation_notes,omitempty"`
	ModeratedByID    *uint                   `json:"moderated_by_id,omitempty"`
	ModeratedAt      *time.Time              `json:"moderated_at,omitempty"`
	TriggerWarnings  TriggerWarnings         `gorm:"type:json" json:"trigger_warnings,omitempty"`
	IsUserGenerated  bool                    `gorm:"default:false" json:"is_user_generated"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category    ArticleCategory `gorm:"foreignKey:ArticleCategoryID" json:"category,omitempty"`
	Author      *User           `gorm:"foreignKey:UserID" json:"author,omitempty"`
	ModeratedBy *User           `gorm:"foreignKey:ModeratedByID" json:"moderated_by,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}

// BeforeCreate generates a slug from Title if not already set
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.Slug == "" {
		a.Slug = slug.GenerateUnique(a.Title)
	}
	return nil
}

// IsPublic returns true if article is publicly visible
func (a *Article) IsPublic() bool {
	return a.Status == ArticleStatusPublished &&
		(a.ModerationStatus == ArticleModerationApproved || !a.IsUserGenerated)
}

// NeedsModeration returns true if article needs moderator review
func (a *Article) NeedsModeration() bool {
	return a.IsUserGenerated && a.ModerationStatus == ArticleModerationFlagged
}

// HasTriggerWarnings returns true if article has any trigger warnings
func (a *Article) HasTriggerWarnings() bool {
	return len(a.TriggerWarnings) > 0
}
