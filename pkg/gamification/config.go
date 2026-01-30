package gamification

type ActivityType string

const (
	ActivityChatAI            ActivityType = "chat_ai"
	ActivityUploadArticle     ActivityType = "upload_article"
	ActivityForumComment      ActivityType = "forum_comment"
	ActivityBreathing         ActivityType = "breathing"
	ActivityAcceptedAnswer    ActivityType = "accepted_answer"     // OP marks user's answer as accepted
	ActivityPostUpvoteGiven   ActivityType = "post_upvote_given"   // User receives an upvote
	ActivityPostUpvoteRemoved ActivityType = "post_upvote_removed" // User loses an upvote (negative)
)

const (
	ExpChatAI            int64 = 10
	ExpUploadArticle     int64 = 20
	ExpForumComment      int64 = 5
	ExpBreathing         int64 = 5
	ExpAcceptedAnswer    int64 = 30 // Bonus for accepted answer
	ExpPostUpvoteGiven   int64 = 5  // Per upvote received
	ExpPostUpvoteRemoved int64 = -5 // When upvote is removed (can be negative)
)

const (
	LimitChatAI          int = 1  // Per day
	LimitForumComment    int = 5  // Per day
	LimitBreathing       int = 6  // Per day (30 XP cap / 5 XP = 6)
	LimitPostUpvoteGiven int = 20 // Max upvotes earning XP per day
)
