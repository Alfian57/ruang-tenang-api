package model

import (
	"time"

	"github.com/google/uuid"
)

// GuildMemberRole defines roles within a guild
type GuildMemberRole string

const (
	GuildRoleLeader GuildMemberRole = "leader"
	GuildRoleAdmin  GuildMemberRole = "admin"
	GuildRoleMember GuildMemberRole = "member"
)

// GuildChallengeType defines types of guild challenges
type GuildChallengeType string

const (
	GuildChallengeXP        GuildChallengeType = "total_xp"
	GuildChallengeTask      GuildChallengeType = "total_tasks"
	GuildChallengeBreathing GuildChallengeType = "total_breathing"
	GuildChallengeJournal   GuildChallengeType = "total_journals"
	GuildChallengeChat      GuildChallengeType = "total_chats"
	GuildChallengeStreak    GuildChallengeType = "total_streak_days"
)

// GuildActivityType defines types of guild activities
type GuildActivityType string

const (
	GuildActivityCreated           GuildActivityType = "guild_created"
	GuildActivityMemberJoined      GuildActivityType = "member_joined"
	GuildActivityMemberLeft        GuildActivityType = "member_left"
	GuildActivityMemberKicked      GuildActivityType = "member_kicked"
	GuildActivityMemberPromoted    GuildActivityType = "member_promoted"
	GuildActivityChallengeCreated  GuildActivityType = "challenge_created"
	GuildActivityChallengeComplete GuildActivityType = "challenge_completed"
	GuildActivityXPContributed     GuildActivityType = "xp_contributed"
)

// Guild represents a group of users working together
type Guild struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Icon        string    `gorm:"size:255;default:'shield'" json:"icon"`
	Banner      string    `gorm:"size:500" json:"banner"`
	LeaderID    uint      `gorm:"not null" json:"leader_id"`
	MaxMembers  int       `gorm:"not null;default:10" json:"max_members"`
	TotalXP     int64     `gorm:"not null;default:0" json:"total_xp"`
	Level       int       `gorm:"not null;default:1" json:"level"`
	IsPublic    bool      `gorm:"default:true" json:"is_public"`
	InviteCode  string    `gorm:"size:20;uniqueIndex" json:"invite_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Leader  *User         `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
	Members []GuildMember `gorm:"foreignKey:GuildID" json:"members,omitempty"`
}

func (Guild) TableName() string {
	return "guilds"
}

// GuildMember represents a user's membership in a guild
type GuildMember struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GuildID       uuid.UUID       `gorm:"type:uuid;not null" json:"guild_id"`
	UserID        uint            `gorm:"not null" json:"user_id"`
	Role          GuildMemberRole `gorm:"size:20;not null;default:'member'" json:"role"`
	XPContributed int64           `gorm:"not null;default:0" json:"xp_contributed"`
	JoinedAt      time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`

	// Relations
	Guild *Guild `gorm:"foreignKey:GuildID" json:"guild,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GuildMember) TableName() string {
	return "guild_members"
}

// GuildChallenge represents a collaborative challenge for guild members
type GuildChallenge struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GuildID       uuid.UUID          `gorm:"type:uuid;not null" json:"guild_id"`
	Title         string             `gorm:"size:200;not null" json:"title"`
	Description   string             `gorm:"type:text" json:"description"`
	ChallengeType GuildChallengeType `gorm:"size:50;not null" json:"challenge_type"`
	TargetValue   int                `gorm:"not null" json:"target_value"`
	CurrentValue  int                `gorm:"not null;default:0" json:"current_value"`
	XPReward      int                `gorm:"not null;default:0" json:"xp_reward"`
	CoinReward    int                `gorm:"not null;default:0" json:"coin_reward"`
	StartsAt      time.Time          `gorm:"not null" json:"starts_at"`
	EndsAt        time.Time          `gorm:"not null" json:"ends_at"`
	IsCompleted   bool               `gorm:"default:false" json:"is_completed"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`

	// Relations
	Guild         *Guild                       `gorm:"foreignKey:GuildID" json:"guild,omitempty"`
	Contributions []GuildChallengeContribution `gorm:"foreignKey:ChallengeID" json:"contributions,omitempty"`
}

func (GuildChallenge) TableName() string {
	return "guild_challenges"
}

// ProgressPercent returns challenge completion percentage
func (gc *GuildChallenge) ProgressPercent() float64 {
	if gc.TargetValue == 0 {
		return 0
	}
	p := float64(gc.CurrentValue) / float64(gc.TargetValue) * 100
	if p > 100 {
		return 100
	}
	return p
}

// IsExpired returns true if the challenge has passed its end date
func (gc *GuildChallenge) IsExpired() bool {
	return time.Now().After(gc.EndsAt)
}

// IsActive returns true if the challenge is currently active
func (gc *GuildChallenge) IsActive() bool {
	now := time.Now()
	return now.After(gc.StartsAt) && now.Before(gc.EndsAt) && !gc.IsCompleted
}

// GuildChallengeContribution tracks individual member contributions to a challenge
type GuildChallengeContribution struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ChallengeID   uuid.UUID `gorm:"type:uuid;not null" json:"challenge_id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	Value         int       `gorm:"not null;default:0" json:"value"`
	ContributedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"contributed_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GuildChallengeContribution) TableName() string {
	return "guild_challenge_contributions"
}

// GuildActivity logs guild events
type GuildActivity struct {
	ID           uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GuildID      uuid.UUID         `gorm:"type:uuid;not null" json:"guild_id"`
	UserID       *uint             `json:"user_id,omitempty"`
	ActivityType GuildActivityType `gorm:"size:50;not null" json:"activity_type"`
	Description  string            `gorm:"type:text;not null" json:"description"`
	CreatedAt    time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GuildActivity) TableName() string {
	return "guild_activities"
}
