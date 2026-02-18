package model

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/pkg/slug"
	"gorm.io/gorm"
)

type Forum struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	UserID     uint   `gorm:"not null" json:"user_id"`
	CategoryID *uint  `json:"category_id"`
	Title      string `gorm:"size:255;not null" json:"title"`
	Slug       string `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	Content    string `gorm:"type:text" json:"content"`

	// Moderation fields
	TriggerWarnings TriggerWarnings `gorm:"type:json" json:"trigger_warnings,omitempty"`
	IsFlagged       bool            `gorm:"default:false" json:"is_flagged"`
	FlaggedReason   string          `gorm:"type:text" json:"flagged_reason,omitempty"`

	// Best answer tracking
	HasAcceptedAnswer bool `gorm:"default:false" json:"has_accepted_answer"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category     *ForumCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Posts        []ForumPost    `gorm:"foreignKey:ForumID" json:"posts,omitempty"`
	Likes        []ForumLike    `gorm:"foreignKey:ForumID" json:"likes,omitempty"`
	LikesCount   int64          `gorm:"-" json:"likes_count"`
	RepliesCount int64          `gorm:"-" json:"replies_count"`
	IsLiked      bool           `gorm:"-" json:"is_liked"`
}

func (Forum) TableName() string {
	return "forums"
}

// BeforeCreate generates a slug from Title if not already set
func (f *Forum) BeforeCreate(tx *gorm.DB) error {
	if f.Slug == "" {
		f.Slug = slug.GenerateUnique(f.Title)
	}
	return nil
}

// HasTriggerWarnings returns true if forum has any trigger warnings
func (f *Forum) HasTriggerWarnings() bool {
	return len(f.TriggerWarnings) > 0
}

// VoteType represents the type of vote (upvote/downvote)
type VoteType string

const (
	VoteTypeUpvote   VoteType = "upvote"
	VoteTypeDownvote VoteType = "downvote"
)

type ForumPost struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	ForumID uint   `gorm:"not null" json:"forum_id"`
	UserID  uint   `gorm:"not null" json:"user_id"`
	Content string `gorm:"type:text;not null" json:"content"`

	// Moderation fields
	IsFlagged     bool   `gorm:"default:false" json:"is_flagged"`
	FlaggedReason string `gorm:"type:text" json:"flagged_reason,omitempty"`

	// Best answer and voting fields
	IsAcceptedAnswer    bool `gorm:"default:false" json:"is_accepted_answer"`
	IsCommunityFavorite bool `gorm:"default:false" json:"is_community_favorite"`
	UpvotesCount        int  `gorm:"default:0" json:"upvotes_count"`
	DownvotesCount      int  `gorm:"default:0" json:"downvotes_count"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Forum Forum           `gorm:"foreignKey:ForumID" json:"forum,omitempty"`
	User  User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Votes []ForumPostVote `gorm:"foreignKey:PostID" json:"votes,omitempty"`

	// Computed fields (not in DB)
	HasUserVoted bool     `gorm:"-" json:"has_user_voted"`
	UserVoteType VoteType `gorm:"-" json:"user_vote_type,omitempty"`
	NetVotes     int      `gorm:"-" json:"net_votes"` // upvotes - downvotes
	ReportsCount int64    `gorm:"-" json:"reports_count,omitempty"`
	IsAutoHidden bool     `gorm:"-" json:"is_auto_hidden"` // True if net_votes < -5
}

func (ForumPost) TableName() string {
	return "forum_posts"
}

// CalculateNetVotes calculates the net vote score
func (p *ForumPost) CalculateNetVotes() int {
	return p.UpvotesCount - p.DownvotesCount
}

// ShouldAutoHide returns true if post should be auto-hidden due to low votes
func (p *ForumPost) ShouldAutoHide() bool {
	return p.CalculateNetVotes() < -5
}
