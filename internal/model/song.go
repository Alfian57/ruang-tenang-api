package model

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/pkg/slug"
	"gorm.io/gorm"
)

type SongCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Slug      string         `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	Thumbnail string         `gorm:"size:500" json:"thumbnail"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Songs []Song `gorm:"foreignKey:SongCategoryID" json:"songs,omitempty"`
}

func (SongCategory) TableName() string {
	return "song_categories"
}

// BeforeCreate generates a slug from Name if not already set
func (c *SongCategory) BeforeCreate(tx *gorm.DB) error {
	if c.Slug == "" {
		c.Slug = slug.GenerateUnique(c.Name)
	}
	return nil
}

type Song struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Title          string         `gorm:"size:255;not null" json:"title"`
	Slug           string         `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	FilePath       string         `gorm:"size:500;not null" json:"file_path"`
	Thumbnail      string         `gorm:"size:500" json:"thumbnail"`
	SongCategoryID uint           `gorm:"not null" json:"song_category_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category SongCategory `gorm:"foreignKey:SongCategoryID" json:"category,omitempty"`
}

func (Song) TableName() string {
	return "songs"
}

// BeforeCreate generates a slug from Title if not already set
func (s *Song) BeforeCreate(tx *gorm.DB) error {
	if s.Slug == "" {
		s.Slug = slug.GenerateUnique(s.Title)
	}
	return nil
}
