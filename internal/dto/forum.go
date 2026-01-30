package dto

// VoteRequest represents a vote request (for future downvote with reason)
type VoteRequest struct {
	Reason string `json:"reason,omitempty"` // Only needed for downvotes (future)
}

// VoteResponse represents a vote response
type VoteResponse struct {
	Message  string `json:"message"`
	Voted    bool   `json:"voted"`
	VoteType string `json:"vote_type,omitempty"`
	NetVotes int    `json:"net_votes,omitempty"`
}

// AcceptAnswerResponse represents the response when marking an answer
type AcceptAnswerResponse struct {
	Message    string `json:"message"`
	IsAccepted bool   `json:"is_accepted"`
	XPAwarded  int64  `json:"xp_awarded,omitempty"`
}

// PostReportRequest represents a request to report a post
type PostReportRequest struct {
	Reason      string `json:"reason" binding:"required"` // spam, harassment, misinformation, self_harm, off_topic, rude, other
	Description string `json:"description,omitempty"`
}

// PostReportResponse represents the response after reporting a post
type PostReportResponse struct {
	Message  string `json:"message"`
	ReportID uint   `json:"report_id,omitempty"`
}

// ReviewReportRequest represents a request to review a post report
type ReviewReportRequest struct {
	Status string `json:"status" binding:"required"` // reviewed, dismissed, actioned
	Notes  string `json:"notes,omitempty"`
}

// ForumPostSortOptions represents the available sorting options
type ForumPostSortOptions struct {
	Top    string `json:"top"`
	Newest string `json:"newest"`
	Oldest string `json:"oldest"`
}

// GetForumPostSortOptions returns the available sort options
func GetForumPostSortOptions() ForumPostSortOptions {
	return ForumPostSortOptions{
		Top:    "top",
		Newest: "newest",
		Oldest: "oldest",
	}
}

// ForumPostWithVoteInfo extends ForumPost with user-specific vote information
type ForumPostWithVoteInfo struct {
	ID                  uint          `json:"id"`
	ForumID             uint          `json:"forum_id"`
	UserID              uint          `json:"user_id"`
	Content             string        `json:"content"`
	IsAcceptedAnswer    bool          `json:"is_accepted_answer"`
	IsCommunityFavorite bool          `json:"is_community_favorite"`
	UpvotesCount        int           `json:"upvotes_count"`
	DownvotesCount      int           `json:"downvotes_count"`
	NetVotes            int           `json:"net_votes"`
	HasUserVoted        bool          `json:"has_user_voted"`
	UserVoteType        string        `json:"user_vote_type,omitempty"`
	IsAutoHidden        bool          `json:"is_auto_hidden"`
	IsFlagged           bool          `json:"is_flagged"`
	CreatedAt           string        `json:"created_at"`
	UpdatedAt           string        `json:"updated_at"`
	User                *UserBriefDTO `json:"user,omitempty"`
}

// UserBriefDTO represents a brief user info for forum posts
type UserBriefDTO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Exp    int64  `json:"exp"`
}

// PostReportReasons returns valid report reasons for frontend
func PostReportReasons() []map[string]string {
	return []map[string]string{
		{"value": "spam", "label": "Spam"},
		{"value": "harassment", "label": "Harassment / Bullying"},
		{"value": "misinformation", "label": "Misinformation"},
		{"value": "self_harm", "label": "Self-harm / Dangerous Content"},
		{"value": "off_topic", "label": "Off-topic / Not Relevant"},
		{"value": "rude", "label": "Rude / Disrespectful"},
		{"value": "other", "label": "Other"},
	}
}
