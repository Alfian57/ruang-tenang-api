package model

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/pkg/slug"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

// ContentWarningPreference represents user's preference for content warnings
type ContentWarningPreference string

const (
	ContentWarningShow        ContentWarningPreference = "show"
	ContentWarningHideAll     ContentWarningPreference = "hide_all"
	ContentWarningAskEachTime ContentWarningPreference = "ask_each_time"
)

type User struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:255;not null" json:"name"`
	Username         string    `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Email            string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password         string    `gorm:"size:255;not null" json:"-"`
	Role             UserRole  `gorm:"type:varchar(20);default:'member'" json:"role"`
	Exp              int64     `gorm:"default:0" json:"exp"`
	GoldCoins        int64     `gorm:"default:0" json:"gold_coins"`
	Avatar           string    `gorm:"size:255;default:''" json:"avatar"`
	IsBlocked        bool      `gorm:"default:false" json:"is_blocked"`
	IsForumBlocked   bool      `gorm:"default:false" json:"is_forum_blocked"`
	ResetToken       string    `gorm:"size:255" json:"-"`
	ResetTokenExpiry time.Time `json:"-"`

	// Moderation fields
	SuspensionEnd            *time.Time               `json:"suspension_end,omitempty"`
	SuspensionReason         string                   `gorm:"type:text" json:"suspension_reason,omitempty"`
	IsBanned                 bool                     `gorm:"default:false" json:"is_banned"`
	BanReason                string                   `gorm:"type:text" json:"ban_reason,omitempty"`
	HasAcceptedAIDisclaimer  bool                     `gorm:"column:has_accepted_ai_disclaimer;default:false" json:"has_accepted_ai_disclaimer"`
	ContentWarningPreference ContentWarningPreference `gorm:"size:20;default:'show'" json:"content_warning_preference"`

	// Profile customization
	ProfileTheme      string `gorm:"size:50;default:'default'" json:"profile_theme"`
	ProfileBanner     string `gorm:"size:500" json:"profile_banner"`
	AvatarBorderColor string `gorm:"size:20" json:"avatar_border_color"`
	Tagline           string `gorm:"size:200" json:"tagline"`
	Bio               string `gorm:"type:text" json:"bio"`

	// Gamification stats
	CurrentStreak         int        `gorm:"default:0" json:"current_streak"`
	LongestStreak         int        `gorm:"default:0" json:"longest_streak"`
	LastActivityDate      *time.Time `json:"last_activity_date,omitempty"`
	TotalActivities       int        `gorm:"default:0" json:"total_activities"`
	LastLoginDate         *time.Time `gorm:"type:date" json:"last_login_date,omitempty"`
	LoginStreak           int        `gorm:"default:0" json:"login_streak"`
	StreakFreezeAvailable bool       `gorm:"default:true" json:"streak_freeze_available"`
	StreakFreezeUsedAt    *time.Time `gorm:"type:date" json:"streak_freeze_used_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	ChatSessions []ChatSession `gorm:"foreignKey:UserID" json:"chat_sessions,omitempty"`
	UserMoods    []UserMood    `gorm:"foreignKey:UserID" json:"user_moods,omitempty"`
	Badges       []UserBadge   `gorm:"foreignKey:UserID" json:"badges,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate generates a username from Name if not already set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Username == "" {
		u.Username = slug.GenerateUnique(u.Name)
	}
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsMember() bool {
	return u.Role == RoleMember
}

// CanModerate returns true if user has moderation privileges (admin only)
func (u *User) CanModerate() bool {
	return u.Role == RoleAdmin
}

// IsSuspended returns true if user is currently suspended
func (u *User) IsSuspended() bool {
	if u.SuspensionEnd == nil {
		return false
	}
	return time.Now().Before(*u.SuspensionEnd)
}

// CanAccess returns true if user is not banned, blocked, or suspended
func (u *User) CanAccess() bool {
	return !u.IsBanned && !u.IsBlocked && !u.IsSuspended()
}
