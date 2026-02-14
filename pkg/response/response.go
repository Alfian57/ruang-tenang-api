package response

import (
	"math"
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/pkg/apperror"
	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/gin-gonic/gin"
)

// Response is the standard API response wrapper
// Aligned with API-CONTEXT.yml response_contract
type Response struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	Code      string `json:"code,omitempty"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// PaginatedData contains paginated data with metadata
type PaginatedData struct {
	Items      any   `json:"items"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResponse is the standard paginated API response
type PaginatedResponse struct {
	Success   bool          `json:"success"`
	Data      PaginatedData `json:"data"`
	RequestID string        `json:"requestId,omitempty"`
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
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data any, message string) {
	if message == "" {
		message = "Created successfully"
	}
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Message:   message,
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
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// Updated sends a success response for update
func Updated(c *gin.Context, data any, message string) {
	if message == "" {
		message = "Updated successfully"
	}
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// ============================
// Paginated Response Functions
// ============================

// Paginated sends a paginated response
func Paginated(c *gin.Context, items any, page, limit int, total int64) {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data: PaginatedData{
			Items:      items,
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
func PaginatedWithMeta(c *gin.Context, items any, page, limit int, total int64, meta map[string]any) {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	response := map[string]any{
		"success": true,
		"data": PaginatedData{
			Items:      items,
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		"requestId": getRequestID(c),
	}

	// Merge meta into response
	for k, v := range meta {
		response[k] = v
	}

	c.JSON(http.StatusOK, response)
}

// ============================
// Error Response Functions
// ============================

// Error sends an error response based on AppError
// Uses "message" field for error message (aligned with API-CONTEXT.yml error_envelope)
func Error(c *gin.Context, err error) {
	appErr := apperror.FromError(err)
	reqID := getRequestID(c)

	if appErr.Details != nil {
		c.JSON(appErr.GetHTTPStatus(), Response{
			Success:   false,
			Message:   appErr.Message,
			Code:      string(appErr.Code),
			Details:   appErr.Details,
			RequestID: reqID,
		})
		return
	}

	c.JSON(appErr.GetHTTPStatus(), Response{
		Success:   false,
		Message:   appErr.Message,
		Code:      string(appErr.Code),
		RequestID: reqID,
	})
}

// ErrorWithMessage sends an error response with a custom message
func ErrorWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success:   false,
		Message:   message,
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
	c.AbortWithStatusJSON(appErr.GetHTTPStatus(), Response{
		Success:   false,
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
	c.AbortWithStatusJSON(http.StatusUnauthorized, Response{
		Success:   false,
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
	c.AbortWithStatusJSON(http.StatusForbidden, Response{
		Success:   false,
		Message:   message,
		Code:      string(apperror.CodeForbidden),
		RequestID: getRequestID(c),
	})
}
