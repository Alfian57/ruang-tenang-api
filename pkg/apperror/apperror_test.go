package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestConstructorsAndStatus(t *testing.T) {
	err := BadRequest("bad input")
	if err.Code != CodeBadRequest || err.GetHTTPStatus() != http.StatusBadRequest {
		t.Fatalf("unexpected bad request error: %+v", err)
	}

	if Unauthorized("").Message == "" || Forbidden("").Message == "" {
		t.Fatal("expected default messages for unauthorized/forbidden")
	}

	if NotFound("Journal").Message != "Journal not found" {
		t.Fatalf("unexpected not found message: %s", NotFound("Journal").Message)
	}
}

func TestWrapAndFromError(t *testing.T) {
	base := errors.New("db timeout")
	wrapped := Wrap(base, Internal("failed"))

	if wrapped.Unwrap() != base {
		t.Fatal("expected unwrapped error to be base")
	}

	if got, ok := IsAppError(wrapped); !ok || got == nil {
		t.Fatal("expected wrapped to be AppError")
	}

	converted := FromError(base)
	if converted.Code != CodeInternal || converted.Err == nil {
		t.Fatalf("unexpected converted error: %+v", converted)
	}

	if FromError(nil) != nil {
		t.Fatal("expected nil from nil error")
	}
}

func TestAdditionalConstructorsAndHelpers(t *testing.T) {
	v := Validation("invalid field")
	if v.Code != CodeValidation || v.GetHTTPStatus() != http.StatusBadRequest {
		t.Fatalf("unexpected validation error: %+v", v)
	}

	details := []FieldError{NewFieldError("email", "required")}
	vd := ValidationWithDetails("invalid payload", details)
	if vd.Code != CodeValidation || vd.Details == nil {
		t.Fatalf("unexpected validation with details: %+v", vd)
	}

	if msg := ContentBlocked("").Message; msg == "" {
		t.Fatal("expected default content blocked message")
	}
	if msg := UserBlocked("").Message; msg == "" {
		t.Fatal("expected default user blocked message")
	}

	custom := New(CodeConflict, "conflict", http.StatusConflict).WithDetails(map[string]string{"k": "v"})
	if custom.Code != CodeConflict || custom.GetHTTPStatus() != http.StatusConflict {
		t.Fatalf("unexpected New() app error: %+v", custom)
	}

	root := errors.New("root cause")
	wrapped := custom.WithError(root)
	if wrapped.Unwrap() != root {
		t.Fatal("expected wrapped root error")
	}
	if wrapped.Error() == "" {
		t.Fatal("expected non-empty Error() string")
	}

	if got := FromError(BadRequest("bad")); got.Code != CodeBadRequest {
		t.Fatalf("expected FromError(AppError) to preserve code, got %+v", got)
	}
}

func TestErrorStringWithoutCauseAndDefaultHTTPStatus(t *testing.T) {
	err := &AppError{Code: CodeInternal, Message: "boom"}
	if err.GetHTTPStatus() != http.StatusInternalServerError {
		t.Fatalf("expected default 500 for zero status code, got %d", err.GetHTTPStatus())
	}

	if got := err.Error(); got != "INTERNAL_ERROR: boom" {
		t.Fatalf("unexpected Error() without cause: %s", got)
	}
}
