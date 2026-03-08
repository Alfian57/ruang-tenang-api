package model

import (
	"time"

	"github.com/google/uuid"
)

// MapUnlockType defines how a region or landmark is unlocked
type MapUnlockType string

const (
	MapUnlockLevel         MapUnlockType = "level"
	MapUnlockActivityCount MapUnlockType = "activity_count"
	MapUnlockStreak        MapUnlockType = "streak"
	MapUnlockBadge         MapUnlockType = "badge"
	MapUnlockXP            MapUnlockType = "xp"
)

// MapRegion represents an area on the progress map
type MapRegion struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RegionKey      string        `gorm:"size:100;not null;uniqueIndex" json:"region_key"`
	Name           string        `gorm:"size:200;not null" json:"name"`
	Description    string        `gorm:"type:text" json:"description"`
	Icon           string        `gorm:"size:50" json:"icon"`
	Image          string        `gorm:"size:500" json:"image"`
	UnlockType     MapUnlockType `gorm:"size:50;not null;default:'level'" json:"unlock_type"`
	UnlockValue    int           `gorm:"not null;default:1" json:"unlock_value"`
	PositionX      int           `gorm:"not null;default:0" json:"position_x"`
	PositionY      int           `gorm:"not null;default:0" json:"position_y"`
	DisplayOrder   int           `gorm:"not null;default:0" json:"display_order"`
	ParentRegionID *uuid.UUID    `gorm:"type:uuid" json:"parent_region_id,omitempty"`
	IsActive       bool          `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`

	// Relations
	ParentRegion *MapRegion    `gorm:"foreignKey:ParentRegionID" json:"parent_region,omitempty"`
	Landmarks    []MapLandmark `gorm:"foreignKey:RegionID" json:"landmarks,omitempty"`
}

func (MapRegion) TableName() string {
	return "map_regions"
}

// MapLandmark represents a point of interest within a region
type MapLandmark struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RegionID       uuid.UUID     `gorm:"type:uuid;not null" json:"region_id"`
	LandmarkKey    string        `gorm:"size:100;not null;uniqueIndex" json:"landmark_key"`
	Name           string        `gorm:"size:200;not null" json:"name"`
	Description    string        `gorm:"type:text" json:"description"`
	Icon           string        `gorm:"size:50" json:"icon"`
	UnlockType     MapUnlockType `gorm:"size:50;not null;default:'activity_count'" json:"unlock_type"`
	UnlockActivity string        `gorm:"size:50" json:"unlock_activity"`
	UnlockValue    int           `gorm:"not null;default:1" json:"unlock_value"`
	PositionX      int           `gorm:"not null;default:0" json:"position_x"`
	PositionY      int           `gorm:"not null;default:0" json:"position_y"`
	XPReward       int           `gorm:"not null;default:0" json:"xp_reward"`
	CoinReward     int           `gorm:"not null;default:0" json:"coin_reward"`
	DisplayOrder   int           `gorm:"not null;default:0" json:"display_order"`
	IsActive       bool          `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time     `json:"created_at"`

	// Relations
	Region *MapRegion `gorm:"foreignKey:RegionID" json:"region,omitempty"`
}

func (MapLandmark) TableName() string {
	return "map_landmarks"
}

// UserMapProgress tracks a user's region unlock progress
type UserMapProgress struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uint       `gorm:"not null" json:"user_id"`
	RegionID   uuid.UUID  `gorm:"type:uuid;not null" json:"region_id"`
	IsUnlocked bool       `gorm:"default:false" json:"is_unlocked"`
	UnlockedAt *time.Time `json:"unlocked_at,omitempty"`

	// Relations
	User   *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Region *MapRegion `gorm:"foreignKey:RegionID" json:"region,omitempty"`
}

func (UserMapProgress) TableName() string {
	return "user_map_progress"
}

// UserLandmarkProgress tracks a user's landmark unlock progress
type UserLandmarkProgress struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uint       `gorm:"not null" json:"user_id"`
	LandmarkID    uuid.UUID  `gorm:"type:uuid;not null" json:"landmark_id"`
	IsUnlocked    bool       `gorm:"default:false" json:"is_unlocked"`
	CurrentValue  int        `gorm:"not null;default:0" json:"current_value"`
	UnlockedAt    *time.Time `json:"unlocked_at,omitempty"`
	RewardClaimed bool       `gorm:"default:false" json:"reward_claimed"`

	// Relations
	User     *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Landmark *MapLandmark `gorm:"foreignKey:LandmarkID" json:"landmark,omitempty"`
}

func (UserLandmarkProgress) TableName() string {
	return "user_landmark_progress"
}
