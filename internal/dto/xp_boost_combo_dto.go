package dto

import (
	"time"

	"github.com/google/uuid"
)

// === XP BOOST DTOs ===

// XPBoostResponse represents an active boost
type XPBoostResponse struct {
	ID               uuid.UUID `json:"id"`
	Multiplier       float64   `json:"multiplier"`
	TriggerType      string    `json:"trigger_type"`
	StartedAt        time.Time `json:"started_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	RemainingSeconds int       `json:"remaining_seconds"`
}

// === COMBO DTOs ===

// ComboStatusResponse shows current combo state
type ComboStatusResponse struct {
	ComboCount     int        `json:"combo_count"`
	Multiplier     float64    `json:"multiplier"`
	NextMultiplier float64    `json:"next_multiplier"`
	LastActivity   string     `json:"last_activity,omitempty"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	ExpiresInSecs  int        `json:"expires_in_seconds"`
}
