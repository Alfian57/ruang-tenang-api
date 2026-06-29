package middleware

import (
	"github.com/Alfian57/ruang-tenang-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// AccountStatus describes whether an authenticated user is allowed to use the
// API right now. It is resolved per-request so that bans/blocks/suspensions
// take effect immediately, not only on the next login.
type AccountStatus struct {
	Allowed bool
	// Reason is a user-facing message (Indonesian) explaining the denial.
	Reason string
}

// AccountStatusResolver returns the current access status for a user ID.
// Implementations should be cheap (cached) since this runs on every
// authenticated request.
type AccountStatusResolver func(userID uint) AccountStatus

// accountStatusResolver is injected once during startup. When nil, the guard
// is a no-op (fails open) so the API keeps working if wiring is missing.
var accountStatusResolver AccountStatusResolver

// SetAccountStatusResolver wires the per-request account status check.
func SetAccountStatusResolver(resolver AccountStatusResolver) {
	accountStatusResolver = resolver
}

// accountStatusInvalidator clears any cached status for a user so moderation
// changes (ban/unban/suspend) apply on the next request immediately.
var accountStatusInvalidator func(userID uint)

// SetAccountStatusInvalidator wires the cache invalidation hook.
func SetAccountStatusInvalidator(invalidator func(userID uint)) {
	accountStatusInvalidator = invalidator
}

// InvalidateAccountStatus clears the cached status for a user. Safe to call
// from anywhere (no-op when not wired).
func InvalidateAccountStatus(userID uint) {
	if accountStatusInvalidator != nil {
		accountStatusInvalidator(userID)
	}
}

// enforceAccountStatus checks the resolver for the user already set in context.
// Returns false (and aborts the request) when access is denied.
func enforceAccountStatus(c *gin.Context) bool {
	if accountStatusResolver == nil {
		return true
	}

	userID, ok := GetUserID(c)
	if !ok {
		return true
	}

	status := accountStatusResolver(userID)
	if !status.Allowed {
		reason := status.Reason
		if reason == "" {
			reason = "Akun Anda sedang dibatasi. Silakan hubungi administrator."
		}
		response.AbortForbidden(c, reason)
		return false
	}
	return true
}
