package dto

import (
	"time"

	"github.com/google/uuid"
)

// === MYSTERY CHEST DTOs ===

// UserChestResponse represents a chest
type UserChestResponse struct {
	ID                 uuid.UUID  `json:"id"`
	Rarity             string     `json:"rarity"`
	RarityIcon         string     `json:"rarity_icon"`
	IsOpened           bool       `json:"is_opened"`
	RewardType         string     `json:"reward_type,omitempty"`
	RewardValue        int        `json:"reward_value,omitempty"`
	RewardLabel        string     `json:"reward_label,omitempty"`
	TriggerType        string     `json:"trigger_type"`
	TriggerDescription string     `json:"trigger_description,omitempty"`
	OpenedAt           *time.Time `json:"opened_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// OpenChestResponse returns the reward from opening a chest
type OpenChestResponse struct {
	ChestID     uuid.UUID `json:"chest_id"`
	Rarity      string    `json:"rarity"`
	RewardType  string    `json:"reward_type"`
	RewardValue int       `json:"reward_value"`
	RewardLabel string    `json:"reward_label"`
}

// ChestFilterRequest for filtering chests
type ChestFilterRequest struct {
	IsOpened *bool  `form:"is_opened"`
	Rarity   string `form:"rarity"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
}
