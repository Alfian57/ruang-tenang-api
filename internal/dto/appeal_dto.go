package dto

import "time"

type CreateAppealRequest struct {
	Reason   string `json:"reason" binding:"required,min=10,max=500"`
	Evidence string `json:"evidence" binding:"omitempty,max=1000"`
}

type ReviewAppealRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Notes  string `json:"notes" binding:"omitempty"`
}

type AppealDTO struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	UserEmail     string    `json:"user_email,omitempty"`
	UserName      string    `json:"user_name,omitempty"`
	Reason        string    `json:"reason"`
	Evidence      string    `json:"evidence,omitempty"`
	Status        string    `json:"status"`
	ReviewerNotes string    `json:"reviewer_notes,omitempty"`
	ReviewerID    *uint     `json:"reviewer_id,omitempty"`
	ReviewerName  string    `json:"reviewer_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
