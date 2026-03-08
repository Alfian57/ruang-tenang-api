package model

import (
	"time"
)

// RewardType constants
type RewardType string

const (
	RewardTypeGeneral RewardType = "general"
	RewardTypeTheme   RewardType = "theme"
	RewardTypeXPBoost RewardType = "xp_boost"
)

// Reward represents a reward that can be claimed with gold coins
type Reward struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Image       string     `gorm:"size:500;default:''" json:"image"`
	CoinCost    int        `gorm:"not null" json:"coin_cost"`
	Stock       int        `gorm:"default:-1" json:"stock"` // -1 = unlimited
	RewardType  RewardType `gorm:"size:50;default:'general'" json:"reward_type"`
	RewardValue string     `gorm:"size:100;default:''" json:"reward_value"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Reward) TableName() string {
	return "rewards"
}

// IsAvailable checks if the reward can be claimed
func (r *Reward) IsAvailable() bool {
	return r.IsActive && (r.Stock == -1 || r.Stock > 0)
}

// RewardClaim represents a user's claim of a reward
type RewardClaim struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	RewardID  uint      `gorm:"not null" json:"reward_id"`
	CoinSpent int       `gorm:"not null" json:"coin_spent"`
	ClaimedAt time.Time `json:"claimed_at"`

	// Relations
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Reward Reward `gorm:"foreignKey:RewardID" json:"reward,omitempty"`
}

func (RewardClaim) TableName() string {
	return "reward_claims"
}
