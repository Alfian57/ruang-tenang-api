package dto

// ErrorCode represents standardized error codes
type ErrorCode string

// Standard error codes
const (
	ErrCodeValidation     ErrorCode = "ERR_VALIDATION"
	ErrCodeUnauthorized   ErrorCode = "ERR_UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "ERR_FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "ERR_NOT_FOUND"
	ErrCodeConflict       ErrorCode = "ERR_CONFLICT"
	ErrCodeRateLimit      ErrorCode = "ERR_RATE_LIMIT"
	ErrCodeQuotaExceeded  ErrorCode = "ERR_QUOTA_EXCEEDED"
	ErrCodeInternal       ErrorCode = "ERR_INTERNAL"
	ErrCodeBadRequest     ErrorCode = "ERR_BAD_REQUEST"
	ErrCodeContentBlocked ErrorCode = "ERR_CONTENT_BLOCKED"
	ErrCodeUserBlocked    ErrorCode = "ERR_USER_BLOCKED"
)

// Common response wrapper
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// ValidationError represents a field-specific validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse contains multiple validation errors
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Code    string            `json:"code"`
	Error   string            `json:"error"`
	Details []ValidationError `json:"details,omitempty"`
}

// Pagination response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// Helper functions
func SuccessResponse(data interface{}, message string) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(err string) Response {
	return Response{
		Success: false,
		Error:   err,
	}
}

// ErrorResponseWithCode returns an error response with a specific error code
func ErrorResponseWithCode(code ErrorCode, err string) Response {
	return Response{
		Success: false,
		Code:    string(code),
		Error:   err,
	}
}

// NewValidationErrorResponse creates a validation error response with field details
func NewValidationErrorResponse(errors []ValidationError) ValidationErrorResponse {
	return ValidationErrorResponse{
		Success: false,
		Code:    string(ErrCodeValidation),
		Error:   "Validation failed",
		Details: errors,
	}
}

// NewSingleValidationError creates a validation error for a single field
func NewSingleValidationError(field, message string) ValidationErrorResponse {
	return NewValidationErrorResponse([]ValidationError{
		{Field: field, Message: message},
	})
}

func NewPaginatedResponse(data interface{}, page, limit int, total int64) PaginatedResponse {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return PaginatedResponse{
		Success:    true,
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}
}
