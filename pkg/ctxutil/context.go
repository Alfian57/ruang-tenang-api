package ctxutil

import (
	"github.com/gin-gonic/gin"
)

// Context keys
const (
	KeyUserID    = "user_id"
	KeyUserEmail = "user_email"
	KeyUserRole  = "user_role"
	KeyRequestID = "request_id"
)

// UserInfo contains authenticated user information from context
type UserInfo struct {
	ID    uint
	Email string
	Role  string
}

// GetUserID extracts user ID from Gin context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(KeyUserID)
	if !exists {
		return 0, false
	}
	if id, ok := userID.(uint); ok {
		return id, true
	}
	return 0, false
}

// MustGetUserID extracts user ID from context, panics if not found
// Use only in protected routes where auth middleware has already validated
func MustGetUserID(c *gin.Context) uint {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user_id not found in context - ensure auth middleware is applied")
	}
	return userID
}

// GetUserEmail extracts user email from Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(KeyUserEmail)
	if !exists {
		return "", false
	}
	if e, ok := email.(string); ok {
		return e, true
	}
	return "", false
}

// GetUserRole extracts user role from Gin context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(KeyUserRole)
	if !exists {
		return "", false
	}
	if r, ok := role.(string); ok {
		return r, true
	}
	return "", false
}

// GetUserInfo extracts all user information from context
func GetUserInfo(c *gin.Context) (*UserInfo, bool) {
	userID, ok := GetUserID(c)
	if !ok {
		return nil, false
	}

	email, _ := GetUserEmail(c)
	role, _ := GetUserRole(c)

	return &UserInfo{
		ID:    userID,
		Email: email,
		Role:  role,
	}, true
}

// MustGetUserInfo extracts user info, panics if not found
func MustGetUserInfo(c *gin.Context) *UserInfo {
	info, ok := GetUserInfo(c)
	if !ok {
		panic("user info not found in context - ensure auth middleware is applied")
	}
	return info
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	role, ok := GetUserRole(c)
	return ok && role == "admin"
}

// IsModerator checks if the current user has moderator or admin role
func IsModerator(c *gin.Context) bool {
	role, ok := GetUserRole(c)
	return ok && (role == "moderator" || role == "admin")
}

// IsMember checks if the current user has member role
func IsMember(c *gin.Context) bool {
	role, ok := GetUserRole(c)
	return ok && role == "member"
}

// IsOwnerOrAdmin checks if the user owns a resource or is admin
func IsOwnerOrAdmin(c *gin.Context, ownerID uint) bool {
	userID, ok := GetUserID(c)
	if !ok {
		return false
	}
	if userID == ownerID {
		return true
	}
	return IsAdmin(c)
}

// IsOwnerOrModerator checks if the user owns a resource or is moderator/admin
func IsOwnerOrModerator(c *gin.Context, ownerID uint) bool {
	userID, ok := GetUserID(c)
	if !ok {
		return false
	}
	if userID == ownerID {
		return true
	}
	return IsModerator(c)
}

// SetUserInfo sets user information in context (typically used by auth middleware)
func SetUserInfo(c *gin.Context, userID uint, email, role string) {
	c.Set(KeyUserID, userID)
	c.Set(KeyUserEmail, email)
	c.Set(KeyUserRole, role)
}

// SetRequestID sets the request ID in context
func SetRequestID(c *gin.Context, requestID string) {
	c.Set(KeyRequestID, requestID)
}

// GetRequestID gets the request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(KeyRequestID); exists {
		if rid, ok := id.(string); ok {
			return rid
		}
	}
	return ""
}
