package dto

import "time"

// Playlist DTOs

// CreatePlaylistRequest represents the request to create a new playlist
type CreatePlaylistRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	IsPublic    bool   `json:"is_public"`
}

// UpdatePlaylistRequest represents the request to update a playlist
type UpdatePlaylistRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	IsPublic    bool   `json:"is_public"`
}

// AddSongToPlaylistRequest represents the request to add a song to a playlist
type AddSongToPlaylistRequest struct {
	SongID uint `json:"song_id" binding:"required"`
}

// AddSongsToPlaylistRequest represents the request to add multiple songs to a playlist
type AddSongsToPlaylistRequest struct {
	SongIDs []uint `json:"song_ids" binding:"required,min=1"`
}

// ReorderPlaylistItemsRequest represents the request to reorder songs in a playlist
type ReorderPlaylistItemsRequest struct {
	ItemIDs []uint `json:"item_ids" binding:"required,min=1"`
}

// PlaylistDTO represents a playlist response
type PlaylistDTO struct {
	ID          uint              `json:"id"`
	UUID        string            `json:"uuid"`
	UserID      uint              `json:"user_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Thumbnail   string            `json:"thumbnail"`
	IsPublic    bool              `json:"is_public"`
	ItemCount   int               `json:"item_count"`
	TotalSongs  int               `json:"total_songs"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	User        *UserBasicDTO     `json:"user,omitempty"`
	Items       []PlaylistItemDTO `json:"items,omitempty"`
}

// PlaylistItemDTO represents a playlist item response
type PlaylistItemDTO struct {
	ID         uint      `json:"id"`
	UUID       string    `json:"uuid"`
	PlaylistID uint      `json:"playlist_id"`
	SongID     uint      `json:"song_id"`
	Position   int       `json:"position"`
	AddedAt    time.Time `json:"added_at"`
	Song       *SongDTO  `json:"song,omitempty"`
}

// UserBasicDTO represents basic user information
type UserBasicDTO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// PlaylistListDTO represents a simplified playlist for list views
type PlaylistListDTO struct {
	ID          uint      `json:"id"`
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	IsPublic    bool      `json:"is_public"`
	ItemCount   int       `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
