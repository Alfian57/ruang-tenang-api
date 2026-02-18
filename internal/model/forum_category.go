package model

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/pkg/slug"
	"gorm.io/gorm"
)

type ForumCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Slug      string         `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ForumCategory) TableName() string {
	return "forum_categories"
}

// BeforeCreate generates a slug from Name if not already set
func (c *ForumCategory) BeforeCreate(tx *gorm.DB) error {
	if c.Slug == "" {
		c.Slug = slug.GenerateUnique(c.Name)
	}
	return nil
}
