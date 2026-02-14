package response

import (
	"math"
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/pkg/apperror"
	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/gin-gonic/gin"
)

// ============================
// Response Structs
// Aligned with API-CONTEXT.yml response_contract
// ============================

// SuccessResponse is the standard success API response.
// Shape: { data, meta (optional), requestId }
type SuccessResponse struct {
	Data      any    `json:"data,omitempty"`
	Meta      any    `json:"meta,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// MessageResponse is a success response with only a message (e.g., delete).
// Shape: { data: { message }, requestId }
type MessageResponse struct {
	Data      any    `json:"data"`
	RequestID string `json:"requestId,omitempty"`
}

// ErrorResponse is the standard error API response.
// Shape: { message, code, details (optional), requestId }
type ErrorResponse struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// getRequestID extracts requestId from gin.Context
func getRequestID(c *gin.Context) string {
	return ctxutil.GetRequestID(c)
}

// ============================
// Success Response Functions
// ============================

// OK sends a 200 OK response with data
func OK(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data any, message string) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Deleted sends a success response for deletion
func Deleted(c *gin.Context, message string) {
	if message == "" {
		message = "Deleted successfully"
	}
	c.JSON(http.StatusOK, MessageResponse{
		Data:      gin.H{"message": message},
		RequestID: getRequestID(c),
	})
}

// Updated sends a success response for update
func Updated(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// ============================
// Paginated Response Functions
// ============================

// Paginated sends a paginated response with meta
func Paginated(c *gin.Context, items any, page, limit int, total int64) {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: items,
		Meta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		RequestID: getRequestID(c),
	})
}

// PaginatedWithMeta sends a paginated response with additional metadata
func PaginatedWithMeta(c *gin.Context, items any, page, limit int, total int64, extra map[string]any) {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	response := map[string]any{
		"data": items,
		"meta": PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		"requestId": getRequestID(c),
	}

	// Merge extra fields into response
	for k, v := range extra {
		response[k] = v
	}

	c.JSON(http.StatusOK, response)
}

// ============================
// Error Response Functions
// ============================

// Error sends an error response based on AppError
// Shape: { message, code, details (optional), requestId }
func Error(c *gin.Context, err error) {
	appErr := apperror.FromError(err)

	c.JSON(appErr.GetHTTPStatus(), ErrorResponse{
		Message:   appErr.Message,
		Code:      string(appErr.Code),
		Details:   appErr.Details,
		RequestID: getRequestID(c),
	})
}

// ErrorWithMessage sends an error response with a custom message
func ErrorWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message:   message,
		Code:      string(apperror.CodeInternal),
		RequestID: getRequestID(c),
	})
}

// BadRequest sends a 400 Bad Request error response
func BadRequest(c *gin.Context, message string) {
	Error(c, apperror.BadRequest(message))
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Authentication required"
	}
	Error(c, apperror.Unauthorized(message))
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access denied"
	}
	Error(c, apperror.Forbidden(message))
}

// NotFound sends a 404 Not Found error response
func NotFound(c *gin.Context, resource string) {
	Error(c, apperror.NotFound(resource))
}

// InternalError sends a 500 Internal Server error response
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(c, apperror.Internal(message))
}

// ValidationError sends a validation error response with field details
func ValidationError(c *gin.Context, errors []apperror.FieldError) {
	Error(c, apperror.ValidationWithDetails("Validation failed", errors))
}

// ============================
// Abort Functions (stops middleware chain)
// ============================

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err error) {
	appErr := apperror.FromError(err)
	c.AbortWithStatusJSON(appErr.GetHTTPStatus(), ErrorResponse{
		Message:   appErr.Message,
		Code:      string(appErr.Code),
		RequestID: getRequestID(c),
	})
}

// AbortUnauthorized aborts with 401 Unauthorized
func AbortUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Authentication required"
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
		Message:   message,
		Code:      string(apperror.CodeUnauthorized),
		RequestID: getRequestID(c),
	})
}

// AbortForbidden aborts with 403 Forbidden
func AbortForbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access denied"
	}
	c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
		Message:   message,
		Code:      string(apperror.CodeForbidden),
		RequestID: getRequestID(c),
	})
}
