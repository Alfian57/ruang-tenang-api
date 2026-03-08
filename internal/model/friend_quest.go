package model

import (
	"time"

	"github.com/google/uuid"
)

// FriendQuestStatus defines quest lifecycle
type FriendQuestStatus string

const (
	FriendQuestPending   FriendQuestStatus = "pending"
	FriendQuestActive    FriendQuestStatus = "active"
	FriendQuestCompleted FriendQuestStatus = "completed"
	FriendQuestExpired   FriendQuestStatus = "expired"
	FriendQuestDeclined  FriendQuestStatus = "declined"
)

// FriendQuestType defines quest challenge types
type FriendQuestType string

const (
	FQTypeXP        FriendQuestType = "total_xp"
	FQTypeBreathing FriendQuestType = "breathing"
	FQTypeJournal   FriendQuestType = "journal"
	FQTypeChat      FriendQuestType = "chat"
	FQTypeMood      FriendQuestType = "mood"
)

// FriendQuest represents a collaborative 1-on-1 mission
type FriendQuest struct {
	ID                uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RequesterID       uint              `gorm:"not null" json:"requester_id"`
	PartnerID         uint              `gorm:"not null" json:"partner_id"`
	Title             string            `gorm:"size:200;not null" json:"title"`
	Description       string            `gorm:"type:text" json:"description"`
	QuestType         FriendQuestType   `gorm:"size:50;not null;default:'total_xp'" json:"quest_type"`
	TargetValue       int               `gorm:"not null" json:"target_value"`
	RequesterProgress int               `gorm:"not null;default:0" json:"requester_progress"`
	PartnerProgress   int               `gorm:"not null;default:0" json:"partner_progress"`
	XPReward          int               `gorm:"not null;default:0" json:"xp_reward"`
	CoinReward        int               `gorm:"not null;default:0" json:"coin_reward"`
	Status            FriendQuestStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	StartsAt          *time.Time        `json:"starts_at,omitempty"`
	EndsAt            *time.Time        `json:"ends_at,omitempty"`
	CompletedAt       *time.Time        `json:"completed_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Requester         *User             `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	Partner           *User             `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
}

func (FriendQuest) TableName() string { return "friend_quests" }

// TotalProgress returns combined progress of both users
func (q *FriendQuest) TotalProgress() int {
	return q.RequesterProgress + q.PartnerProgress
}

// ProgressPercent returns progress percentage
func (q *FriendQuest) ProgressPercent() float64 {
	if q.TargetValue == 0 {
		return 0
	}
	p := float64(q.TotalProgress()) / float64(q.TargetValue) * 100
	if p > 100 {
		return 100
	}
	return p
}
