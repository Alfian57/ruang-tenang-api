package models

import (
	"time"

	"gorm.io/gorm"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportTypeArticle   ReportType = "article"
	ReportTypeForum     ReportType = "forum"
	ReportTypeForumPost ReportType = "forum_post"
	ReportTypeUser      ReportType = "user"
)

// ReportReason represents reasons for reporting
type ReportReason string

const (
	ReportReasonMisinformation ReportReason = "misinformation"
	ReportReasonHarmful        ReportReason = "harmful"
	ReportReasonHarassment     ReportReason = "harassment"
	ReportReasonSpam           ReportReason = "spam"
	ReportReasonImpersonation  ReportReason = "impersonation"
	ReportReasonOther          ReportReason = "other"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusReviewing ReportStatus = "reviewing"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

// ActionTaken represents actions taken on a report
type ActionTaken string

const (
	ActionTakenNone           ActionTaken = "none"
	ActionTakenContentRemoved ActionTaken = "content_removed"
	ActionTakenUserWarned     ActionTaken = "user_warned"
	ActionTakenUserSuspended  ActionTaken = "user_suspended"
	ActionTakenUserBanned     ActionTaken = "user_banned"
)

// UserReport represents a user report
type UserReport struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	ReporterID        uint           `gorm:"not null" json:"reporter_id"`
	ReportType        ReportType     `gorm:"size:50;not null" json:"report_type"`
	ReportedContentID *uint          `json:"reported_content_id,omitempty"`
	ReportedUserID    *uint          `json:"reported_user_id,omitempty"`
	Reason            ReportReason   `gorm:"size:100;not null" json:"reason"`
	Description       string         `gorm:"type:text" json:"description,omitempty"`
	Status            ReportStatus   `gorm:"size:50;default:'pending'" json:"status"`
	HandledByID       *uint          `json:"handled_by_id,omitempty"`
	HandledAt         *time.Time     `json:"handled_at,omitempty"`
	ActionTaken       ActionTaken    `gorm:"size:100" json:"action_taken,omitempty"`
	ModeratorNotes    string         `gorm:"type:text" json:"moderator_notes,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Reporter     *User `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	ReportedUser *User `gorm:"foreignKey:ReportedUserID" json:"reported_user,omitempty"`
	HandledBy    *User `gorm:"foreignKey:HandledByID" json:"handled_by,omitempty"`
}

func (UserReport) TableName() string {
	return "user_reports"
}
