package model

import (
	"time"

	"github.com/google/uuid"
)

// ChestRarity defines chest rarity levels
type ChestRarity string

const (
	ChestCommon    ChestRarity = "common"
	ChestRare      ChestRarity = "rare"
	ChestEpic      ChestRarity = "epic"
	ChestLegendary ChestRarity = "legendary"
)

// ChestRewardType defines what a chest can contain
type ChestRewardType string

const (
	ChestRewardXP           ChestRewardType = "xp"
	ChestRewardCoins        ChestRewardType = "coins"
	ChestRewardStreakFreeze ChestRewardType = "streak_freeze"
	ChestRewardXPBoost      ChestRewardType = "xp_boost"
)

// UserChest represents a chest earned by a user
type UserChest struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             uint            `gorm:"not null" json:"user_id"`
	Rarity             ChestRarity     `gorm:"size:20;not null;default:'common'" json:"rarity"`
	IsOpened           bool            `gorm:"default:false" json:"is_opened"`
	RewardType         ChestRewardType `gorm:"size:50" json:"reward_type,omitempty"`
	RewardValue        int             `gorm:"default:0" json:"reward_value"`
	RewardLabel        string          `gorm:"size:200" json:"reward_label,omitempty"`
	TriggerType        string          `gorm:"size:50;not null;default:'milestone'" json:"trigger_type"`
	TriggerDescription string          `gorm:"size:200" json:"trigger_description,omitempty"`
	OpenedAt           *time.Time      `json:"opened_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	User               *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserChest) TableName() string { return "user_chests" }

// RarityIcon returns an emoji for the rarity
func (c *UserChest) RarityIcon() string {
	switch c.Rarity {
	case ChestLegendary:
		return "🏆"
	case ChestEpic:
		return "💎"
	case ChestRare:
		return "✨"
	default:
		return "📦"
	}
}
