package model

import (
	"time"

	"github.com/google/uuid"
)

// LeagueDivision represents a league tier (Bronze → Diamond)
type LeagueDivision struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Icon           string    `gorm:"size:50;not null" json:"icon"`
	Tier           int       `gorm:"not null;uniqueIndex" json:"tier"`
	Color          string    `gorm:"size:20;not null;default:'#888888'" json:"color"`
	MinRank        int       `gorm:"not null;default:0" json:"min_rank"`
	PromotionSlots int       `gorm:"not null;default:10" json:"promotion_slots"`
	DemotionSlots  int       `gorm:"not null;default:5" json:"demotion_slots"`
	CreatedAt      time.Time `json:"created_at"`
}

func (LeagueDivision) TableName() string { return "league_divisions" }

// LeagueSeason represents a weekly competition period
type LeagueSeason struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WeekNumber  int       `gorm:"not null" json:"week_number"`
	Year        int       `gorm:"not null" json:"year"`
	StartsAt    time.Time `gorm:"not null" json:"starts_at"`
	EndsAt      time.Time `gorm:"not null" json:"ends_at"`
	IsActive    bool      `gorm:"default:false" json:"is_active"`
	IsProcessed bool      `gorm:"default:false" json:"is_processed"`
	CreatedAt   time.Time `json:"created_at"`
}

func (LeagueSeason) TableName() string { return "league_seasons" }

// LeagueParticipant represents a user's participation in a weekly league
type LeagueParticipant struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SeasonID   uuid.UUID       `gorm:"type:uuid;not null" json:"season_id"`
	UserID     uint            `gorm:"not null" json:"user_id"`
	DivisionID int             `gorm:"not null" json:"division_id"`
	WeeklyXP   int64           `gorm:"not null;default:0" json:"weekly_xp"`
	Rank       int             `gorm:"not null;default:0" json:"rank"`
	IsPromoted bool            `gorm:"default:false" json:"is_promoted"`
	IsDemoted  bool            `gorm:"default:false" json:"is_demoted"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Season     *LeagueSeason   `gorm:"foreignKey:SeasonID" json:"season,omitempty"`
	User       *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Division   *LeagueDivision `gorm:"foreignKey:DivisionID" json:"division,omitempty"`
}

func (LeagueParticipant) TableName() string { return "league_participants" }
