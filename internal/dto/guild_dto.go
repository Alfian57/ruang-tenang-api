package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// Request DTOs
// ==========================================

// CreateGuildRequest represents the request to create a guild
type CreateGuildRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=50"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateGuildRequest represents the request to update a guild
type UpdateGuildRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=3,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Icon        *string `json:"icon" binding:"omitempty,max=50"`
	Banner      *string `json:"banner" binding:"omitempty,max=500"`
	IsPublic    *bool   `json:"is_public"`
}

// CreateGuildChallengeRequest represents the request to create a guild challenge
type CreateGuildChallengeRequest struct {
	Title         string `json:"title" binding:"required,min=3,max=200"`
	Description   string `json:"description" binding:"max=500"`
	ChallengeType string `json:"challenge_type" binding:"required,oneof=total_xp total_tasks total_breathing total_journals total_chats total_streak_days"`
	TargetValue   int    `json:"target_value" binding:"required,min=1"`
	XPReward      int    `json:"xp_reward" binding:"min=0,max=500"`
	CoinReward    int    `json:"coin_reward" binding:"min=0,max=100"`
	DurationDays  int    `json:"duration_days" binding:"required,min=1,max=30"`
}

// ==========================================
// Response DTOs
// ==========================================

// GuildResponse represents a guild in API responses
type GuildResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Banner      string    `json:"banner"`
	LeaderID    uint      `json:"leader_id"`
	LeaderName  string    `json:"leader_name"`
	MaxMembers  int       `json:"max_members"`
	MemberCount int       `json:"member_count"`
	TotalXP     int64     `json:"total_xp"`
	Level       int       `json:"level"`
	IsPublic    bool      `json:"is_public"`
	InviteCode  string    `json:"invite_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// GuildDetailResponse extends GuildResponse with members and challenges
type GuildDetailResponse struct {
	GuildResponse
	Members            []GuildMemberResponse    `json:"members"`
	ActiveChallenges   []GuildChallengeResponse `json:"active_challenges"`
	RecentActivities   []GuildActivityResponse  `json:"recent_activities"`
	IsCurrentUserGuild bool                     `json:"is_current_user_guild"`
	CurrentUserRole    string                   `json:"current_user_role,omitempty"`
}

// GuildMemberResponse represents a guild member
type GuildMemberResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	Name          string    `json:"name"`
	Avatar        string    `json:"avatar"`
	Role          string    `json:"role"`
	XPContributed int64     `json:"xp_contributed"`
	UserLevel     int       `json:"user_level"`
	JoinedAt      time.Time `json:"joined_at"`
}

// GuildChallengeResponse represents a guild challenge
type GuildChallengeResponse struct {
	ID              uuid.UUID                      `json:"id"`
	Title           string                         `json:"title"`
	Description     string                         `json:"description"`
	ChallengeType   string                         `json:"challenge_type"`
	TargetValue     int                            `json:"target_value"`
	CurrentValue    int                            `json:"current_value"`
	ProgressPercent float64                        `json:"progress_percent"`
	XPReward        int                            `json:"xp_reward"`
	CoinReward      int                            `json:"coin_reward"`
	StartsAt        time.Time                      `json:"starts_at"`
	EndsAt          time.Time                      `json:"ends_at"`
	IsCompleted     bool                           `json:"is_completed"`
	IsExpired       bool                           `json:"is_expired"`
	IsActive        bool                           `json:"is_active"`
	TopContributors []GuildChallengeContributorDTO `json:"top_contributors,omitempty"`
	CreatedAt       time.Time                      `json:"created_at"`
}

// GuildChallengeContributorDTO represents a contributor's stats
type GuildChallengeContributorDTO struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Value    int    `json:"value"`
}

// GuildActivityResponse represents a guild activity log entry
type GuildActivityResponse struct {
	ID           uuid.UUID `json:"id"`
	ActivityType string    `json:"activity_type"`
	Description  string    `json:"description"`
	Username     string    `json:"username,omitempty"`
	Avatar       string    `json:"avatar,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// GuildLeaderboardEntry represents a guild in the leaderboard
type GuildLeaderboardEntry struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	TotalXP     int64     `json:"total_xp"`
	Level       int       `json:"level"`
	MemberCount int       `json:"member_count"`
	Rank        int       `json:"rank"`
}

// MyGuildResponse represents the user's guild info
type MyGuildResponse struct {
	Guild         *GuildResponse `json:"guild,omitempty"`
	MemberRole    string         `json:"member_role,omitempty"`
	XPContributed int64          `json:"xp_contributed"`
	IsMember      bool           `json:"is_member"`
}
