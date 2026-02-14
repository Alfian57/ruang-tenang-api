package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

// Standard error codes
const (
	CodeValidation     ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	CodeForbidden      ErrorCode = "FORBIDDEN"
	CodeNotFound       ErrorCode = "NOT_FOUND"
	CodeConflict       ErrorCode = "CONFLICT"
	CodeRateLimit      ErrorCode = "ERR_RATE_LIMIT"
	CodeInternal       ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest     ErrorCode = "ERR_BAD_REQUEST"
	CodeContentBlocked ErrorCode = "ERR_CONTENT_BLOCKED"
	CodeUserBlocked    ErrorCode = "ERR_USER_BLOCKED"
)

// AppError represents an application-level error with HTTP status code
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	StatusCode int       `json:"-"`
	Details    any       `json:"details,omitempty"`
	Err        error     `json:"-"` // Original error (not exposed to client)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// IsAppError checks if an error is an AppError and returns it
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetHTTPStatus returns the HTTP status code for the error
func (e *AppError) GetHTTPStatus() int {
	if e.StatusCode != 0 {
		return e.StatusCode
	}
	return http.StatusInternalServerError
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// ============================
// Constructor Functions
// ============================

// New creates a new AppError
func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// Validation creates a 400 Validation error
func Validation(message string) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// ValidationWithDetails creates a validation error with field-specific details
func ValidationWithDetails(message string, details []FieldError) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    details,
	}
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NotFound creates a 404 Not Found error
func NotFound(resource string) *AppError {
	message := "Resource not found"
	if resource != "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return &AppError{
		Code:       CodeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// RateLimited creates a 429 Rate Limit error
func RateLimited(message string) *AppError {
	if message == "" {
		message = "Too many requests, please try again later"
	}
	return &AppError{
		Code:       CodeRateLimit,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

// Internal creates a 500 Internal Server error
func Internal(message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:       CodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// ContentBlocked creates a content blocked error
func ContentBlocked(message string) *AppError {
	if message == "" {
		message = "Content has been blocked due to policy violation"
	}
	return &AppError{
		Code:       CodeContentBlocked,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// UserBlocked creates a user blocked error
func UserBlocked(message string) *AppError {
	if message == "" {
		message = "User account has been blocked"
	}
	return &AppError{
		Code:       CodeUserBlocked,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// ============================
// Field Error for Validation
// ============================

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewFieldError creates a new field error
func NewFieldError(field, message string) FieldError {
	return FieldError{Field: field, Message: message}
}

// ============================
// Common Error Instances
// ============================

var (
	ErrUnauthorized       = Unauthorized("")
	ErrForbidden          = Forbidden("")
	ErrNotFound           = NotFound("")
	ErrInternalServer     = Internal("")
	ErrRateLimited        = RateLimited("")
	ErrEmailAlreadyExists = Conflict("Email already registered")
	ErrInvalidCredentials = Unauthorized("Invalid email or password")
	ErrTokenExpired       = Unauthorized("Token has expired")
	ErrInvalidToken       = Unauthorized("Invalid token")
)

// ============================
// Error Wrapping Utilities
// ============================

// Wrap wraps an error with an AppError
func Wrap(err error, appErr *AppError) *AppError {
	return appErr.WithError(err)
}

// FromError converts a standard error to AppError
// If it's already an AppError, returns it as-is
// Otherwise, wraps it as an Internal error
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := IsAppError(err); ok {
		return appErr
	}
	return Internal(err.Error()).WithError(err)
}
