package model

import (
	"time"
)

type AppealStatus string

const (
	AppealStatusPending  AppealStatus = "pending"
	AppealStatusApproved AppealStatus = "approved"
	AppealStatusRejected AppealStatus = "rejected"
)

type Appeal struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	UserID        uint         `gorm:"not null" json:"user_id"`
	Reason        string       `gorm:"type:text;not null" json:"reason"`
	Evidence      string       `gorm:"type:text" json:"evidence,omitempty"`
	Status        AppealStatus `gorm:"size:20;default:'pending'" json:"status"`
	ReviewerNotes string       `gorm:"type:text" json:"reviewer_notes,omitempty"`
	ReviewerID    *uint        `json:"reviewer_id,omitempty"`
	Reviewer      *User        `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user"`
}

func (Appeal) TableName() string {
	return "appeals"
}
