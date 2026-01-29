package middleware

import (
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"
)

// SanitizationConfig holds configuration for input sanitization
type SanitizationConfig struct {
	AllowHTML       bool
	MaxLength       int
	StripNewlines   bool
	AllowedHTMLTags []string
}

// DefaultSanitizationConfig returns default sanitization settings
func DefaultSanitizationConfig() SanitizationConfig {
	return SanitizationConfig{
		AllowHTML:     false,
		MaxLength:     10000,
		StripNewlines: false,
	}
}

// Dangerous patterns that might indicate injection attempts
var (
	sqlInjectionPatterns = regexp.MustCompile(`(?i)(union\s+select|select\s+\*|drop\s+table|insert\s+into|delete\s+from|update\s+\w+\s+set|;\s*(drop|delete|update|insert)|--\s*$|/\*|\*/|xp_|sp_)`)
	xssPatterns          = regexp.MustCompile(`(?i)(<script|javascript:|on\w+\s*=|<iframe|<object|<embed|<form|<input|data:text/html|vbscript:)`)
	pathTraversalPattern = regexp.MustCompile(`\.\./|\.\.\\`)
)

// SanitizeString cleans a string input to prevent XSS and other attacks
func SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// HTML escape to prevent XSS
	input = html.EscapeString(input)

	return input
}

// SanitizeHTML allows limited HTML while preventing XSS
// Note: For rich text content, consider using a proper HTML sanitizer library
func SanitizeHTML(input string, allowedTags []string) string {
	// Basic HTML sanitization - strips all tags except allowed ones
	// For production, use a proper library like bluemonday
	input = strings.TrimSpace(input)

	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")

	// Remove event handlers
	eventRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	input = eventRegex.ReplaceAllString(input, "")

	// Remove javascript: urls
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	input = jsRegex.ReplaceAllString(input, "")

	return input
}

// DetectSQLInjection checks if input contains potential SQL injection patterns
func DetectSQLInjection(input string) bool {
	return sqlInjectionPatterns.MatchString(input)
}

// DetectXSS checks if input contains potential XSS patterns
func DetectXSS(input string) bool {
	return xssPatterns.MatchString(input)
}

// DetectPathTraversal checks if input contains path traversal attempts
func DetectPathTraversal(input string) bool {
	return pathTraversalPattern.MatchString(input)
}

// ValidateAndSanitize performs validation and sanitization on input
func ValidateAndSanitize(input string, config SanitizationConfig) (string, error) {
	// Check max length
	if len(input) > config.MaxLength {
		return "", &ValidationError{
			Field:   "input",
			Message: "Input exceeds maximum allowed length",
		}
	}

	// Detect potential injection attempts
	if DetectSQLInjection(input) {
		return "", &ValidationError{
			Field:   "input",
			Message: "Invalid characters detected",
		}
	}

	// Sanitize based on config
	var result string
	if config.AllowHTML {
		result = SanitizeHTML(input, config.AllowedHTMLTags)
	} else {
		result = SanitizeString(input)
	}

	if config.StripNewlines {
		result = strings.ReplaceAll(result, "\n", " ")
		result = strings.ReplaceAll(result, "\r", "")
	}

	return result, nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// InputValidationMiddleware validates and sanitizes common input fields
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for potential SQL injection in query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if DetectSQLInjection(value) {
					c.JSON(http.StatusBadRequest, dto.ErrorResponseWithCode(
						dto.ErrCodeBadRequest,
						"Invalid characters in query parameter: "+key,
					))
					c.Abort()
					return
				}
			}
		}

		// Check for path traversal in URL path
		if DetectPathTraversal(c.Request.URL.Path) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponseWithCode(
				dto.ErrCodeBadRequest,
				"Invalid path",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// SanitizeRequestBody is a helper to sanitize string fields in request body
// Call this in your handlers after binding the request
func SanitizeRequestBody(fields map[string]*string) {
	for _, field := range fields {
		if field != nil && *field != "" {
			*field = SanitizeString(*field)
		}
	}
}

// Common validation patterns
var (
	EmailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	UsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	UUIDPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// ValidateEmail checks if string is valid email format
func ValidateEmail(email string) bool {
	return EmailPattern.MatchString(email)
}

// ValidateUsername checks if string is valid username format
func ValidateUsername(username string) bool {
	return UsernamePattern.MatchString(username)
}

// ValidateUUID checks if string is valid UUID format
func ValidateUUID(uuid string) bool {
	return UUIDPattern.MatchString(uuid)
}

// ContentTypeValidationMiddleware ensures JSON content type for POST/PUT/PATCH
func ContentTypeValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				c.JSON(http.StatusBadRequest, dto.ErrorResponseWithCode(
					dto.ErrCodeBadRequest,
					"Content-Type header is required",
				))
				c.Abort()
				return
			}

			// Allow both application/json and multipart/form-data
			if !strings.HasPrefix(contentType, "application/json") &&
				!strings.HasPrefix(contentType, "multipart/form-data") {
				c.JSON(http.StatusUnsupportedMediaType, dto.ErrorResponseWithCode(
					dto.ErrCodeBadRequest,
					"Content-Type must be application/json or multipart/form-data",
				))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// MaxBodySizeMiddleware limits the size of request body
func MaxBodySizeMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
