package model

import (
	"time"

	"github.com/google/uuid"
)

// StreakSociety defines an exclusive club tier
type StreakSociety struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Icon          string    `gorm:"size:50;not null" json:"icon"`
	MinStreak     int       `gorm:"not null;uniqueIndex" json:"min_streak"`
	BorderColor   string    `gorm:"size:20;not null;default:'#888888'" json:"border_color"`
	BadgeGlow     bool      `gorm:"default:false" json:"badge_glow"`
	ExclusiveChat bool      `gorm:"default:false" json:"exclusive_chat"`
	DisplayOrder  int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt     time.Time `json:"created_at"`
}

func (StreakSociety) TableName() string { return "streak_societies" }

// UserSocietyMembership tracks which society a user belongs to
type UserSocietyMembership struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	SocietyID int            `gorm:"not null" json:"society_id"`
	JoinedAt  time.Time      `json:"joined_at"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Society   *StreakSociety `gorm:"foreignKey:SocietyID" json:"society,omitempty"`
}

func (UserSocietyMembership) TableName() string { return "user_society_memberships" }
