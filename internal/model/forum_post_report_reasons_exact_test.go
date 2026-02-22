package model

import "testing"

func TestValidPostReportReasons_ContainsExpectedSet(t *testing.T) {
	reasons := ValidPostReportReasons()
	if len(reasons) != 7 {
		t.Fatalf("expected exactly 7 reasons, got %d", len(reasons))
	}

	expected := map[ForumPostReportReason]bool{
		PostReportReasonSpam:           true,
		PostReportReasonHarassment:     true,
		PostReportReasonMisinformation: true,
		PostReportReasonSelfHarm:       true,
		PostReportReasonOffTopic:       true,
		PostReportReasonRude:           true,
		PostReportReasonOther:          true,
	}

	for _, reason := range reasons {
		if !expected[reason] {
			t.Fatalf("unexpected reason in list: %s", reason)
		}
	}
}

func TestIsValidPostReportReason_IsCaseSensitive(t *testing.T) {
	if IsValidPostReportReason("SPAM") {
		t.Fatal("expected uppercase SPAM to be invalid because validator is case-sensitive")
	}
	if !IsValidPostReportReason("spam") {
		t.Fatal("expected lowercase spam to be valid")
	}
}
