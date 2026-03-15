package model

import (
	"time"
)

type LevelConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Level           int       `gorm:"uniqueIndex;not null" json:"level"`
	MinExp          int       `gorm:"not null;default:0" json:"min_exp"`
	BadgeName       string    `gorm:"size:100;not null" json:"badge_name"`
	BadgeIcon       string    `gorm:"size:255;not null" json:"badge_icon"`
	TaskDescription string    `gorm:"type:text" json:"task_description"`
	TierName        string    `gorm:"size:50" json:"tier_name"`
	TierColor       string    `gorm:"size:20" json:"tier_color"`
	Description     string    `gorm:"type:text" json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (LevelConfig) TableName() string {
	return "level_configs"
}
