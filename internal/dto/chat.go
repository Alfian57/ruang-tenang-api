package dto

import "time"

// Chat Session DTOs
type CreateChatSessionRequest struct {
	Title string `json:"title" binding:"required,min=1,max=255"`
}

type ChatSessionDTO struct {
	ID           uint             `json:"id"`
	Title        string           `json:"title"`
	IsBookmarked bool             `json:"is_bookmarked"`
	IsFavorite   bool             `json:"is_favorite"`
	LastMessage  *ChatMessageDTO  `json:"last_message,omitempty"`
	Messages     []ChatMessageDTO `json:"messages,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type ChatSessionListDTO struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	IsBookmarked bool   `json:"is_bookmarked"`
	IsFavorite   bool   `json:"is_favorite"`
	LastMessage  string `json:"last_message"`
	CreatedAt    string `json:"created_at"`
}

// Chat Message DTOs
type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

type ChatMessageDTO struct {
	ID         uint      `json:"id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	IsLiked    bool      `json:"is_liked"`
	IsDisliked bool      `json:"is_disliked"`
	CreatedAt  time.Time `json:"created_at"`
}

// Query params
type ChatSessionQueryParams struct {
	Filter string `form:"filter"` // all, bookmarked, favorites
	Search string `form:"search"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
}
