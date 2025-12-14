package models

import (
	"time"
)

type MoodType string

const (
	MoodHappy        MoodType = "happy"
	MoodNeutral      MoodType = "neutral"
	MoodAngry        MoodType = "angry"
	MoodDisappointed MoodType = "disappointed"
	MoodSad          MoodType = "sad"
	MoodCrying       MoodType = "crying"
)

type UserMood struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Mood      MoodType  `gorm:"type:varchar(20);not null" json:"mood"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserMood) TableName() string {
	return "user_moods"
}

// GetMoodEmoji returns emoji for the mood
func (m *UserMood) GetMoodEmoji() string {
	switch m.Mood {
	case MoodHappy:
		return "😊"
	case MoodNeutral:
		return "😐"
	case MoodAngry:
		return "😠"
	case MoodDisappointed:
		return "😞"
	case MoodSad:
		return "😢"
	case MoodCrying:
		return "😭"
	default:
		return "🙂"
	}
}
