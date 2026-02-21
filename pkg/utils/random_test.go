package utils

import "testing"

func TestGenerateRandomString(t *testing.T) {
	got, err := GenerateRandomString(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 16 {
		t.Fatalf("expected hex length 16, got %d", len(got))
	}

	got2, err := GenerateRandomString(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == got2 {
		t.Fatal("expected different random strings across calls")
	}
}
