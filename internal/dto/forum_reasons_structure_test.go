package dto

import "testing"

func TestPostReportReasons_HaveValueAndLabelKeys(t *testing.T) {
	reasons := PostReportReasons()
	if len(reasons) != 7 {
		t.Fatalf("expected exactly 7 reasons, got %d", len(reasons))
	}

	for _, reason := range reasons {
		if reason["value"] == "" {
			t.Fatalf("expected non-empty reason value: %+v", reason)
		}
		if reason["label"] == "" {
			t.Fatalf("expected non-empty reason label: %+v", reason)
		}
	}
}
