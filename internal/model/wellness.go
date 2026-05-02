package model

import (
	"time"

	"github.com/google/uuid"
)

type WellnessPlanStatus string

const (
	WellnessPlanStatusActive    WellnessPlanStatus = "active"
	WellnessPlanStatusCompleted WellnessPlanStatus = "completed"
	WellnessPlanStatusArchived  WellnessPlanStatus = "archived"
)

type WellnessPlanItemStatus string

const (
	WellnessPlanItemStatusPending   WellnessPlanItemStatus = "pending"
	WellnessPlanItemStatusCompleted WellnessPlanItemStatus = "completed"
	WellnessPlanItemStatusSkipped   WellnessPlanItemStatus = "skipped"
)

type UserWellnessProfile struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	UserID                uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	InitialMood           string     `gorm:"size:50;not null;default:''" json:"initial_mood"`
	GoalsJSON             string     `gorm:"type:jsonb;not null;default:'[]'" json:"goals_json"`
	HabitsJSON            string     `gorm:"type:jsonb;not null;default:'[]'" json:"habits_json"`
	TourCompletedAt       *time.Time `json:"tour_completed_at,omitempty"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (UserWellnessProfile) TableName() string {
	return "user_wellness_profiles"
}

type WellnessPlan struct {
	ID                uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uint               `gorm:"not null;index" json:"user_id"`
	ProfileID         *uint              `json:"profile_id,omitempty"`
	Title             string             `gorm:"size:180;not null" json:"title"`
	Summary           string             `gorm:"type:text;not null;default:''" json:"summary"`
	Status            WellnessPlanStatus `gorm:"size:30;not null;default:'active'" json:"status"`
	StartsOn          time.Time          `gorm:"type:date;not null" json:"starts_on"`
	EndsOn            time.Time          `gorm:"type:date;not null" json:"ends_on"`
	GeneratedFromMood string             `gorm:"size:50;not null;default:''" json:"generated_from_mood"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	Items             []WellnessPlanItem `gorm:"foreignKey:PlanID" json:"items,omitempty"`
}

func (WellnessPlan) TableName() string {
	return "wellness_plans"
}

type WellnessPlanItem struct {
	ID           uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlanID       uuid.UUID              `gorm:"type:uuid;not null;index" json:"plan_id"`
	UserID       uint                   `gorm:"not null;index" json:"user_id"`
	DayNumber    int                    `gorm:"not null" json:"day_number"`
	ItemDate     time.Time              `gorm:"type:date;not null" json:"item_date"`
	Title        string                 `gorm:"size:180;not null" json:"title"`
	Description  string                 `gorm:"type:text;not null;default:''" json:"description"`
	ActionType   string                 `gorm:"size:50;not null" json:"action_type"`
	Route        string                 `gorm:"size:255;not null;default:''" json:"route"`
	Status       WellnessPlanItemStatus `gorm:"size:30;not null;default:'pending'" json:"status"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	MetadataJSON string                 `gorm:"type:jsonb;not null;default:'{}'" json:"metadata_json"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

func (WellnessPlanItem) TableName() string {
	return "wellness_plan_items"
}

type WeeklyInsightSnapshot struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID              uint      `gorm:"not null;index" json:"user_id"`
	WeekStart           time.Time `gorm:"type:date;not null" json:"week_start"`
	WeekEnd             time.Time `gorm:"type:date;not null" json:"week_end"`
	MoodSummaryJSON     string    `gorm:"type:jsonb;not null;default:'{}'" json:"mood_summary_json"`
	ActivitySummaryJSON string    `gorm:"type:jsonb;not null;default:'{}'" json:"activity_summary_json"`
	InsightJSON         string    `gorm:"type:jsonb;not null;default:'{}'" json:"insight_json"`
	PremiumSectionsJSON string    `gorm:"type:jsonb;not null;default:'{}'" json:"premium_sections_json"`
	Narrative           string    `gorm:"type:text;not null;default:''" json:"narrative"`
	RecommendationsJSON string    `gorm:"type:jsonb;not null;default:'[]'" json:"recommendations_json"`
	IsAIEnhanced        bool      `gorm:"not null;default:false" json:"is_ai_enhanced"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (WeeklyInsightSnapshot) TableName() string {
	return "weekly_insight_snapshots"
}

type WellnessNeedEvent struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	UserID              uint      `gorm:"not null;index" json:"user_id"`
	Condition           string    `gorm:"size:50;not null" json:"condition"`
	RecommendationsJSON string    `gorm:"type:jsonb;not null;default:'[]'" json:"recommendations_json"`
	CreatedAt           time.Time `json:"created_at"`
}

func (WellnessNeedEvent) TableName() string {
	return "wellness_need_events"
}
