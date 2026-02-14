package model

import (
	"time"

	"github.com/google/uuid"
)

// StoryCategory represents a category for inspiring stories
type StoryCategory struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Slug         string    `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	Description  string    `gorm:"type:text" json:"description"`
	Icon         string    `gorm:"size:50;default:'📖'" json:"icon"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (StoryCategory) TableName() string {
	return "story_categories"
}

// StoryStatus represents the moderation status of an inspiring story
type StoryStatus string

const (
	StoryStatusPending           StoryStatus = "pending"
	StoryStatusApproved          StoryStatus = "approved"
	StoryStatusRejected          StoryStatus = "rejected"
	StoryStatusRevisionRequested StoryStatus = "revision_requested"
)

// InspiringStory represents a user's inspiring mental health story
type InspiringStory struct {
	ID                 uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuthorID           uint        `gorm:"not null" json:"author_id"`
	Title              string      `gorm:"size:200;not null" json:"title"`
	Content            string      `gorm:"type:text;not null" json:"content"`
	CoverImage         string      `gorm:"size:500" json:"cover_image"`
	IsAnonymous        bool        `gorm:"default:false" json:"is_anonymous"`
	HasTriggerWarning  bool        `gorm:"default:false" json:"has_trigger_warning"`
	TriggerWarningText string      `gorm:"size:500" json:"trigger_warning_text"`
	Status             StoryStatus `gorm:"size:20;default:'pending'" json:"status"`
	ModeratorID        *uint       `json:"moderator_id,omitempty"`
	ModerationFeedback string      `gorm:"type:text" json:"moderation_feedback,omitempty"`
	ModeratedAt        *time.Time  `json:"moderated_at,omitempty"`
	ViewCount          int         `gorm:"default:0" json:"view_count"`
	HeartCount         int         `gorm:"default:0" json:"heart_count"`
	CommentCount       int         `gorm:"default:0" json:"comment_count"`
	IsFeatured         bool        `gorm:"default:false" json:"is_featured"`
	FeaturedAt         *time.Time  `json:"featured_at,omitempty"`
	FeaturedUntil      *time.Time  `json:"featured_until,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	PublishedAt        *time.Time  `json:"published_at,omitempty"`

	// Relations
	Author     *User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Moderator  *User           `gorm:"foreignKey:ModeratorID" json:"moderator,omitempty"`
	Categories []StoryCategory `gorm:"many2many:story_category_relations" json:"categories,omitempty"`
	Tags       []StoryTag      `gorm:"foreignKey:StoryID" json:"tags,omitempty"`
	Hearts     []StoryHeart    `gorm:"foreignKey:StoryID" json:"hearts,omitempty"`
	Comments   []StoryComment  `gorm:"foreignKey:StoryID" json:"comments,omitempty"`
}

func (InspiringStory) TableName() string {
	return "inspiring_stories"
}

// StoryTag represents a custom tag for an inspiring story
type StoryTag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoryID   uuid.UUID `gorm:"type:uuid;not null" json:"story_id"`
	Tag       string    `gorm:"size:50;not null" json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

func (StoryTag) TableName() string {
	return "story_tags"
}

// StoryHeart represents a heart (appreciation) given to a story
type StoryHeart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoryID   uuid.UUID `gorm:"type:uuid;not null" json:"story_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (StoryHeart) TableName() string {
	return "story_hearts"
}

// StoryComment represents a supportive comment on an inspiring story
type StoryComment struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoryID      uuid.UUID `gorm:"type:uuid;not null" json:"story_id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	Content      string    `gorm:"size:500;not null" json:"content"`
	HeartCount   int       `gorm:"default:0" json:"heart_count"`
	IsHidden     bool      `gorm:"default:false" json:"is_hidden"`
	HiddenReason string    `gorm:"size:255" json:"hidden_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	User   *User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Hearts []StoryCommentHeart `gorm:"foreignKey:CommentID" json:"hearts,omitempty"`
}

func (StoryComment) TableName() string {
	return "story_comments"
}

// StoryCommentHeart represents a heart given to a comment
type StoryCommentHeart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CommentID uuid.UUID `gorm:"type:uuid;not null" json:"comment_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (StoryCommentHeart) TableName() string {
	return "story_comment_hearts"
}

// StoryCategoryRelation is the join table for story-category many-to-many
type StoryCategoryRelation struct {
	StoryID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"story_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id"`
}

func (StoryCategoryRelation) TableName() string {
	return "story_category_relations"
}
