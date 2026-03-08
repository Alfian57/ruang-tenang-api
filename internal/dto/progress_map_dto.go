package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// Response DTOs
// ==========================================

// MapRegionResponse represents a map region in API responses
type MapRegionResponse struct {
	ID                uuid.UUID             `json:"id"`
	RegionKey         string                `json:"region_key"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Icon              string                `json:"icon"`
	Image             string                `json:"image"`
	UnlockType        string                `json:"unlock_type"`
	UnlockValue       int                   `json:"unlock_value"`
	PositionX         int                   `json:"position_x"`
	PositionY         int                   `json:"position_y"`
	DisplayOrder      int                   `json:"display_order"`
	IsUnlocked        bool                  `json:"is_unlocked"`
	UnlockedAt        *time.Time            `json:"unlocked_at,omitempty"`
	Landmarks         []MapLandmarkResponse `json:"landmarks,omitempty"`
	TotalLandmarks    int                   `json:"total_landmarks"`
	UnlockedLandmarks int                   `json:"unlocked_landmarks"`
}

// MapLandmarkResponse represents a landmark in API responses
type MapLandmarkResponse struct {
	ID              uuid.UUID  `json:"id"`
	LandmarkKey     string     `json:"landmark_key"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Icon            string     `json:"icon"`
	UnlockType      string     `json:"unlock_type"`
	UnlockActivity  string     `json:"unlock_activity,omitempty"`
	UnlockValue     int        `json:"unlock_value"`
	PositionX       int        `json:"position_x"`
	PositionY       int        `json:"position_y"`
	XPReward        int        `json:"xp_reward"`
	CoinReward      int        `json:"coin_reward"`
	IsUnlocked      bool       `json:"is_unlocked"`
	CurrentValue    int        `json:"current_value"`
	ProgressPercent float64    `json:"progress_percent"`
	UnlockedAt      *time.Time `json:"unlocked_at,omitempty"`
	RewardClaimed   bool       `json:"reward_claimed"`
}

// FullMapResponse represents the complete map with all regions
type FullMapResponse struct {
	Regions           []MapRegionResponse `json:"regions"`
	TotalRegions      int                 `json:"total_regions"`
	UnlockedRegions   int                 `json:"unlocked_regions"`
	TotalLandmarks    int                 `json:"total_landmarks"`
	UnlockedLandmarks int                 `json:"unlocked_landmarks"`
	OverallProgress   float64             `json:"overall_progress"`
}

// MapProgressSummary gives a brief summary of map progress
type MapProgressSummary struct {
	UnlockedRegions   int        `json:"unlocked_regions"`
	TotalRegions      int        `json:"total_regions"`
	UnlockedLandmarks int        `json:"unlocked_landmarks"`
	TotalLandmarks    int        `json:"total_landmarks"`
	OverallProgress   float64    `json:"overall_progress"`
	LatestUnlock      string     `json:"latest_unlock,omitempty"`
	LatestUnlockAt    *time.Time `json:"latest_unlock_at,omitempty"`
}
