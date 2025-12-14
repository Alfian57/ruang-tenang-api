package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      UserRole       `gorm:"type:varchar(20);default:'member'" json:"role"`
	IsBlocked bool           `gorm:"default:false" json:"is_blocked"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	ChatSessions []ChatSession `gorm:"foreignKey:UserID" json:"chat_sessions,omitempty"`
	UserMoods    []UserMood    `gorm:"foreignKey:UserID" json:"user_moods,omitempty"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsMember() bool {
	return u.Role == RoleMember
}
