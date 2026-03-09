package dto

import "time"

type CreateBroadcastRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Body        string  `json:"body" binding:"required"`
	Icon        string  `json:"icon"`
	URL         string  `json:"url"`
	ScheduledAt *string `json:"scheduled_at"`
}

type UpdateBroadcastRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Body        string  `json:"body" binding:"required"`
	Icon        string  `json:"icon"`
	URL         string  `json:"url"`
	ScheduledAt *string `json:"scheduled_at"`
}

type BroadcastResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Icon        string     `json:"icon"`
	URL         string     `json:"url"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`
	SentCount   int        `json:"sent_count"`
	FailedCount int        `json:"failed_count"`
	CreatedBy   uint       `json:"created_by"`
	CreatorName string     `json:"creator_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
