package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// Breathing Technique DTOs
// ==========================================

// BreathingTechniqueResponse represents a breathing technique
type BreathingTechniqueResponse struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug,omitempty"`
	Description        string    `json:"description,omitempty"`
	Benefits           string    `json:"benefits,omitempty"`
	BestFor            string    `json:"best_for,omitempty"`
	InhaleDuration     int       `json:"inhale_duration"`
	InhaleHoldDuration int       `json:"inhale_hold_duration"`
	ExhaleDuration     int       `json:"exhale_duration"`
	ExhaleHoldDuration int       `json:"exhale_hold_duration"`
	TotalCycleDuration int       `json:"total_cycle_duration"`
	Icon               string    `json:"icon"`
	Color              string    `json:"color"`
	AnimationType      string    `json:"animation_type"`
	Difficulty         string    `json:"difficulty"`
	Category           string    `json:"category"`
	Origin             string    `json:"origin,omitempty"`
	IsSystem           bool      `json:"is_system"`
	IsFavorite         bool      `json:"is_favorite"`
	CreatedAt          time.Time `json:"created_at"`
}

// CreateBreathingTechniqueRequest for creating custom technique
type CreateBreathingTechniqueRequest struct {
	Name               string `json:"name" binding:"required,min=2,max=100"`
	Description        string `json:"description"`
	InhaleDuration     int    `json:"inhale_duration" binding:"required,min=1,max=10"`
	InhaleHoldDuration int    `json:"inhale_hold_duration" binding:"min=0,max=10"`
	ExhaleDuration     int    `json:"exhale_duration" binding:"required,min=1,max=15"`
	ExhaleHoldDuration int    `json:"exhale_hold_duration" binding:"min=0,max=10"`
	Icon               string `json:"icon"`
	Color              string `json:"color"`
}

// UpdateBreathingTechniqueRequest for updating custom technique
type UpdateBreathingTechniqueRequest struct {
	Name               *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description        *string `json:"description"`
	InhaleDuration     *int    `json:"inhale_duration" binding:"omitempty,min=1,max=10"`
	InhaleHoldDuration *int    `json:"inhale_hold_duration" binding:"omitempty,min=0,max=10"`
	ExhaleDuration     *int    `json:"exhale_duration" binding:"omitempty,min=1,max=15"`
	ExhaleHoldDuration *int    `json:"exhale_hold_duration" binding:"omitempty,min=0,max=10"`
	Icon               *string `json:"icon"`
	Color              *string `json:"color"`
}

// ==========================================
// Breathing Session DTOs
// ==========================================

// StartBreathingSessionRequest for starting a new session
type StartBreathingSessionRequest struct {
	TechniqueID           uuid.UUID `json:"technique_id" binding:"required"`
	TargetDurationSeconds int       `json:"target_duration_seconds" binding:"required,min=60,max=1800"` // 1-30 min
	VoiceGuidanceEnabled  bool      `json:"voice_guidance_enabled"`
	BackgroundSound       string    `json:"background_sound"`
	HapticFeedbackEnabled bool      `json:"haptic_feedback_enabled"`
	MoodBefore            string    `json:"mood_before"`
}

// CompleteBreathingSessionRequest for completing a session
type CompleteBreathingSessionRequest struct {
	DurationSeconds     int    `json:"duration_seconds" binding:"required,min=0"`
	CyclesCompleted     int    `json:"cycles_completed" binding:"min=0"`
	Completed           bool   `json:"completed"`
	CompletedPercentage int    `json:"completed_percentage" binding:"min=0,max=100"`
	MoodAfter           string `json:"mood_after"`
}

