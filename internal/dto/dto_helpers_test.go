package dto

import "testing"

func TestResponseHelpers(t *testing.T) {
	s := SuccessResponse(map[string]any{"ok": true}, "done")
	if !s.Success || s.Message != "done" {
		t.Fatalf("unexpected success response: %+v", s)
	}

	e := ErrorResponse("failed")
	if e.Success || e.Error != "failed" {
		t.Fatalf("unexpected error response: %+v", e)
	}

	ec := ErrorResponseWithCode(ErrCodeForbidden, "forbidden")
	if ec.Code != string(ErrCodeForbidden) {
		t.Fatalf("unexpected code response: %+v", ec)
	}

	ve := NewValidationErrorResponse([]ValidationError{{Field: "name", Message: "required"}})
	if ve.Success || ve.Code != string(ErrCodeValidation) || len(ve.Details) != 1 {
		t.Fatalf("unexpected validation response: %+v", ve)
	}

	single := NewSingleValidationError("email", "invalid")
	if len(single.Details) != 1 || single.Details[0].Field != "email" {
		t.Fatalf("unexpected single validation response: %+v", single)
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	r := NewPaginatedResponse([]int{1, 2}, 1, 2, 5)
	if !r.Success || r.TotalPages != 3 || r.TotalItems != 5 {
		t.Fatalf("unexpected paginated response: %+v", r)
	}
}

func TestForumDTOHelpers(t *testing.T) {
	opts := GetForumPostSortOptions()
	if opts.Top != "top" || opts.Newest != "newest" || opts.Oldest != "oldest" {
		t.Fatalf("unexpected sort options: %+v", opts)
	}

	reasons := PostReportReasons()
	if len(reasons) < 3 {
		t.Fatalf("expected report reasons")
	}
	foundSpam := false
	for _, reason := range reasons {
		if reason["value"] == "spam" {
			foundSpam = true
			break
		}
	}
	if !foundSpam {
		t.Fatalf("expected spam reason in list")
	}
}
