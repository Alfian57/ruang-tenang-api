package service

import (
	"errors"
	"testing"
)

func TestServiceErrorHelpers(t *testing.T) {
	se := &ServiceError{Code: "TEST", Message: "something happened"}
	if se.Error() != "something happened" {
		t.Fatalf("unexpected error string: %s", se.Error())
	}

	if got, ok := IsServiceError(se); !ok || got.Code != "TEST" {
		t.Fatalf("expected service error detection, got %+v, %v", got, ok)
	}

	if got, ok := IsServiceError(errors.New("plain")); ok || got != nil {
		t.Fatalf("expected non-service error, got %+v, %v", got, ok)
	}
}
