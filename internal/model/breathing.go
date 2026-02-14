package model

import (
	"time"

	"github.com/google/uuid"
)

// BreathingTechnique represents a breathing technique (system or custom)
type BreathingTechnique struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name               string    `json:"name" gorm:"size:100;not null"`
	Slug               *string   `json:"slug" gorm:"size:100;uniqueIndex"`
	Description        *string   `json:"description"`
	Benefits           *string   `json:"benefits"`
	BestFor            *string   `json:"best_for"`
	InhaleDuration     int       `json:"inhale_duration" gorm:"not null;default:4"`
	InhaleHoldDuration int       `json:"inhale_hold_duration" gorm:"not null;default:0"`
	ExhaleDuration     int       `json:"exhale_duration" gorm:"not null;default:4"`
	ExhaleHoldDuration int       `json:"exhale_hold_duration" gorm:"not null;default:0"`
	Icon               string    `json:"icon" gorm:"size:50;default:'🌬️'"`
	Color              string    `json:"color" gorm:"size:20;default:'#6366F1'"`
	AnimationType      string    `json:"animation_type" gorm:"size:50;default:'circle'"`
	Difficulty         string    `json:"difficulty" gorm:"size:20;default:'easy'"`
	Category           string    `json:"category" gorm:"size:50;default:'general'"`
	Origin             *string   `json:"origin"`
	IsSystem           bool      `json:"is_system" gorm:"default:false"`
	IsActive           bool      `json:"is_active" gorm:"default:true"`
	UserID             *int      `json:"user_id" gorm:"index"`
	User               *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName specifies the table name for BreathingTechnique
func (BreathingTechnique) TableName() string {
	return "breathing_techniques"
}

// GetTotalCycleDuration returns the total duration of one breathing cycle in seconds
func (bt *BreathingTechnique) GetTotalCycleDuration() int {
	return bt.InhaleDuration + bt.InhaleHoldDuration + bt.ExhaleDuration + bt.ExhaleHoldDuration
}

// GetCyclesForDuration returns how many cycles fit in the given duration
func (bt *BreathingTechnique) GetCyclesForDuration(durationSeconds int) int {
	cycleDuration := bt.GetTotalCycleDuration()
	if cycleDuration == 0 {
		return 0
	}
	return durationSeconds / cycleDuration
}

// BreathingSession represents a user's breathing practice session
type BreathingSession struct {
	ID                    uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID                int                 `json:"user_id" gorm:"not null;index"`
	User                  *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TechniqueID           uuid.UUID           `json:"technique_id" gorm:"type:uuid;not null;index"`
	Technique             *BreathingTechnique `json:"technique,omitempty" gorm:"foreignKey:TechniqueID"`
	DurationSeconds       int                 `json:"duration_seconds" gorm:"not null"`
	TargetDurationSeconds int                 `json:"target_duration_seconds" gorm:"not null"`
	CyclesCompleted       int                 `json:"cycles_completed" gorm:"not null;default:0"`
	VoiceGuidanceEnabled  bool                `json:"voice_guidance_enabled" gorm:"default:false"`
	BackgroundSound       *string             `json:"background_sound" gorm:"size:50"`
	HapticFeedbackEnabled bool                `json:"haptic_feedback_enabled" gorm:"default:false"`
	Completed             bool                `json:"completed" gorm:"default:false"`
	CompletedPercentage   int                 `json:"completed_percentage" gorm:"default:0"`
	StartedAt             time.Time           `json:"started_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	EndedAt               *time.Time          `json:"ended_at"`
	XPEarned              int                 `json:"xp_earned" gorm:"default:0"`
	MoodBefore            *string             `json:"mood_before" gorm:"size:20"`
	MoodAfter             *string             `json:"mood_after" gorm:"size:20"`
	CreatedAt             time.Time           `json:"created_at"`
}

// TableName specifies the table name for BreathingSession
func (BreathingSession) TableName() string {
	return "breathing_sessions"
}

// BreathingPreference represents user's breathing exercise preferences
type BreathingPreference struct {
	ID                     int                 `json:"id" gorm:"primaryKey"`
	UserID                 int                 `json:"user_id" gorm:"uniqueIndex;not null"`
	User                   *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	DefaultDurationSeconds int                 `json:"default_duration_seconds" gorm:"default:300"`
	DefaultTechniqueID     *uuid.UUID          `json:"default_technique_id" gorm:"type:uuid"`
	DefaultTechnique       *BreathingTechnique `json:"default_technique,omitempty" gorm:"foreignKey:DefaultTechniqueID"`
	VoiceGuidance          string              `json:"voice_guidance" gorm:"size:20;default:'ask'"`
	BackgroundSound        string              `json:"background_sound" gorm:"size:20;default:'ask'"`
	DefaultBackgroundSound string              `json:"default_background_sound" gorm:"size:50;default:'none'"`
	HapticFeedback         bool                `json:"haptic_feedback" gorm:"default:true"`
	AnimationSpeed         string              `json:"animation_speed" gorm:"size:20;default:'normal'"`
	Theme                  string              `json:"theme" gorm:"size:50;default:'default'"`
	ReminderEnabled        bool                `json:"reminder_enabled" gorm:"default:false"`
	ReminderTime           *string             `json:"reminder_time"`
	ReminderDays           string              `json:"reminder_days" gorm:"size:20;default:'1234567'"`
	TutorialCompleted      bool                `json:"tutorial_completed" gorm:"default:false"`
	CurrentStreak          int                 `json:"current_streak" gorm:"default:0"`
	LongestStreak          int                 `json:"longest_streak" gorm:"default:0"`
	LastPracticeDate       *time.Time          `json:"last_practice_date" gorm:"type:date"`
	StreakFreezeAvailable  bool                `json:"streak_freeze_available" gorm:"default:true"`
	StreakFreezeUsedAt     *time.Time          `json:"streak_freeze_used_at" gorm:"type:date"`
	DailyXPEarned          int                 `json:"daily_xp_earned" gorm:"default:0"`
	DailyXPDate            *time.Time          `json:"daily_xp_date" gorm:"type:date"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
}

// TableName specifies the table name for BreathingPreference
func (BreathingPreference) TableName() string {
	return "breathing_preferences"
}

// BreathingFavorite represents a user's favorite breathing technique
type BreathingFavorite struct {
	ID          int                 `json:"id" gorm:"primaryKey"`
	UserID      int                 `json:"user_id" gorm:"not null;index"`
	User        *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TechniqueID uuid.UUID           `json:"technique_id" gorm:"type:uuid;not null"`
	Technique   *BreathingTechnique `json:"technique,omitempty" gorm:"foreignKey:TechniqueID"`
	SortOrder   int                 `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time           `json:"created_at"`
}

// TableName specifies the table name for BreathingFavorite
func (BreathingFavorite) TableName() string {
	return "breathing_favorites"
}
