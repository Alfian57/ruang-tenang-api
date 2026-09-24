package model

import "time"

type PhoneVerification struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"not null;uniqueIndex"`
	ChallengeHash string    `gorm:"size:64;not null;uniqueIndex"`
	CodeHash      string    `gorm:"size:64;not null"`
	Attempts      int       `gorm:"not null"`
	RememberMe    bool      `gorm:"not null"`
	ExpiresAt     time.Time `gorm:"not null"`
	CreatedAt     time.Time
}

func (PhoneVerification) TableName() string { return "user_phone_verifications" }
