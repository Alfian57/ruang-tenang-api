package dto

import (
	"time"

	"github.com/google/uuid"
)

// === TIMED CHALLENGE DTOs ===

// TimedChallengeTemplateResponse describes a challenge template
type TimedChallengeTemplateResponse struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	ChallengeType   string `json:"challenge_type"`
	TargetValue     int    `json:"target_value"`
	DurationMinutes int    `json:"duration_minutes"`
	XPReward        int    `json:"xp_reward"`
	CoinReward      int    `json:"coin_reward"`
	Icon            string `json:"icon"`
}

// UserTimedChallengeResponse represents an active challenge
type UserTimedChallengeResponse struct {
	ID               uuid.UUID                      `json:"id"`
	Template         TimedChallengeTemplateResponse `json:"template"`
	CurrentValue     int                            `json:"current_value"`
	TargetValue      int                            `json:"target_value"`
	ProgressPercent  float64                        `json:"progress_percent"`
	Status           string                         `json:"status"`
	StartedAt        time.Time                      `json:"started_at"`
	ExpiresAt        time.Time                      `json:"expires_at"`
	RemainingSeconds int                            `json:"remaining_seconds"`
	CompletedAt      *time.Time                     `json:"completed_at,omitempty"`
}

// StartTimedChallengeRequest to start a new timed challenge
type StartTimedChallengeRequest struct {
	TemplateID int `json:"template_id" binding:"required"`
}

// TimedChallengeFilterRequest for filtering challenges
type TimedChallengeFilterRequest struct {
	Status string `form:"status"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}
