package dto

import "time"

// Article DTOs
type ArticleCategoryDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ArticleAuthorDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ArticleDTO struct {
	ID         uint               `json:"id"`
	Title      string             `json:"title"`
	Thumbnail  string             `json:"thumbnail"`
	Content    string             `json:"content"`
	CategoryID uint               `json:"category_id"`
	Category   ArticleCategoryDTO `json:"category,omitempty"`
	UserID     *uint              `json:"user_id,omitempty"`
	Author     *ArticleAuthorDTO  `json:"author,omitempty"`
	Status     string             `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type ArticleListDTO struct {
	ID         uint               `json:"id"`
	Title      string             `json:"title"`
	Thumbnail  string             `json:"thumbnail"`
	Excerpt    string             `json:"excerpt"`
	CategoryID uint               `json:"category_id"`
	Category   ArticleCategoryDTO `json:"category,omitempty"`
	UserID     *uint              `json:"user_id,omitempty"`
	Author     *ArticleAuthorDTO  `json:"author,omitempty"`
	Status     string             `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
}

// User request DTOs (for members)
type CreateUserArticleRequest struct {
	Title      string `json:"title" binding:"required"`
	Thumbnail  string `json:"thumbnail"`
	Content    string `json:"content" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

type UpdateUserArticleRequest struct {
	Title      string `json:"title" binding:"required"`
	Thumbnail  string `json:"thumbnail"`
	Content    string `json:"content" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

// Admin request DTOs
type CreateArticleRequest struct {
	Title      string `json:"title" binding:"required"`
	Thumbnail  string `json:"thumbnail"`
	Content    string `json:"content" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

type UpdateArticleRequest struct {
	Title      string `json:"title" binding:"required"`
	Thumbnail  string `json:"thumbnail"`
	Content    string `json:"content" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

type CreateArticleCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateSongCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	Thumbnail string `json:"thumbnail"`
}

type CreateSongRequest struct {
	Title      string `json:"title" binding:"required"`
	FilePath   string `json:"file_path" binding:"required"`
	Thumbnail  string `json:"thumbnail"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

// Query params
type ArticleQueryParams struct {
	CategoryID uint   `form:"category_id"`
	Search     string `form:"search"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=10"`
	Status     string `form:"status"`
	UserID     uint   `form:"user_id"`
}
