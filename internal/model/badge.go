package model

import (
	"time"

	"github.com/google/uuid"
)

// BadgeRequirementType defines how a badge is earned
type BadgeRequirementType string

const (
	BadgeRequirementStreak        BadgeRequirementType = "streak"
	BadgeRequirementActivityCount BadgeRequirementType = "activity_count"
	BadgeRequirementLevel         BadgeRequirementType = "level"
	BadgeRequirementManual        BadgeRequirementType = "manual"
	BadgeRequirementStory         BadgeRequirementType = "story"
	BadgeRequirementXP            BadgeRequirementType = "xp"
)

// BadgeDefinition defines an achievement badge
type BadgeDefinition struct {
	ID               uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BadgeKey         string               `gorm:"size:100;not null;uniqueIndex" json:"badge_key"`
	BadgeName        string               `gorm:"size:200;not null" json:"badge_name"`
	Description      string               `gorm:"type:text" json:"description"`
	Icon             string               `gorm:"size:50" json:"icon"`
	Category         string               `gorm:"size:50;default:'general'" json:"category"`
	RequirementType  BadgeRequirementType `gorm:"size:50;not null" json:"requirement_type"`
	RequirementValue int                  `gorm:"default:0" json:"requirement_value"`
	IsActive         bool                 `gorm:"default:true" json:"is_active"`
	DisplayOrder     int                  `gorm:"default:0" json:"display_order"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

func (BadgeDefinition) TableName() string {
	return "badge_definitions"
}

// UserBadge represents a badge earned by a user
type UserBadge struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	BadgeID     uuid.UUID `gorm:"type:uuid;not null" json:"badge_id"`
	EarnedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"earned_at"`
	IsShowcased bool      `gorm:"default:false" json:"is_showcased"`

	// Relations
	User  *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Badge *BadgeDefinition `gorm:"foreignKey:BadgeID" json:"badge,omitempty"`
}

func (UserBadge) TableName() string {
	return "user_badges"
}