// BreathingSessionResponse represents a completed session
type BreathingSessionResponse struct {
	ID                    uuid.UUID                   `json:"id"`
	TechniqueID           uuid.UUID                   `json:"technique_id"`
	Technique             *BreathingTechniqueResponse `json:"technique,omitempty"`
	DurationSeconds       int                         `json:"duration_seconds"`
	TargetDurationSeconds int                         `json:"target_duration_seconds"`
	CyclesCompleted       int                         `json:"cycles_completed"`
	VoiceGuidanceEnabled  bool                        `json:"voice_guidance_enabled"`
	BackgroundSound       string                      `json:"background_sound,omitempty"`
	HapticFeedbackEnabled bool                        `json:"haptic_feedback_enabled"`
	Completed             bool                        `json:"completed"`
	CompletedPercentage   int                         `json:"completed_percentage"`
	StartedAt             time.Time                   `json:"started_at"`
	EndedAt               *time.Time                  `json:"ended_at,omitempty"`
	XPEarned              int                         `json:"xp_earned"`
	MoodBefore            string                      `json:"mood_before,omitempty"`
	MoodAfter             string                      `json:"mood_after,omitempty"`
}

// SessionCompletionResult returned after completing a session
type SessionCompletionResult struct {
	Session           BreathingSessionResponse `json:"session"`
	XPEarned          int                      `json:"xp_earned"`
	BonusXP           int                      `json:"bonus_xp"`
	BonusReason       string                   `json:"bonus_reason,omitempty"`
	TotalXP           int                      `json:"total_xp"`
	NewStreak         int                      `json:"new_streak"`
	StreakMilestone   bool                     `json:"streak_milestone"`
	StreakMilestoneXP int                      `json:"streak_milestone_xp,omitempty"`
	DailyXPRemaining  int                      `json:"daily_xp_remaining"`
	NewBadges         []string                 `json:"new_badges,omitempty"`
}

// ==========================================
// Breathing Preferences DTOs
// ==========================================

// BreathingPreferencesResponse represents user preferences
type BreathingPreferencesResponse struct {
	DefaultDurationSeconds int        `json:"default_duration_seconds"`
	DefaultTechniqueID     *uuid.UUID `json:"default_technique_id,omitempty"`
	VoiceGuidance          string     `json:"voice_guidance"`
	BackgroundSound        string     `json:"background_sound"`
	DefaultBackgroundSound string     `json:"default_background_sound"`
	HapticFeedback         bool       `json:"haptic_feedback"`
	AnimationSpeed         string     `json:"animation_speed"`
	Theme                  string     `json:"theme"`
	ReminderEnabled        bool       `json:"reminder_enabled"`
	ReminderTime           string     `json:"reminder_time,omitempty"`
	ReminderDays           string     `json:"reminder_days"`
	TutorialCompleted      bool       `json:"tutorial_completed"`
}

// UpdateBreathingPreferencesRequest for updating preferences
type UpdateBreathingPreferencesRequest struct {
	DefaultDurationSeconds *int       `json:"default_duration_seconds" binding:"omitempty,min=60,max=1800"`
	DefaultTechniqueID     *uuid.UUID `json:"default_technique_id"`
	VoiceGuidance          *string    `json:"voice_guidance" binding:"omitempty,oneof=always_on always_off ask"`
	BackgroundSound        *string    `json:"background_sound" binding:"omitempty,oneof=always_on always_off ask"`
	DefaultBackgroundSound *string    `json:"default_background_sound"`
	HapticFeedback         *bool      `json:"haptic_feedback"`
	AnimationSpeed         *string    `json:"animation_speed" binding:"omitempty,oneof=slow normal fast"`
	Theme                  *string    `json:"theme"`
	ReminderEnabled        *bool      `json:"reminder_enabled"`
	ReminderTime           *string    `json:"reminder_time"`
	ReminderDays           *string    `json:"reminder_days"`
	TutorialCompleted      *bool      `json:"tutorial_completed"`
}

// ==========================================
// Breathing Stats DTOs
// ==========================================

// BreathingDailyStats represents daily breathing statistics
type BreathingDailyStats struct {
	Date              string `json:"date"`
	SessionsCount     int    `json:"sessions_count"`
	TotalMinutes      int    `json:"total_minutes"`
	FavoriteTechnique string `json:"favorite_technique,omitempty"`
}

