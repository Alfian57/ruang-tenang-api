package dto

import "time"

// Level Config DTOs
type LevelConfigDTO struct {
	ID        uint      `json:"id"`
	Level     int       `json:"level"`
	MinExp    int       `json:"min_exp"`
	BadgeName string    `json:"badge_name"`
	BadgeIcon string    `json:"badge_icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateLevelConfigRequest struct {
	Level     int    `json:"level" binding:"gte=1"`
	MinExp    int    `json:"min_exp" binding:"gte=0"`
	BadgeName string `json:"badge_name" binding:"required,min=1,max=100"`
	BadgeIcon string `json:"badge_icon" binding:"required,min=1,max=50"`
}

type UpdateLevelConfigRequest struct {
	Level     int    `json:"level" binding:"gte=1"`
	MinExp    int    `json:"min_exp" binding:"gte=0"`
	BadgeName string `json:"badge_name" binding:"required,min=1,max=100"`
	BadgeIcon string `json:"badge_icon" binding:"required,min=1,max=50"`
}

// EXP History DTOs
type ExpHistoryDTO struct {
	ID           uint      `json:"id"`
	ActivityType string    `json:"activity_type"`
	Points       int       `json:"points"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

type ExpHistoryResponse struct {
	Data       []ExpHistoryDTO `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

type ExpHistoryFilterRequest struct {
	ActivityType string `form:"activity_type"`
	StartDate    string `form:"start_date"` // Format: YYYY-MM-DD
	EndDate      string `form:"end_date"`   // Format: YYYY-MM-DD
	Page         int    `form:"page,default=1"`
	Limit        int    `form:"limit,default=10"`
}

// User level info for frontend
type UserLevelInfo struct {
	Level          int    `json:"level"`
	BadgeName      string `json:"badge_name"`
	BadgeIcon      string `json:"badge_icon"`
	CurrentExp     int64  `json:"current_exp"`
	NextLevelExp   *int   `json:"next_level_exp,omitempty"`
	ExpToNextLevel *int   `json:"exp_to_next_level,omitempty"`
}
