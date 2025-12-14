package models

import (
	"time"

	"gorm.io/gorm"
)

type SongCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
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

type Song struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Title          string         `gorm:"size:255;not null" json:"title"`
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
