package utils

import (
	"errors"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	original := randReadFn
	t.Cleanup(func() { randReadFn = original })

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

	zero, err := GenerateRandomString(0)
	if err != nil {
		t.Fatalf("unexpected zero-length error: %v", err)
	}
	if zero != "" {
		t.Fatalf("expected empty string for zero length, got %q", zero)
	}
}

func TestGenerateRandomString_ReadError(t *testing.T) {
	original := randReadFn
	t.Cleanup(func() { randReadFn = original })

	randReadFn = func(p []byte) (int, error) {
		return 0, errors.New("read failed")
	}

	got, err := GenerateRandomString(4)
	if err == nil {
		t.Fatal("expected error")
	}
	if got != "" {
		t.Fatalf("expected empty output on error, got %q", got)
	}
}
