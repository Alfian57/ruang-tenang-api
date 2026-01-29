package dto

import "github.com/google/uuid"

// LockedFeatureResponse represents a locked feature with unlock requirements
type LockedFeatureResponse struct {
	ID            uuid.UUID `json:"id"`
	FeatureKey    string    `json:"feature_key"`
	FeatureName   string    `json:"feature_name"`
	Description   string    `json:"description"`
	Icon          string    `json:"icon"`
	Category      string    `json:"category"`
	RequiredLevel int       `json:"required_level"`
	LevelsAway    int       `json:"levels_away"`
}

// FeaturesByLevelResponse represents features grouped by level
type FeaturesByLevelResponse struct {
	Level     int                     `json:"level"`
	TierName  string                  `json:"tier_name"`
	TierColor string                  `json:"tier_color"`
	Features  []FeatureUnlockResponse `json:"features"`
}

// UserFeaturesResponse represents user's feature unlock status
type UserFeaturesResponse struct {
	CurrentLevel     int                     `json:"current_level"`
	TotalUnlocked    int                     `json:"total_unlocked"`
	TotalFeatures    int                     `json:"total_features"`
	UnlockedFeatures []FeatureUnlockResponse `json:"unlocked_features"`
	LockedFeatures   []LockedFeatureResponse `json:"locked_features"`
}

// FeatureAccessResponse represents access check result for a feature
type FeatureAccessResponse struct {
	HasAccess     bool   `json:"has_access"`
	FeatureKey    string `json:"feature_key,omitempty"`
	FeatureName   string `json:"feature_name,omitempty"`
	RequiredLevel int    `json:"required_level,omitempty"`
	CurrentLevel  int    `json:"current_level,omitempty"`
	LevelsAway    int    `json:"levels_away,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// FeatureCategoryInfo represents feature category metadata
type FeatureCategoryInfo struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
