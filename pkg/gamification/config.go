package gamification

type ActivityType string

const (
	ActivityChatAI        ActivityType = "chat_ai"
	ActivityUploadArticle ActivityType = "upload_article"
	ActivityForumComment  ActivityType = "forum_comment"
)

const (
	ExpChatAI        int64 = 10
	ExpUploadArticle int64 = 20
	ExpForumComment  int64 = 5
)

const (
	LimitChatAI       int = 1 // Per day
	LimitForumComment int = 5 // Per day
)
