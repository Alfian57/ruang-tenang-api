package model

import "time"

// ForumPostVote represents a user's vote on a forum post
type ForumPostVote struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null" json:"post_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	VoteType  VoteType  `gorm:"type:varchar(10);not null;default:'upvote'" json:"vote_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Post ForumPost `gorm:"foreignKey:PostID" json:"post,omitempty"`
	User User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ForumPostVote) TableName() string {
	return "forum_post_votes"
}

// IsUpvote returns true if this is an upvote
func (v *ForumPostVote) IsUpvote() bool {
	return v.VoteType == VoteTypeUpvote
}

// IsDownvote returns true if this is a downvote
func (v *ForumPostVote) IsDownvote() bool {
	return v.VoteType == VoteTypeDownvote
}
