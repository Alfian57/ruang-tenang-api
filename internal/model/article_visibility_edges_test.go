package model

import "testing"

func TestArticleVisibility_ForSystemArticleWithoutModerationApproval(t *testing.T) {
	article := &Article{
		Status:           ArticleStatusPublished,
		IsUserGenerated:  false,
		ModerationStatus: ArticleModerationPending,
	}

	if !article.IsPublic() {
		t.Fatal("expected non-user-generated published article to be public")
	}
}

func TestArticleVisibility_NotPublishedIsNeverPublic(t *testing.T) {
	article := &Article{
		Status:           ArticleStatusDraft,
		IsUserGenerated:  false,
		ModerationStatus: ArticleModerationApproved,
	}

	if article.IsPublic() {
		t.Fatal("expected draft article to not be public")
	}
}

func TestArticleNeedsModeration_OnlyFlaggedUGC(t *testing.T) {
	nonUGC := &Article{IsUserGenerated: false, ModerationStatus: ArticleModerationFlagged}
	if nonUGC.NeedsModeration() {
		t.Fatal("expected non-UGC article to not require moderation")
	}

	ugcPending := &Article{IsUserGenerated: true, ModerationStatus: ArticleModerationPending}
	if ugcPending.NeedsModeration() {
		t.Fatal("expected pending UGC article to not require flagged moderation flow")
	}
}
