package model

import (
	"time"
)

// DailyTaskType represents the type of daily task
type DailyTaskType string

const (
	TaskTypeDailyLogin   DailyTaskType = "daily_login"
	TaskTypeRecordMood   DailyTaskType = "record_mood"
	TaskTypeChatAI       DailyTaskType = "chat_ai"
	TaskTypeReadArticle  DailyTaskType = "read_article"
	TaskTypeListenSongs  DailyTaskType = "listen_songs"
	TaskTypeWriteJournal DailyTaskType = "write_journal"
	TaskTypeCommentForum DailyTaskType = "comment_forum"
	TaskTypeBreathing    DailyTaskType = "breathing_exercise"

	TaskTypePremiumChatDeepDive DailyTaskType = "premium_chat_deep_dive"
	TaskTypePremiumBreathingPro DailyTaskType = "premium_breathing_pro"
)

// DailyTaskConfig holds configuration for each task type
type DailyTaskConfig struct {
	Type        DailyTaskType
	Name        string
	Description string
	Icon        string
	XPReward    int
	CoinReward  int
	TargetCount int
	PremiumOnly bool
}

// GetDailyTaskConfigs returns base task configurations.
func GetDailyTaskConfigs() []DailyTaskConfig {
	return []DailyTaskConfig{
		{
			Type:        TaskTypeDailyLogin,
			Name:        "Login Harian",
			Description: "Login ke aplikasi hari ini",
			Icon:        "🎯",
			XPReward:    5,
			CoinReward:  1,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeChatAI,
			Name:        "Chat dengan AI",
			Description: "Kirim 3 pesan ke AI companion",
			Icon:        "💬",
			XPReward:    30,
			CoinReward:  5,
			TargetCount: 3,
		},
		{
			Type:        TaskTypeReadArticle,
			Name:        "Baca Artikel",
			Description: "Baca sebuah artikel",
			Icon:        "📖",
			XPReward:    20,
			CoinReward:  3,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeListenSongs,
			Name:        "Dengarkan Musik",
			Description: "Dengarkan 3 lagu relaksasi",
			Icon:        "🎵",
			XPReward:    15,
			CoinReward:  2,
			TargetCount: 3,
		},
		{
			Type:        TaskTypeWriteJournal,
			Name:        "Tulis Jurnal",
			Description: "Tulis sebuah jurnal",
			Icon:        "✍️",
			XPReward:    40,
			CoinReward:  6,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeCommentForum,
			Name:        "Komentar di Forum",
			Description: "Berikan komentar di forum",
			Icon:        "💭",
			XPReward:    15,
			CoinReward:  2,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeBreathing,
			Name:        "Latihan Pernafasan",
			Description: "Lakukan latihan pernafasan selesai",
			Icon:        "🌬️",
			XPReward:    25,
			CoinReward:  4,
			TargetCount: 1,
		},
	}
}

// GetPremiumDailyTaskConfigs returns extra premium-only missions.
func GetPremiumDailyTaskConfigs() []DailyTaskConfig {
	return []DailyTaskConfig{
		{
			Type:        TaskTypePremiumChatDeepDive,
			Name:        "Deep Chat Premium",
			Description: "Lanjutkan 6 pesan reflektif dengan AI",
			Icon:        "✨",
			XPReward:    55,
			CoinReward:  8,
			TargetCount: 6,
			PremiumOnly: true,
		},
		{
			Type:        TaskTypePremiumBreathingPro,
			Name:        "Breathing Pro",
			Description: "Selesaikan 2 sesi pernafasan fokus",
			Icon:        "🫧",
			XPReward:    45,
			CoinReward:  7,
			TargetCount: 2,
			PremiumOnly: true,
		},
	}
}

// GetDailyTaskConfigsByTier returns daily tasks based on user entitlement.
func GetDailyTaskConfigsByTier(includePremium bool) []DailyTaskConfig {
	configs := make([]DailyTaskConfig, 0, len(GetDailyTaskConfigs())+len(GetPremiumDailyTaskConfigs()))
	configs = append(configs, GetDailyTaskConfigs()...)
	if includePremium {
		configs = append(configs, GetPremiumDailyTaskConfigs()...)
	}

	return configs
}

// GetTaskConfig returns the configuration for a specific task type
func GetTaskConfig(taskType DailyTaskType) *DailyTaskConfig {
	configs := GetDailyTaskConfigsByTier(true)
	for _, config := range configs {
		if config.Type == taskType {
			return &config
		}
	}
	return nil
}

// GetTotalPossibleXP returns the total possible XP from base daily tasks.
func GetTotalPossibleXP() int {
	total := 0
	for _, config := range GetDailyTaskConfigs() {
		total += config.XPReward
	}
	return total
}

// GetTotalPossibleCoins returns the total possible coins from base daily tasks.
func GetTotalPossibleCoins() int {
	total := 0
	for _, config := range GetDailyTaskConfigs() {
		total += config.CoinReward
	}
	return total
}

// DailyTask represents a user's daily task
type DailyTask struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	UserID       uint          `gorm:"not null" json:"user_id"`
	TaskType     DailyTaskType `gorm:"type:varchar(50);not null" json:"task_type"`
	TaskDate     time.Time     `gorm:"type:date;not null" json:"task_date"`
	TargetCount  int           `gorm:"default:1" json:"target_count"`
	CurrentCount int           `gorm:"default:0" json:"current_count"`
	IsCompleted  bool          `gorm:"default:false" json:"is_completed"`
	IsClaimed    bool          `gorm:"default:false" json:"is_claimed"`
	XPReward     int           `gorm:"default:0" json:"xp_reward"`
	CoinReward   int           `gorm:"default:0" json:"coin_reward"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	ClaimedAt    *time.Time    `json:"claimed_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	// Virtual fields (not in DB)
	TaskName        string `gorm:"-" json:"task_name"`
	TaskDescription string `gorm:"-" json:"task_description"`
	TaskIcon        string `gorm:"-" json:"task_icon"`
	PremiumOnly     bool   `gorm:"-" json:"premium_only"`
}

func (DailyTask) TableName() string {
	return "daily_tasks"
}

// PopulateTaskInfo fills in the virtual fields from config
func (d *DailyTask) PopulateTaskInfo() {
	config := GetTaskConfig(d.TaskType)
	if config != nil {
		d.TaskName = config.Name
		d.TaskDescription = config.Description
		d.TaskIcon = config.Icon
		d.PremiumOnly = config.PremiumOnly
	}
}

// Progress returns the progress percentage (0-100)
func (d *DailyTask) Progress() int {
	if d.TargetCount == 0 {
		return 0
	}
	progress := (d.CurrentCount * 100) / d.TargetCount
	if progress > 100 {
		return 100
	}
	return progress
}

// DailyTaskSummary provides a summary of daily tasks
type DailyTaskSummary struct {
	Date               time.Time   `json:"date"`
	TotalTasks         int         `json:"total_tasks"`
	CompletedTasks     int         `json:"completed_tasks"`
	ClaimedTasks       int         `json:"claimed_tasks"`
	TotalXPEarned      int         `json:"total_xp_earned"`
	TotalXPPossible    int         `json:"total_xp_possible"`
	TotalCoinsEarned   int         `json:"total_coins_earned"`
	TotalCoinsPossible int         `json:"total_coins_possible"`
	Tasks              []DailyTask `json:"tasks"`
	LoginStreak        int         `json:"login_streak"`
}
