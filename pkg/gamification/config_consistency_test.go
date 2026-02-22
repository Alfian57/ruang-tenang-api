package gamification

import "testing"

func TestActivityTypes_AreNonEmpty(t *testing.T) {
	activities := []ActivityType{
		ActivityChatAI,
		ActivityUploadArticle,
		ActivityForumComment,
		ActivityBreathing,
		ActivityAcceptedAnswer,
		ActivityPostUpvoteGiven,
		ActivityPostUpvoteRemoved,
		ActivityStoryApproved,
		ActivityHeartReceived,
	}

	seen := map[ActivityType]bool{}
	for _, activity := range activities {
		if activity == "" {
			t.Fatal("activity type should not be empty")
		}
		if seen[activity] {
			t.Fatalf("duplicate activity type found: %s", activity)
		}
		seen[activity] = true
	}
}

func TestExpAndLimits_BasicSanity(t *testing.T) {
	if ExpChatAI <= 0 || ExpUploadArticle <= 0 || ExpForumComment <= 0 || ExpBreathing <= 0 {
		t.Fatal("core XP rewards should be positive")
	}
	if ExpAcceptedAnswer <= 0 || ExpPostUpvoteGiven <= 0 || ExpStoryApproved <= 0 || ExpHeartReceived <= 0 {
		t.Fatal("bonus XP rewards should be positive")
	}
	if ExpPostUpvoteRemoved >= 0 {
		t.Fatal("upvote removed XP should be negative")
	}

	if LimitChatAI <= 0 || LimitForumComment <= 0 || LimitBreathing <= 0 || LimitPostUpvoteGiven <= 0 {
		t.Fatal("daily limits should be positive")
	}
	if LimitStoryApproved <= 0 || LimitHeartReceived <= 0 {
		t.Fatal("story/heart daily limits should be positive")
	}
}
