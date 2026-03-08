package model

import (
	"time"

	"github.com/google/uuid"
)

// SpinRewardType defines what a spin can give
type SpinRewardType string

const (
	SpinRewardXP           SpinRewardType = "xp"
	SpinRewardCoins        SpinRewardType = "coins"
	SpinRewardStreakFreeze SpinRewardType = "streak_freeze"
	SpinRewardXPBoost      SpinRewardType = "xp_boost"
	SpinRewardNothing      SpinRewardType = "nothing"
)

// SpinReward defines a slot on the roulette wheel
type SpinReward struct {
	ID          int            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Icon        string         `gorm:"size:50;not null" json:"icon"`
	RewardType  SpinRewardType `gorm:"size:50;not null" json:"reward_type"`
	RewardValue int            `gorm:"not null;default:0" json:"reward_value"`
	Weight      int            `gorm:"not null;default:100" json:"weight"`
	Rarity      string         `gorm:"size:20;not null;default:'common'" json:"rarity"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
}

func (SpinReward) TableName() string { return "spin_rewards" }

// UserSpin records a user's daily spin result
type UserSpin struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uint        `gorm:"not null" json:"user_id"`
	RewardID  int         `gorm:"not null" json:"reward_id"`
	SpinDate  time.Time   `gorm:"type:date;not null" json:"spin_date"`
	CreatedAt time.Time   `json:"created_at"`
	User      *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Reward    *SpinReward `gorm:"foreignKey:RewardID" json:"reward,omitempty"`
}

func (UserSpin) TableName() string { return "user_spins" }
