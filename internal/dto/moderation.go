package dto

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

// ========================
// Content Moderation DTOs
// ========================

// AI Moderation Result
type AIModerationResult struct {
	Status       model.ArticleModerationStatus `json:"status"`     // "approved", "flagged", "rejected"
	Confidence   float64                        `json:"confidence"` // 0-100
	Reasons      []string                       `json:"reasons"`    // List of flag reasons
	FlagCategory string                         `json:"flag_category,omitempty"`
	Severity     string                         `json:"severity,omitempty"`
	Suggestions  string                         `json:"suggestions,omitempty"` // AI suggestions for author
}

// Moderator Article Actions
type ModerateArticleRequest struct {
	Action          string   `json:"action" binding:"required,oneof=approve reject request_edit"`
	Notes           string   `json:"notes"`
	TriggerWarnings []string `json:"trigger_warnings,omitempty"`
}

type ModerationQueueParams struct {
	Status   string `form:"status"`   // flagged, pending, all
	Severity string `form:"severity"` // low, medium, high
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}

type ModerationQueueItem struct {
	ID               uint                           `json:"id"`
	Title            string                         `json:"title"`
	Excerpt          string                         `json:"excerpt"`
	AuthorID         uint                           `json:"author_id"`
	AuthorName       string                         `json:"author_name"`
	ModerationStatus model.ArticleModerationStatus `json:"moderation_status"`
	Severity         string                         `json:"severity,omitempty"`
	FlagReasons      []string                       `json:"flag_reasons,omitempty"`
	CreatedAt        time.Time                      `json:"created_at"`
	UpdatedAt        time.Time                      `json:"updated_at"`
}

// Content Flag DTOs
type ContentFlagDTO struct {
	ID              uint       `json:"id"`
	ContentType     string     `json:"content_type"`
	ContentID       uint       `json:"content_id"`
	FlagType        string     `json:"flag_type"`
	FlagCategory    string     `json:"flag_category"`
	Severity        string     `json:"severity"`
	AIConfidence    *float64   `json:"ai_confidence,omitempty"`
	AIReason        string     `json:"ai_reason,omitempty"`
	FlaggedByID     *uint      `json:"flagged_by_id,omitempty"`
	FlaggedByName   string     `json:"flagged_by_name,omitempty"`
	IsResolved      bool       `json:"is_resolved"`
	ResolvedByID    *uint      `json:"resolved_by_id,omitempty"`
	ResolvedByName  string     `json:"resolved_by_name,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes string     `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AddContentFlagRequest struct {
	ContentType  string `json:"content_type" binding:"required,oneof=article forum forum_post"`
	ContentID    uint   `json:"content_id" binding:"required"`
	FlagCategory string `json:"flag_category" binding:"required"`
	Severity     string `json:"severity,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type ResolveFlagRequest struct {
	Resolution string `json:"resolution" binding:"required"` // Notes about resolution
}

// ========================
// Trigger Warning DTOs
// ========================

type TriggerWarningRequest struct {
	ContentType     string   `json:"content_type" binding:"required,oneof=article forum"`
	ContentID       uint     `json:"content_id" binding:"required"`
	TriggerWarnings []string `json:"trigger_warnings" binding:"required"`
}

type ContentWithWarning struct {
	ContentType     string   `json:"content_type"`
	ContentID       uint     `json:"content_id"`
	Title           string   `json:"title"`
	TriggerWarnings []string `json:"trigger_warnings"`
	HasWarning      bool     `json:"has_warning"`
}

// ========================
// Report DTOs
// ========================

type CreateReportRequest struct {
	ReportType  string `json:"report_type" binding:"required,oneof=article forum forum_post user"`
	ContentID   *uint  `json:"content_id,omitempty"`
	UserID      *uint  `json:"user_id,omitempty"` // For user reports
	Reason      string `json:"reason" binding:"required,oneof=misinformation harmful harassment spam impersonation other"`
	Description string `json:"description,omitempty"`
}

type ReportDTO struct {
	ID                uint       `json:"id"`
	ReporterID        uint       `json:"reporter_id"`
	ReporterName      string     `json:"reporter_name"`
	ReportType        string     `json:"report_type"`
	ReportedContentID *uint      `json:"reported_content_id,omitempty"`
	ReportedUserID    *uint      `json:"reported_user_id,omitempty"`
	ReportedUserName  string     `json:"reported_user_name,omitempty"`
	Reason            string     `json:"reason"`
	Description       string     `json:"description,omitempty"`
	Status            string     `json:"status"`
	HandledByID       *uint      `json:"handled_by_id,omitempty"`
	HandledByName     string     `json:"handled_by_name,omitempty"`
	HandledAt         *time.Time `json:"handled_at,omitempty"`
	ActionTaken       string     `json:"action_taken,omitempty"`
	ModeratorNotes    string     `json:"moderator_notes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	// Additional content info for moderators
	ContentPreview string `json:"content_preview,omitempty"`
	ContentTitle   string `json:"content_title,omitempty"`
}

type HandleReportRequest struct {
	Action   string `json:"action" binding:"required,oneof=dismiss warn remove_content suspend ban"`
	Notes    string `json:"notes,omitempty"`
	Duration int    `json:"duration,omitempty"` // Days for suspension (1, 7, 30)
}

type ReportQueryParams struct {
	Status     string `form:"status"`      // pending, reviewing, resolved, dismissed
	ReportType string `form:"report_type"` // article, forum, forum_post, user
	Reason     string `form:"reason"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
}

// ========================
// Block DTOs
// ========================

type BlockUserRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

type BlockDTO struct {
	ID            uint      `json:"id"`
	BlockerID     uint      `json:"blocker_id"`
	BlockedID     uint      `json:"blocked_id"`
	BlockedName   string    `json:"blocked_name"`
	BlockedAvatar string    `json:"blocked_avatar,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type BlockListResponse struct {
	Blocks     []BlockDTO `json:"blocks"`
	TotalCount int64      `json:"total_count"`
}

// ========================
// Strike DTOs
// ========================

type CreateStrikeRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	ReportID *uint  `json:"report_id,omitempty"`
	Reason   string `json:"reason" binding:"required"`
	Severity string `json:"severity" binding:"required,oneof=warning minor major"`
	Notes    string `json:"notes,omitempty"`
}

