package model

import (
	"testing"
	"time"
)

func TestUserSuspension_PastDateNotSuspended(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	u := &User{SuspensionEnd: &past}

	if u.IsSuspended() {
		t.Fatal("expected user with past suspension end to not be suspended")
	}
}

func TestUserCanAccess_FalseWhenBannedOrBlocked(t *testing.T) {
	banned := &User{IsBanned: true}
	if banned.CanAccess() {
		t.Fatal("expected banned user to not have access")
	}

	blocked := &User{IsBlocked: true}
	if blocked.CanAccess() {
		t.Fatal("expected blocked user to not have access")
	}
}
