package dto

import (
	"time"
)

// === STREAK SOCIETY DTOs ===

// StreakSocietyResponse represents a society tier
type StreakSocietyResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Icon          string `json:"icon"`
	MinStreak     int    `json:"min_streak"`
	BorderColor   string `json:"border_color"`
	BadgeGlow     bool   `json:"badge_glow"`
	ExclusiveChat bool   `json:"exclusive_chat"`
	MemberCount   int64  `json:"member_count"`
	IsMember      bool   `json:"is_member"`
}

// StreakSocietyOverviewResponse shows all societies + user status
type StreakSocietyOverviewResponse struct {
	CurrentStreak  int                     `json:"current_streak"`
	CurrentSociety *StreakSocietyResponse  `json:"current_society,omitempty"`
	AllSocieties   []StreakSocietyResponse `json:"all_societies"`
}

// SocietyMemberResponse shows a member of a society
type SocietyMemberResponse struct {
	UserID   uint      `json:"user_id"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar,omitempty"`
	Streak   int       `json:"current_streak"`
	JoinedAt time.Time `json:"joined_at"`
}
