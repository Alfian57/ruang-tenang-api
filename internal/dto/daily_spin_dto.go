package dto

import (
	"time"
)

// === DAILY SPIN DTOs ===

// SpinRewardSlotResponse describes a slot on the wheel
type SpinRewardSlotResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	RewardType  string `json:"reward_type"`
	RewardValue int    `json:"reward_value"`
	Rarity      string `json:"rarity"`
}

// SpinWheelResponse shows the wheel + user status
type SpinWheelResponse struct {
	Slots        []SpinRewardSlotResponse `json:"slots"`
	HasSpunToday bool                     `json:"has_spun_today"`
	LastSpinAt   *time.Time               `json:"last_spin_at,omitempty"`
}

// SpinResultResponse returns the spin outcome
type SpinResultResponse struct {
	SlotIndex   int    `json:"slot_index"`
	RewardName  string `json:"reward_name"`
	RewardIcon  string `json:"reward_icon"`
	RewardType  string `json:"reward_type"`
	RewardValue int    `json:"reward_value"`
	Rarity      string `json:"rarity"`
}
