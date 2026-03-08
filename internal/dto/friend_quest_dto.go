package dto

import (
	"time"

	"github.com/google/uuid"
)

// === FRIEND QUEST DTOs ===

// CreateFriendQuestRequest to invite a friend for a quest
type CreateFriendQuestRequest struct {
	PartnerID   uint   `json:"partner_id" binding:"required"`
	Title       string `json:"title" binding:"required,min=3,max=200"`
	Description string `json:"description" binding:"max=500"`
	QuestType   string `json:"quest_type" binding:"required,oneof=total_xp breathing journal chat mood"`
	TargetValue int    `json:"target_value" binding:"required,min=1"`
	XPReward    int    `json:"xp_reward" binding:"required,min=1"`
	CoinReward  int    `json:"coin_reward" binding:"min=0"`
	DurationHrs int    `json:"duration_hours" binding:"required,min=1,max=168"`
}

// FriendQuestResponse represents a friend quest
type FriendQuestResponse struct {
	ID                uuid.UUID     `json:"id"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	QuestType         string        `json:"quest_type"`
	TargetValue       int           `json:"target_value"`
	RequesterProgress int           `json:"requester_progress"`
	PartnerProgress   int           `json:"partner_progress"`
	TotalProgress     int           `json:"total_progress"`
	ProgressPercent   float64       `json:"progress_percent"`
	XPReward          int           `json:"xp_reward"`
	CoinReward        int           `json:"coin_reward"`
	Status            string        `json:"status"`
	Requester         QuestUserInfo `json:"requester"`
	Partner           QuestUserInfo `json:"partner"`
	StartsAt          *time.Time    `json:"starts_at,omitempty"`
	EndsAt            *time.Time    `json:"ends_at,omitempty"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
}

// QuestUserInfo is a minimal user info for quest display
type QuestUserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
}

// FriendQuestFilterRequest for filtering quests
type FriendQuestFilterRequest struct {
	Status string `form:"status"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}
