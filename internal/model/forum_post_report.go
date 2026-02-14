package model

import "time"

// ForumPostReportReason represents the reason for reporting a post
// Using different name to avoid conflict with UserReport's ReportReason
type ForumPostReportReason string

const (
	PostReportReasonSpam           ForumPostReportReason = "spam"
	PostReportReasonHarassment     ForumPostReportReason = "harassment"
	PostReportReasonMisinformation ForumPostReportReason = "misinformation"
	PostReportReasonSelfHarm       ForumPostReportReason = "self_harm"
	PostReportReasonOffTopic       ForumPostReportReason = "off_topic"
	PostReportReasonRude           ForumPostReportReason = "rude"
	PostReportReasonOther          ForumPostReportReason = "other"
)

// ForumPostReportStatus represents the status of a post report
type ForumPostReportStatus string

const (
	PostReportStatusPending   ForumPostReportStatus = "pending"
	PostReportStatusReviewed  ForumPostReportStatus = "reviewed"
	PostReportStatusDismissed ForumPostReportStatus = "dismissed"
	PostReportStatusActioned  ForumPostReportStatus = "actioned"
)

// ForumPostReport represents a user's report on a forum post
type ForumPostReport struct {
	ID             uint                  `gorm:"primaryKey" json:"id"`
	PostID         uint                  `gorm:"not null" json:"post_id"`
	ReporterID     uint                  `gorm:"not null" json:"reporter_id"`
	Reason         ForumPostReportReason `gorm:"type:varchar(50);not null" json:"reason"`
	Description    string                `gorm:"type:text" json:"description,omitempty"`
	Status         ForumPostReportStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ReviewedBy     *uint                 `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time            `json:"reviewed_at,omitempty"`
	ModeratorNotes string                `gorm:"type:text" json:"moderator_notes,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`

	// Relations
	Post     ForumPost `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Reporter User      `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	Reviewer *User     `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (ForumPostReport) TableName() string {
	return "forum_post_reports"
}

// IsPending returns true if report is pending review
func (r *ForumPostReport) IsPending() bool {
	return r.Status == PostReportStatusPending
}

// ValidPostReportReasons returns all valid report reasons for posts
func ValidPostReportReasons() []ForumPostReportReason {
	return []ForumPostReportReason{
		PostReportReasonSpam,
		PostReportReasonHarassment,
		PostReportReasonMisinformation,
		PostReportReasonSelfHarm,
		PostReportReasonOffTopic,
		PostReportReasonRude,
		PostReportReasonOther,
	}
}

// IsValidPostReportReason checks if a reason is valid for post reports
func IsValidPostReportReason(reason string) bool {
	for _, r := range ValidPostReportReasons() {
		if string(r) == reason {
			return true
		}
	}
	return false
}
