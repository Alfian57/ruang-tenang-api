package model

import (
	"time"

	"github.com/google/uuid"
)

// TimedChallengeStatus defines challenge states
type TimedChallengeStatus string

const (
	TimedChallengeActive    TimedChallengeStatus = "active"
	TimedChallengeCompleted TimedChallengeStatus = "completed"
	TimedChallengeExpired   TimedChallengeStatus = "expired"
)

// TimedChallengeTemplate defines a reusable quest template
type TimedChallengeTemplate struct {
	ID              int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title           string    `gorm:"size:200;not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	ChallengeType   string    `gorm:"size:50;not null" json:"challenge_type"`
	TargetValue     int       `gorm:"not null" json:"target_value"`
	DurationMinutes int       `gorm:"not null;default:60" json:"duration_minutes"`
	XPReward        int       `gorm:"not null;default:50" json:"xp_reward"`
	CoinReward      int       `gorm:"not null;default:0" json:"coin_reward"`
	Icon            string    `gorm:"size:50;not null;default:'⚡'" json:"icon"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

func (TimedChallengeTemplate) TableName() string { return "timed_challenge_templates" }

// UserTimedChallenge represents an assigned timed challenge instance
type UserTimedChallenge struct {
	ID           uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uint                    `gorm:"not null" json:"user_id"`
	TemplateID   int                     `gorm:"not null" json:"template_id"`
	CurrentValue int                     `gorm:"not null;default:0" json:"current_value"`
	Status       TimedChallengeStatus    `gorm:"size:20;not null;default:'active'" json:"status"`
	StartedAt    time.Time               `gorm:"not null" json:"started_at"`
	ExpiresAt    time.Time               `gorm:"not null" json:"expires_at"`
	CompletedAt  *time.Time              `json:"completed_at,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
	User         *User                   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Template     *TimedChallengeTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

func (UserTimedChallenge) TableName() string { return "user_timed_challenges" }

// IsExpired checks if challenge time has run out
func (c *UserTimedChallenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// RemainingSeconds returns seconds until expiry
func (c *UserTimedChallenge) RemainingSeconds() int {
	remaining := time.Until(c.ExpiresAt).Seconds()
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}
