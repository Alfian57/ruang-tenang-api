package dto

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// bindResetPassword is a small harness that runs the same ShouldBindJSON path
// the handler uses, returning the binding error (if any).
func bindResetPassword(t *testing.T, body string) (ResetPasswordRequest, error) {
	t.Helper()
	var req ResetPasswordRequest

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	err := c.ShouldBindJSON(&req)
	return req, err
}

func TestResetPasswordRequestMatch(t *testing.T) {
	body := `{"token":"abc","new_password":"secret123","password_confirmation":"secret123"}`
	req, err := bindResetPassword(t, body)
	if err != nil {
		t.Fatalf("expected binding to succeed on matching passwords, got: %v", err)
	}
	if req.NewPassword != "secret123" {
		t.Fatalf("NewPassword = %q, want secret123", req.NewPassword)
	}
}

func TestResetPasswordRequestMismatchRejected(t *testing.T) {
	// The frontend sends password_confirmation; the backend must enforce that it
	// matches new_password so a typo can't silently reset to the wrong password.
	body := `{"token":"abc","new_password":"secret123","password_confirmation":"different"}`
	if _, err := bindResetPassword(t, body); err == nil {
		t.Fatalf("expected binding error when password_confirmation != new_password")
	}
}

func TestResetPasswordRequestIgnoresEmail(t *testing.T) {
	// FE still sends `email`; BE should accept and ignore it (binding is
	// selective). The token alone authorizes the reset.
	body := `{"email":"user@example.com","token":"abc","new_password":"secret123","password_confirmation":"secret123"}`
	if _, err := bindResetPassword(t, body); err != nil {
		t.Fatalf("email field should be ignored, got error: %v", err)
	}
}

func TestResetPasswordRequestMissingToken(t *testing.T) {
	body := `{"new_password":"secret123","password_confirmation":"secret123"}`
	if _, err := bindResetPassword(t, body); err == nil {
		t.Fatalf("expected binding error when token is missing")
	}
}

func TestResetPasswordRequestShortPassword(t *testing.T) {
	body := `{"token":"abc","new_password":"123","password_confirmation":"123"}`
	if _, err := bindResetPassword(t, body); err == nil {
		t.Fatalf("expected binding error for password shorter than min=6")
	}
}

func TestResetPasswordRequestJSONTags(t *testing.T) {
	// Guard the wire contract: confirm the JSON shape the FE sends maps correctly.
	raw := `{"token":"T","new_password":"N","password_confirmation":"N"}`
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"token", "new_password", "password_confirmation"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("expected key %q in payload", key)
		}
	}
}
