package model

import (
	"time"

	"github.com/google/uuid"
)

// FeatureDefinition defines an unlockable feature
type FeatureDefinition struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FeatureKey    string    `gorm:"size:100;not null;uniqueIndex" json:"feature_key"`
	FeatureName   string    `gorm:"size:200;not null" json:"feature_name"`
	Description   string    `gorm:"type:text" json:"description"`
	Icon          string    `gorm:"size:50" json:"icon"`
	RequiredLevel int       `gorm:"not null;default:1" json:"required_level"`
	Category      string    `gorm:"size:50;default:'general'" json:"category"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	DisplayOrder  int       `gorm:"default:0" json:"display_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (FeatureDefinition) TableName() string {
	return "feature_definitions"
}

// UserFeatureUnlock tracks which features a user has unlocked
type UserFeatureUnlock struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	FeatureID  uuid.UUID `gorm:"type:uuid;not null" json:"feature_id"`
	UnlockedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"unlocked_at"`

	// Relations
	User    *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Feature *FeatureDefinition `gorm:"foreignKey:FeatureID" json:"feature,omitempty"`
}

func (UserFeatureUnlock) TableName() string {
	return "user_feature_unlocks"
}
