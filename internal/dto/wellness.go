package dto

import "time"

type WellnessOnboardingRequest struct {
	InitialMood string   `json:"initial_mood" binding:"required,max=50"`
	Goals       []string `json:"goals" binding:"required,min=1,max=5,dive,max=80"`
	Habits      []string `json:"habits" binding:"omitempty,max=8,dive,max=80"`
}

type WellnessProfileDTO struct {
	ID                    uint       `json:"id"`
	UserID                uint       `json:"user_id"`
	InitialMood           string     `json:"initial_mood"`
	Goals                 []string   `json:"goals"`
	Habits                []string   `json:"habits"`
	TourCompletedAt       *time.Time `json:"tour_completed_at,omitempty"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
}

type WellnessOnboardingResponse struct {
	NeedsOnboarding bool                `json:"needs_onboarding"`
	Profile         *WellnessProfileDTO `json:"profile,omitempty"`
	Plan            *WellnessPlanDTO    `json:"plan,omitempty"`
}

type WellnessPlanItemDTO struct {
	ID          string         `json:"id"`
	DayNumber   int            `json:"day_number"`
	ItemDate    string         `json:"item_date"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ActionType  string         `json:"action_type"`
	Route       string         `json:"route"`
	Status      string         `json:"status"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Metadata    map[string]any `json:"metadata"`
}

type WellnessPlanDTO struct {
	ID                string                `json:"id"`
	Title             string                `json:"title"`
	Summary           string                `json:"summary"`
	Status            string                `json:"status"`
	StartsOn          string                `json:"starts_on"`
	EndsOn            string                `json:"ends_on"`
	GeneratedFromMood string                `json:"generated_from_mood"`
	CompletionPercent int                   `json:"completion_percent"`
	Items             []WellnessPlanItemDTO `json:"items"`
}

type WellnessNeedNowRequest struct {
	Condition string `json:"condition" binding:"required,oneof=cemas capek sedih marah bingung fokus"`
}

type WellnessRecommendationDTO struct {
	Type        string         `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Route       string         `json:"route"`
	Prompt      string         `json:"prompt,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Locked      bool           `json:"locked"`
}

type WellnessNeedNowResponse struct {
	Condition       string                      `json:"condition"`
	Title           string                      `json:"title"`
	Description     string                      `json:"description"`
	Recommendations []WellnessRecommendationDTO `json:"recommendations"`
}

type WeeklyInsightResponse struct {
	ID              string                      `json:"id"`
	WeekStart       string                      `json:"week_start"`
	WeekEnd         string                      `json:"week_end"`
	MoodSummary     map[string]any              `json:"mood_summary"`
	ActivitySummary map[string]any              `json:"activity_summary"`
	Insight         map[string]any              `json:"insight"`
	Narrative       string                      `json:"narrative"`
	Recommendations []WellnessRecommendationDTO `json:"recommendations"`
	PremiumPreview  map[string]any              `json:"premium_preview"`
	IsPremium       bool                        `json:"is_premium"`
	IsAIEnhanced    bool                        `json:"is_ai_enhanced"`
	GeneratedAt     time.Time                   `json:"generated_at"`
}

type WellnessTourCompleteResponse struct {
	TourCompletedAt time.Time `json:"tour_completed_at"`
}

type WellnessJourneyNodeDTO struct {
	Key         string  `json:"key"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Value       int     `json:"value"`
	Target      int     `json:"target"`
	Progress    float64 `json:"progress"`
	Route       string  `json:"route"`
	Tone        string  `json:"tone"`
}

type WellnessJourneyMapResponse struct {
	Title              string                    `json:"title"`
	Narrative          string                    `json:"narrative"`
	OverallProgress    int                       `json:"overall_progress"`
	Streak             int                       `json:"streak"`
	Nodes              []WellnessJourneyNodeDTO  `json:"nodes"`
	NextRecommendation WellnessRecommendationDTO `json:"next_recommendation"`
}
