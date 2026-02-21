package service

import (
	"testing"

	"github.com/Alfian57/ruang-tenang-api/pkg/gamification"
)

func TestGetDailyLimit(t *testing.T) {
	cases := []struct {
		activity gamification.ActivityType
		want     int
	}{
		{gamification.ActivityChatAI, gamification.LimitChatAI},
		{gamification.ActivityForumComment, gamification.LimitForumComment},
		{gamification.ActivityPostUpvoteGiven, gamification.LimitPostUpvoteGiven},
		{gamification.ActivityStoryApproved, gamification.LimitStoryApproved},
		{gamification.ActivityHeartReceived, gamification.LimitHeartReceived},
		{gamification.ActivityUploadArticle, 0},
	}

	for _, tc := range cases {
		if got := getDailyLimit(tc.activity); got != tc.want {
			t.Fatalf("activity %s: expected %d, got %d", tc.activity, tc.want, got)
		}
	}
}

func TestGetActivityDescription(t *testing.T) {
	cases := []struct {
		activity gamification.ActivityType
		contains string
	}{
		{gamification.ActivityChatAI, "chat"},
		{gamification.ActivityUploadArticle, "artikel"},
		{gamification.ActivityForumComment, "forum"},
		{gamification.ActivityAcceptedAnswer, "solusi"},
		{gamification.ActivityPostUpvoteGiven, "upvote"},
		{gamification.ActivityPostUpvoteRemoved, "Kehilangan"},
		{gamification.ActivityStoryApproved, "Cerita"},
		{gamification.ActivityHeartReceived, "heart"},
		{gamification.ActivityType("unknown"), "Aktivitas lainnya"},
	}

	for _, tc := range cases {
		got := getActivityDescription(tc.activity)
		if got == "" {
			t.Fatalf("activity %s must have description", tc.activity)
		}
	}
}
