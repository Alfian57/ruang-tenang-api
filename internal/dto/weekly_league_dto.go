package dto

import (
	"time"

	"github.com/google/uuid"
)

// === WEEKLY LEAGUE DTOs ===

// LeagueDivisionResponse represents a league division
type LeagueDivisionResponse struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Tier           int    `json:"tier"`
	Color          string `json:"color"`
	PromotionSlots int    `json:"promotion_slots"`
	DemotionSlots  int    `json:"demotion_slots"`
}

// LeagueSeasonResponse represents the current season
type LeagueSeasonResponse struct {
	ID         uuid.UUID `json:"id"`
	WeekNumber int       `json:"week_number"`
	Year       int       `json:"year"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	IsActive   bool      `json:"is_active"`
}

// LeagueParticipantResponse represents a leaderboard entry
type LeagueParticipantResponse struct {
	Rank       int    `json:"rank"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar,omitempty"`
	WeeklyXP   int64  `json:"weekly_xp"`
	IsPromoted bool   `json:"is_promoted"`
	IsDemoted  bool   `json:"is_demoted"`
	IsMe       bool   `json:"is_me"`
}

// LeagueOverviewResponse is the main league view
type LeagueOverviewResponse struct {
	Season       LeagueSeasonResponse        `json:"season"`
	Division     LeagueDivisionResponse      `json:"division"`
	MyRank       int                         `json:"my_rank"`
	MyWeeklyXP   int64                       `json:"my_weekly_xp"`
	Leaderboard  []LeagueParticipantResponse `json:"leaderboard"`
	TimeLeftSecs int                         `json:"time_left_seconds"`
}