// BreathingOverallStats represents overall breathing statistics
type BreathingOverallStats struct {
	TotalSessions          int     `json:"total_sessions"`
	TotalMinutes           int     `json:"total_minutes"`
	CurrentStreak          int     `json:"current_streak"`
	LongestStreak          int     `json:"longest_streak"`
	MostUsedTechnique      string  `json:"most_used_technique,omitempty"`
	MostUsedTechniqueID    string  `json:"most_used_technique_id,omitempty"`
	AverageSessionsPerWeek float64 `json:"average_sessions_per_week"`
	CompletionRate         float64 `json:"completion_rate"` // % of sessions completed fully
}

// BreathingStatsResponse combines all stats
type BreathingStatsResponse struct {
	Today      BreathingDailyStats   `json:"today"`
	Overall    BreathingOverallStats `json:"overall"`
	StreakInfo StreakInfo            `json:"streak_info"`
}

// StreakInfo provides detailed streak information
type StreakInfo struct {
	CurrentStreak         int    `json:"current_streak"`
	LongestStreak         int    `json:"longest_streak"`
	LastPracticeDate      string `json:"last_practice_date,omitempty"`
	StreakFreezeAvailable bool   `json:"streak_freeze_available"`
	DaysUntilStreakBreak  int    `json:"days_until_streak_break"` // 0 = must practice today, 1 = can skip today
}

// BreathingCalendarDay represents a day in the calendar view
type BreathingCalendarDay struct {
	Date           string   `json:"date"`
	SessionsCount  int      `json:"sessions_count"`
	TotalMinutes   int      `json:"total_minutes"`
	TechniquesUsed []string `json:"techniques_used"`
	Intensity      int      `json:"intensity"` // 0-3 for heat map (0=none, 1=light, 2=medium, 3=high)
}

// BreathingCalendarResponse for monthly calendar view
type BreathingCalendarResponse struct {
	Month int                    `json:"month"`
	Year  int                    `json:"year"`
	Days  []BreathingCalendarDay `json:"days"`
}

// TechniqueUsageStats for pie chart
type TechniqueUsageStats struct {
	TechniqueID   string  `json:"technique_id"`
	TechniqueName string  `json:"technique_name"`
	SessionsCount int     `json:"sessions_count"`
	TotalMinutes  int     `json:"total_minutes"`
	Percentage    float64 `json:"percentage"`
}

// ==========================================
// Smart Recommendations DTOs
// ==========================================

// TechniqueRecommendation represents a recommended technique
type TechniqueRecommendation struct {
	Technique BreathingTechniqueResponse `json:"technique"`
	Reason    string                     `json:"reason"`
	Priority  int                        `json:"priority"` // 1 = highest
}

// RecommendationsResponse for smart recommendations
type RecommendationsResponse struct {
	BasedOnMood     []TechniqueRecommendation `json:"based_on_mood,omitempty"`
	BasedOnTime     []TechniqueRecommendation `json:"based_on_time,omitempty"`
	BasedOnActivity []TechniqueRecommendation `json:"based_on_activity,omitempty"`
	DefaultPick     *TechniqueRecommendation  `json:"default_pick,omitempty"`
}

// ==========================================
// Session History DTOs
// ==========================================

// SessionHistoryRequest for fetching session history
type SessionHistoryRequest struct {
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	TechniqueID string `form:"technique_id"`
	Page        int    `form:"page" binding:"min=1"`
	Limit       int    `form:"limit" binding:"min=1,max=50"`
}

// SessionHistoryResponse for paginated session history
type SessionHistoryResponse struct {
	Sessions   []BreathingSessionResponse `json:"sessions"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}

// ==========================================
// Widget DTOs
// ==========================================

// BreathingWidgetData for dashboard widget
type BreathingWidgetData struct {
	CurrentStreak      int                         `json:"current_streak"`
	TodaySessions      int                         `json:"today_sessions"`
	TodayMinutes       int                         `json:"today_minutes"`
	DailyGoalMinutes   int                         `json:"daily_goal_minutes"`
	DailyGoalProgress  int                         `json:"daily_goal_progress"` // percentage
	FavoriteTechnique  *BreathingTechniqueResponse `json:"favorite_technique,omitempty"`
	LastSessionAt      *time.Time                  `json:"last_session_at,omitempty"`
	NeedsPracticeToday bool                        `json:"needs_practice_today"`
	StreakAtRisk       bool                        `json:"streak_at_risk"`
}
