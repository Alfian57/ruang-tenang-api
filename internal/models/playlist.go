package models

import (
	"time"

	"gorm.io/gorm"
)

// Playlist represents a user's custom playlist
type Playlist struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null" json:"user_id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	Thumbnail       string         `gorm:"size:500" json:"thumbnail"`
	IsPublic        bool           `gorm:"default:false" json:"is_public"`
	IsAdminPlaylist bool           `gorm:"default:false" json:"is_admin_playlist"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User  User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items []PlaylistItem `gorm:"foreignKey:PlaylistID" json:"items,omitempty"`
}

func (Playlist) TableName() string {
	return "playlists"
}

// PlaylistItem represents a song in a playlist with ordering
type PlaylistItem struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PlaylistID uint           `gorm:"not null" json:"playlist_id"`
	SongID     uint           `gorm:"not null" json:"song_id"`
	Position   int            `gorm:"not null;default:0" json:"position"`
	AddedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"added_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Playlist Playlist `gorm:"foreignKey:PlaylistID" json:"playlist,omitempty"`
	Song     Song     `gorm:"foreignKey:SongID" json:"song,omitempty"`
}

func (PlaylistItem) TableName() string {
	return "playlist_items"
}
