package model

import "testing"

func TestForumPostShouldAutoHide_ThresholdBoundary(t *testing.T) {
	atBoundary := &ForumPost{UpvotesCount: 0, DownvotesCount: 5}
	if atBoundary.CalculateNetVotes() != -5 {
		t.Fatalf("expected net votes -5, got %d", atBoundary.CalculateNetVotes())
	}
	if atBoundary.ShouldAutoHide() {
		t.Fatal("expected post with net votes -5 to stay visible")
	}

	belowBoundary := &ForumPost{UpvotesCount: 1, DownvotesCount: 7}
	if belowBoundary.CalculateNetVotes() != -6 {
		t.Fatalf("expected net votes -6, got %d", belowBoundary.CalculateNetVotes())
	}
	if !belowBoundary.ShouldAutoHide() {
		t.Fatal("expected post with net votes below -5 to auto-hide")
	}
}
