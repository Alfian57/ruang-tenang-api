package ctxutil

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

func TestSetAndGetUserInfo(t *testing.T) {
	c := newGinContext()
	SetUserInfo(c, 7, "user@example.com", "moderator")

	if id, ok := GetUserID(c); !ok || id != 7 {
		t.Fatalf("unexpected user id: %v %v", id, ok)
	}
	if email, ok := GetUserEmail(c); !ok || email != "user@example.com" {
		t.Fatalf("unexpected email: %v %v", email, ok)
	}
	if role, ok := GetUserRole(c); !ok || role != "moderator" {
		t.Fatalf("unexpected role: %v %v", role, ok)
	}

	info, ok := GetUserInfo(c)
	if !ok || info.ID != 7 || info.Email != "user@example.com" || info.Role != "moderator" {
		t.Fatalf("unexpected user info: %+v", info)
	}
}

func TestRoleAndOwnershipChecks(t *testing.T) {
	c := newGinContext()
	SetUserInfo(c, 11, "admin@example.com", "admin")

	if !IsAdmin(c) || !IsModerator(c) || IsMember(c) {
		t.Fatal("unexpected role checks for admin")
	}
	if !IsOwnerOrAdmin(c, 99) {
		t.Fatal("admin should pass owner/admin check")
	}

	SetUserInfo(c, 12, "member@example.com", "member")
	if !IsMember(c) || IsAdmin(c) || IsModerator(c) {
		t.Fatal("unexpected role checks for member")
	}
	if !IsOwnerOrModerator(c, 12) {
		t.Fatal("owner should pass owner/moderator check")
	}
}

func TestMustGetUserIDPanics(t *testing.T) {
	c := newGinContext()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = MustGetUserID(c)
}

func TestRequestIDHelpers(t *testing.T) {
	c := newGinContext()
	if GetRequestID(c) != "" {
		t.Fatal("expected empty request id")
	}

	SetRequestID(c, "rid-123")
	if got := GetRequestID(c); got != "rid-123" {
		t.Fatalf("unexpected request id: %s", got)
	}
}

func TestGetters_MissingOrWrongTypes(t *testing.T) {
	c := newGinContext()

	if id, ok := GetUserID(c); ok || id != 0 {
		t.Fatalf("expected empty user id, got %v %v", id, ok)
	}
	if email, ok := GetUserEmail(c); ok || email != "" {
		t.Fatalf("expected empty email, got %q %v", email, ok)
	}
	if role, ok := GetUserRole(c); ok || role != "" {
		t.Fatalf("expected empty role, got %q %v", role, ok)
	}
	if info, ok := GetUserInfo(c); ok || info != nil {
		t.Fatalf("expected nil info, got %+v %v", info, ok)
	}

	c.Set(KeyUserID, "not-uint")
	c.Set(KeyUserEmail, 123)
	c.Set(KeyUserRole, 999)
	c.Set(KeyRequestID, 456)

	if id, ok := GetUserID(c); ok || id != 0 {
		t.Fatalf("expected invalid user id cast to fail, got %v %v", id, ok)
	}
	if email, ok := GetUserEmail(c); ok || email != "" {
		t.Fatalf("expected invalid email cast to fail, got %q %v", email, ok)
	}
	if role, ok := GetUserRole(c); ok || role != "" {
		t.Fatalf("expected invalid role cast to fail, got %q %v", role, ok)
	}
	if got := GetRequestID(c); got != "" {
		t.Fatalf("expected empty request id on invalid type, got %q", got)
	}
}

func TestMustGetUserInfoPanicsAndOwnershipWithoutUser(t *testing.T) {
	c := newGinContext()

	if IsOwnerOrAdmin(c, 1) {
		t.Fatal("expected owner/admin false when user missing")
	}
	if IsOwnerOrModerator(c, 1) {
		t.Fatal("expected owner/moderator false when user missing")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = MustGetUserInfo(c)
}
