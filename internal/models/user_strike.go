package models

import (
	"time"
)

// StrikeSeverity represents strike severity levels
type StrikeSeverity string

const (
	StrikeSeverityWarning StrikeSeverity = "warning"
	StrikeSeverityMinor   StrikeSeverity = "minor"
	StrikeSeverityMajor   StrikeSeverity = "major"
)

// UserStrike represents a violation strike against a user
type UserStrike struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null" json:"user_id"`
	ReportID   *uint          `json:"report_id,omitempty"`
	Reason     string         `gorm:"size:255;not null" json:"reason"`
	Severity   StrikeSeverity `gorm:"size:20;default:'warning'" json:"severity"`
	IssuedByID uint           `gorm:"not null" json:"issued_by_id"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	Notes      string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`

	// Relations
	User     *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Report   *UserReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	IssuedBy *User       `gorm:"foreignKey:IssuedByID" json:"issued_by,omitempty"`
}

func (UserStrike) TableName() string {
	return "user_strikes"
}
