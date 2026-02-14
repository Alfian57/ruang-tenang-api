package middleware

import (
	"strings"

	"github.com/Alfian57/ruang-tenang-api/pkg/response"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortUnauthorized(c, "Authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.AbortUnauthorized(c, "Invalid authorization header format")
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			response.AbortUnauthorized(c, "Invalid or expired token")
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			response.AbortForbidden(c, "Admin access required")
			return
		}
		c.Next()
	}
}

// MemberMiddleware checks if user has member role
func MemberMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "member" {
			response.AbortForbidden(c, "Member access required")
			return
		}
		c.Next()
	}
}

// ModeratorMiddleware checks if user has moderator or admin role
func ModeratorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			response.AbortForbidden(c, "Moderator access required")
			return
		}
		roleStr, ok := role.(string)
		if !ok || (roleStr != "moderator" && roleStr != "admin") {
			response.AbortForbidden(c, "Moderator access required")
			return
		}
		c.Next()
	}
}

// ModeratorOrAdminMiddleware is an alias for ModeratorMiddleware
// (both moderators and admins can access moderation features)
var ModeratorOrAdminMiddleware = ModeratorMiddleware

// GetUserID helper to get user ID from context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// GetUserRole helper to get user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	if !ok {
		return "", false
	}
	return roleStr, true
}
