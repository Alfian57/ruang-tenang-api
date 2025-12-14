package dto

import "time"

// User Mood DTOs
type CreateMoodRequest struct {
	Mood string `json:"mood" binding:"required,oneof=happy neutral angry disappointed sad crying"`
}

type UserMoodDTO struct {
	ID        uint      `json:"id"`
	Mood      string    `json:"mood"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type MoodHistoryDTO struct {
	Moods      []UserMoodDTO `json:"moods"`
	TotalCount int64         `json:"total_count"`
}

// Query params
type MoodQueryParams struct {
	StartDate string `form:"start_date"` // YYYY-MM-DD
	EndDate   string `form:"end_date"`   // YYYY-MM-DD
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=30"`
}
