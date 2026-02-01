package models

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
)

// DailyTaskConfig holds configuration for each task type
type DailyTaskConfig struct {
	Type        DailyTaskType
	Name        string
	Description string
	Icon        string
	XPReward    int
	TargetCount int
}

// GetDailyTaskConfigs returns all task configurations
func GetDailyTaskConfigs() []DailyTaskConfig {
	return []DailyTaskConfig{
		{
			Type:        TaskTypeDailyLogin,
			Name:        "Login Harian",
			Description: "Login ke aplikasi hari ini",
			Icon:        "🎯",
			XPReward:    5,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeChatAI,
			Name:        "Chat dengan AI",
			Description: "Kirim 3 pesan ke AI companion",
			Icon:        "💬",
			XPReward:    30,
			TargetCount: 3,
		},
		{
			Type:        TaskTypeReadArticle,
			Name:        "Baca Artikel",
			Description: "Baca sebuah artikel",
			Icon:        "📖",
			XPReward:    20,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeListenSongs,
			Name:        "Dengarkan Musik",
			Description: "Dengarkan 3 lagu relaksasi",
			Icon:        "🎵",
			XPReward:    15,
			TargetCount: 3,
		},
		{
			Type:        TaskTypeWriteJournal,
			Name:        "Tulis Jurnal",
			Description: "Tulis sebuah jurnal",
			Icon:        "✍️",
			XPReward:    40,
			TargetCount: 1,
		},
		{
			Type:        TaskTypeCommentForum,
			Name:        "Komentar di Forum",
			Description: "Berikan komentar di forum",
			Icon:        "💭",
			XPReward:    15,
			TargetCount: 1,
		},
	}
}

// GetTaskConfig returns the configuration for a specific task type
func GetTaskConfig(taskType DailyTaskType) *DailyTaskConfig {
	configs := GetDailyTaskConfigs()
	for _, config := range configs {
		if config.Type == taskType {
			return &config
		}
	}
	return nil
}

// GetTotalPossibleXP returns the total possible XP from all daily tasks
func GetTotalPossibleXP() int {
	total := 0
	for _, config := range GetDailyTaskConfigs() {
		total += config.XPReward
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
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	ClaimedAt    *time.Time    `json:"claimed_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	// Virtual fields (not in DB)
	TaskName        string `gorm:"-" json:"task_name"`
	TaskDescription string `gorm:"-" json:"task_description"`
	TaskIcon        string `gorm:"-" json:"task_icon"`
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
	Date            time.Time   `json:"date"`
	TotalTasks      int         `json:"total_tasks"`
	CompletedTasks  int         `json:"completed_tasks"`
	ClaimedTasks    int         `json:"claimed_tasks"`
	TotalXPEarned   int         `json:"total_xp_earned"`
	TotalXPPossible int         `json:"total_xp_possible"`
	Tasks           []DailyTask `json:"tasks"`
	LoginStreak     int         `json:"login_streak"`
}
