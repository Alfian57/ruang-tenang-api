package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFirstIP(t *testing.T) {
	tests := []struct {
		name string
		xff  string
		want string
	}{
		{name: "single ip", xff: "203.0.113.5", want: "203.0.113.5"},
		{name: "chain left-most is client", xff: "203.0.113.5, 10.0.0.1, 10.0.0.2", want: "203.0.113.5"},
		{name: "trimmed whitespace", xff: " 203.0.113.5 , 10.0.0.1", want: "203.0.113.5"},
		{name: "empty returns empty", xff: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstIP(tt.xff); got != tt.want {
				t.Fatalf("firstIP(%q) = %q, want %q", tt.xff, got, tt.want)
			}
		})
	}
}

func TestGetClientIPLeftMostFromForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1, 10.0.0.2")

	if got := getClientIP(c); got != "203.0.113.9" {
		t.Fatalf("getClientIP = %q, want left-most client 203.0.113.9", got)
	}
}

// TestGetClientIPNoSpoofBypass ensures a client cannot present a unique
// X-Forwarded-For per request to get a fresh rate-limit bucket via the whole
// header. Only the left-most IP should key the bucket.
func TestGetClientIPNoSpoofBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	// Attacker appends a unique IP to bypass; we must take only the first.
	c.Request.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.99")

	got := getClientIP(c)
	if got != "198.51.100.7" {
		t.Fatalf("getClientIP = %q, want 198.51.100.7 (unique suffix must not be used)", got)
	}
}