type StrikeDTO struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	UserName     string     `json:"user_name"`
	ReportID     *uint      `json:"report_id,omitempty"`
	Reason       string     `json:"reason"`
	Severity     string     `json:"severity"`
	IssuedByID   uint       `json:"issued_by_id"`
	IssuedByName string     `json:"issued_by_name"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     bool       `json:"is_active"`
	Notes        string     `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ========================
// Moderator Action DTOs
// ========================

type ModeratorActionDTO struct {
	ID            uint      `json:"id"`
	ModeratorID   uint      `json:"moderator_id"`
	ModeratorName string    `json:"moderator_name"`
	ActionType    string    `json:"action_type"`
	TargetType    string    `json:"target_type"`
	TargetID      uint      `json:"target_id"`
	PreviousState string    `json:"previous_state,omitempty"`
	NewState      string    `json:"new_state,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type ModeratorActionQueryParams struct {
	ModeratorID uint   `form:"moderator_id"`
	ActionType  string `form:"action_type"`
	TargetType  string `form:"target_type"`
	Page        int    `form:"page,default=1"`
	Limit       int    `form:"limit,default=20"`
}

// ========================
// User Suspension DTOs
// ========================

type SuspendUserRequest struct {
	Duration int    `json:"duration" binding:"required,oneof=1 7 30"` // Days
	Reason   string `json:"reason" binding:"required"`
}

type BanUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ========================
// Crisis Detection DTOs
// ========================

type CrisisKeywordDTO struct {
	ID       uint   `json:"id"`
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	Language string `json:"language"`
	IsActive bool   `json:"is_active"`
	Notes    string `json:"notes,omitempty"`
}

type CreateCrisisKeywordRequest struct {
	Keyword  string `json:"keyword" binding:"required"`
	Category string `json:"category" binding:"required,oneof=self_harm suicide severe_depression emergency"`
	Severity string `json:"severity" binding:"required,oneof=medium high critical"`
	Language string `json:"language,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type CrisisDetectionResponse struct {
	IsCrisis       bool     `json:"is_crisis"`
	Keywords       []string `json:"keywords,omitempty"`
	Category       string   `json:"category,omitempty"`
	Severity       string   `json:"severity,omitempty"`
	HelplineNumber string   `json:"helpline_number,omitempty"`
	HelplineInfo   string   `json:"helpline_info,omitempty"`
}

// ========================
// AI Disclaimer DTOs
// ========================

type AcceptDisclaimerRequest struct {
	Accepted bool `json:"accepted" binding:"required"`
}

type UpdateContentWarningPreferenceRequest struct {
	Preference string `json:"preference" binding:"required,oneof=show hide_all ask_each_time"`
}

// ========================
// Moderation Dashboard Stats
// ========================

type ModerationStatsDTO struct {
	PendingArticles      int64 `json:"pending_articles"`
	FlaggedArticles      int64 `json:"flagged_articles"`
	PendingReports       int64 `json:"pending_reports"`
	ResolvedReportsToday int64 `json:"resolved_reports_today"`
	ActiveStrikes        int64 `json:"active_strikes"`
	SuspendedUsers       int64 `json:"suspended_users"`
	BannedUsers          int64 `json:"banned_users"`
}
