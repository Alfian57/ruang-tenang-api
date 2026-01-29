package gamification

type ActivityType string

const (
	ActivityChatAI        ActivityType = "chat_ai"
	ActivityUploadArticle ActivityType = "upload_article"
	ActivityForumComment  ActivityType = "forum_comment"
	ActivityBreathing     ActivityType = "breathing"
)

const (
	ExpChatAI        int64 = 10
	ExpUploadArticle int64 = 20
	ExpForumComment  int64 = 5
	ExpBreathing     int64 = 5
)

const (
	LimitChatAI       int = 1 // Per day
	LimitForumComment int = 5 // Per day
	LimitBreathing    int = 6 // Per day (30 XP cap / 5 XP = 6)
)
