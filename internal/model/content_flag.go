package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ContentFlagType represents the type of flag
type ContentFlagType string

const (
	ContentFlagTypeAIModeration   ContentFlagType = "ai_moderation"
	ContentFlagTypeTriggerWarning ContentFlagType = "trigger_warning"
	ContentFlagTypeManual         ContentFlagType = "manual"
)

// ContentFlagCategory represents categories of content flags
type ContentFlagCategory string

const (
	ContentFlagCategoryMisinformation ContentFlagCategory = "misinformation"
	ContentFlagCategoryHarmfulAdvice  ContentFlagCategory = "harmful_advice"
	ContentFlagCategorySelfHarm       ContentFlagCategory = "self_harm"
	ContentFlagCategorySuicide        ContentFlagCategory = "suicide"
	ContentFlagCategoryAbuse          ContentFlagCategory = "abuse"
	ContentFlagCategoryTrauma         ContentFlagCategory = "trauma"
	ContentFlagCategoryEatingDisorder ContentFlagCategory = "eating_disorder"
	ContentFlagCategorySubstance      ContentFlagCategory = "substance"
	ContentFlagCategoryHateSpeech     ContentFlagCategory = "hate_speech"
	ContentFlagCategorySpam           ContentFlagCategory = "spam"
	ContentFlagCategoryOther          ContentFlagCategory = "other"
)

// FlagSeverity represents severity levels
type FlagSeverity string

const (
	FlagSeverityLow    FlagSeverity = "low"
	FlagSeverityMedium FlagSeverity = "medium"
	FlagSeverityHigh   FlagSeverity = "high"
)

// ContentFlag represents a content moderation flag
type ContentFlag struct {
	ID              uint                `gorm:"primaryKey" json:"id"`
	ContentType     string              `gorm:"size:50;not null" json:"content_type"` // 'article', 'forum', 'forum_post'
	ContentID       uint                `gorm:"not null" json:"content_id"`
	FlagType        ContentFlagType     `gorm:"size:50;not null" json:"flag_type"`
	FlagCategory    ContentFlagCategory `gorm:"size:100;not null" json:"flag_category"`
	Severity        FlagSeverity        `gorm:"size:20;default:'medium'" json:"severity"`
	AIConfidence    *float64            `gorm:"type:decimal(5,2)" json:"ai_confidence,omitempty"`
	AIReason        string              `gorm:"type:text" json:"ai_reason,omitempty"`
	FlaggedByID     *uint               `json:"flagged_by_id,omitempty"`
	IsResolved      bool                `gorm:"default:false" json:"is_resolved"`
	ResolvedByID    *uint               `json:"resolved_by_id,omitempty"`
	ResolvedAt      *time.Time          `json:"resolved_at,omitempty"`
	ResolutionNotes string              `gorm:"type:text" json:"resolution_notes,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`

	// Relations
	FlaggedBy  *User `gorm:"foreignKey:FlaggedByID" json:"flagged_by,omitempty"`
	ResolvedBy *User `gorm:"foreignKey:ResolvedByID" json:"resolved_by,omitempty"`
}

func (ContentFlag) TableName() string {
	return "content_flags"
}

// TriggerWarnings is a custom type for JSON array of trigger warnings
type TriggerWarnings []string

func (tw TriggerWarnings) Value() (driver.Value, error) {
	if tw == nil {
		return nil, nil
	}
	return json.Marshal(tw)
}

func (tw *TriggerWarnings) Scan(value interface{}) error {
	if value == nil {
		*tw = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, tw)
}
