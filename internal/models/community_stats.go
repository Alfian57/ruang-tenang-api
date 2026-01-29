package models

import (
	"time"

	"github.com/google/uuid"
)

// MonthlyHallOfFame tracks featured users per level per month
type MonthlyHallOfFame struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Level     int       `gorm:"not null" json:"level"`
	Month     int       `gorm:"not null" json:"month"` // 1-12
	Year      int       `gorm:"not null" json:"year"`
	Rank      int       `gorm:"not null" json:"rank"` // 1, 2, or 3
	MonthlyXP int       `gorm:"not null" json:"monthly_xp"`
	Message   string    `gorm:"size:300" json:"message"` // Optional inspiring message
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (MonthlyHallOfFame) TableName() string {
	return "monthly_hall_of_fame"
}

// CommunityStats tracks monthly community statistics
type CommunityStats struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Month                  int       `gorm:"not null" json:"month"`
	Year                   int       `gorm:"not null" json:"year"`
	TotalXPEarned          int64     `gorm:"default:0" json:"total_xp_earned"`
	ActiveMembers          int       `gorm:"default:0" json:"active_members"`
	TotalAchievements      int       `gorm:"default:0" json:"total_achievements"`
	NewMembers             int       `gorm:"default:0" json:"new_members"`
	TotalStoriesPublished  int       `gorm:"default:0" json:"total_stories_published"`
	TotalArticlesPublished int       `gorm:"default:0" json:"total_articles_published"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (CommunityStats) TableName() string {
	return "community_stats"
}
