package model

import (
	"time"
)

// CrisisCategory represents categories of crisis keywords
type CrisisCategory string

const (
	CrisisCategorySelfHarm         CrisisCategory = "self_harm"
	CrisisCategorySuicide          CrisisCategory = "suicide"
	CrisisCategorySevereDepression CrisisCategory = "severe_depression"
	CrisisCategoryEmergency        CrisisCategory = "emergency"
)

// CrisisSeverity represents severity levels for crisis detection
type CrisisSeverity string

const (
	CrisisSeverityMedium   CrisisSeverity = "medium"
	CrisisSeverityHigh     CrisisSeverity = "high"
	CrisisSeverityCritical CrisisSeverity = "critical"
)

// CrisisKeyword represents a keyword for crisis detection
type CrisisKeyword struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Keyword   string         `gorm:"size:255;not null" json:"keyword"`
	Category  CrisisCategory `gorm:"size:100;not null" json:"category"`
	Severity  CrisisSeverity `gorm:"size:20;default:'high'" json:"severity"`
	Language  string         `gorm:"size:10;default:'id'" json:"language"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	Notes     string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (CrisisKeyword) TableName() string {
	return "crisis_keywords"
}

// CrisisDetectionResult represents the result of crisis detection
type CrisisDetectionResult struct {
	IsCrisis        bool           `json:"is_crisis"`
	Keywords        []string       `json:"keywords,omitempty"`
	Category        CrisisCategory `json:"category,omitempty"`
	Severity        CrisisSeverity `json:"severity,omitempty"`
	CrisisResponse  string         `json:"crisis_response,omitempty"`
	EmergencyNumber string         `json:"emergency_number,omitempty"`
}
