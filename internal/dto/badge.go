package dto

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
)

// BadgeDefinitionResponse represents a badge definition for API responses
type BadgeDefinitionResponse struct {
	ID               uuid.UUID                   `json:"id"`
	BadgeKey         string                      `json:"badge_key"`
	BadgeName        string                      `json:"badge_name"`
	Description      string                      `json:"description"`
	Icon             string                      `json:"icon"`
	Category         string                      `json:"category"`
	RequirementType  model.BadgeRequirementType `json:"requirement_type"`
	RequirementValue int                         `json:"requirement_value"`
}

// BadgeCategoryInfo represents badge category metadata
type BadgeCategoryInfo struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// BadgeCategoryStats represents badge statistics per category
type BadgeCategoryStats struct {
	Category string `json:"category"`
	Earned   int    `json:"earned"`
	Total    int    `json:"total"`
}

// BadgeProgressResponse represents progress towards earning a badge
type BadgeProgressResponse struct {
	BadgeID         uuid.UUID `json:"badge_id"`
	BadgeKey        string    `json:"badge_key"`
	BadgeName       string    `json:"badge_name"`
	Description     string    `json:"description"`
	Icon            string    `json:"icon"`
	Category        string    `json:"category"`
	Earned          bool      `json:"earned"`
	CurrentValue    int       `json:"current_value"`
	TargetValue     int       `json:"target_value"`
	ProgressPercent float64   `json:"progress_percent"`
}
