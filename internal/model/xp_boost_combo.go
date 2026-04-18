package model

import (
	"time"

	"github.com/google/uuid"
)

// XPBoostTrigger defines what triggered the boost
type XPBoostTrigger string

const (
	BoostTriggerActivityChain XPBoostTrigger = "activity_chain"
	BoostTriggerChest         XPBoostTrigger = "chest"
	BoostTriggerSpin          XPBoostTrigger = "spin"
	BoostTriggerReward        XPBoostTrigger = "reward"
)

// XPBoost represents an active double XP timer
type XPBoost struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	Multiplier  float64        `gorm:"type:decimal(3,1);not null;default:2.0" json:"multiplier"`
	TriggerType XPBoostTrigger `gorm:"size:50;not null;default:'activity_chain'" json:"trigger_type"`
	StartedAt   time.Time      `gorm:"not null" json:"started_at"`
	ExpiresAt   time.Time      `gorm:"not null" json:"expires_at"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (XPBoost) TableName() string { return "xp_boosts" }

// IsExpired checks if the boost has expired
func (b *XPBoost) IsExpired() bool {
	return time.Now().After(b.ExpiresAt)
}

// UserCombo tracks the current combo chain state
type UserCombo struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uint       `gorm:"not null;uniqueIndex" json:"user_id"`
	ComboCount       int        `gorm:"not null;default:0" json:"combo_count"`
	Multiplier       float64    `gorm:"type:decimal(3,1);not null;default:1.0" json:"multiplier"`
	LastActivityType string     `gorm:"size:50" json:"last_activity_type"`
	LastActivityAt   *time.Time `json:"last_activity_at,omitempty"`
	SessionStartedAt *time.Time `json:"session_started_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	User             *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserCombo) TableName() string { return "user_combos" }

// ComboMultiplier returns the multiplier for a given combo count
func ComboMultiplier(count int) float64 {
	switch {
	case count >= 4:
		return 3.0
	case count == 3:
		return 2.5
	case count == 2:
		return 2.0
	case count == 1:
		return 1.5
	default:
		return 1.0
	}
}
