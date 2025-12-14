package dto

import "time"

// Song Category DTOs
type SongCategoryDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Thumbnail string    `json:"thumbnail"`
	SongCount int       `json:"song_count,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Song DTOs
type SongDTO struct {
	ID         uint            `json:"id"`
	Title      string          `json:"title"`
	FilePath   string          `json:"file_path"`
	Thumbnail  string          `json:"thumbnail"`
	CategoryID uint            `json:"category_id"`
	Category   SongCategoryDTO `json:"category,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type SongListDTO struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	FilePath   string `json:"file_path"`
	Thumbnail  string `json:"thumbnail"`
	CategoryID uint   `json:"category_id"`
}
