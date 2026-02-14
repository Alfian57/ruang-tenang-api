package model

import (
	"time"
)

// ModeratorActionType represents types of moderator actions
type ModeratorActionType string

const (
	ModeratorActionArticleApproved      ModeratorActionType = "article_approved"
	ModeratorActionArticleRejected      ModeratorActionType = "article_rejected"
	ModeratorActionArticleRequestEdit   ModeratorActionType = "article_request_edit"
	ModeratorActionContentRemoved       ModeratorActionType = "content_removed"
	ModeratorActionContentRestored      ModeratorActionType = "content_restored"
	ModeratorActionTriggerWarningAdded  ModeratorActionType = "trigger_warning_added"
	ModeratorActionTriggerWarningRemove ModeratorActionType = "trigger_warning_removed"
	ModeratorActionUserWarned           ModeratorActionType = "user_warned"
	ModeratorActionUserSuspended        ModeratorActionType = "user_suspended"
	ModeratorActionUserBanned           ModeratorActionType = "user_banned"
	ModeratorActionUserUnbanned         ModeratorActionType = "user_unbanned"
	ModeratorActionReportDismissed      ModeratorActionType = "report_dismissed"
	ModeratorActionReportResolved       ModeratorActionType = "report_resolved"
	ModeratorActionFlagResolved         ModeratorActionType = "flag_resolved"
)

// ModeratorAction represents an audit log of moderator actions
type ModeratorAction struct {
	ID            uint                `gorm:"primaryKey" json:"id"`
	ModeratorID   uint                `gorm:"not null" json:"moderator_id"`
	ActionType    ModeratorActionType `gorm:"size:100;not null" json:"action_type"`
	TargetType    string              `gorm:"size:50;not null" json:"target_type"` // 'article', 'forum', 'forum_post', 'user', 'report'
	TargetID      uint                `gorm:"not null" json:"target_id"`
	PreviousState string              `gorm:"type:text" json:"previous_state,omitempty"`
	NewState      string              `gorm:"type:text" json:"new_state,omitempty"`
	Reason        string              `gorm:"type:text" json:"reason,omitempty"`
	Notes         string              `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`

	// Relations
	Moderator *User `gorm:"foreignKey:ModeratorID" json:"moderator,omitempty"`
}

func (ModeratorAction) TableName() string {
	return "moderator_actions"
}
