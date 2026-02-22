package dto

import "testing"

func TestNewPaginatedResponse_ExactDivision(t *testing.T) {
	resp := NewPaginatedResponse([]int{1, 2, 3}, 2, 5, 10)
	if resp.TotalPages != 2 {
		t.Fatalf("expected 2 total pages, got %d", resp.TotalPages)
	}
	if resp.Page != 2 || resp.Limit != 5 || resp.TotalItems != 10 {
		t.Fatalf("unexpected pagination metadata: %+v", resp)
	}
}

func TestErrorResponseWithCode_SetsStringCode(t *testing.T) {
	resp := ErrorResponseWithCode(ErrCodeRateLimit, "too many requests")
	if resp.Success {
		t.Fatal("expected error response success=false")
	}
	if resp.Code != string(ErrCodeRateLimit) {
		t.Fatalf("expected code %s, got %s", ErrCodeRateLimit, resp.Code)
	}
	if resp.Error != "too many requests" {
		t.Fatalf("unexpected error message: %s", resp.Error)
	}
}
