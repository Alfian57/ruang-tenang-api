package handlers

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/pkg/apperror"
	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/Alfian57/ruang-tenang-api/pkg/queryutil"
	"github.com/Alfian57/ruang-tenang-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// BaseHandler provides common handler utilities
type BaseHandler struct{}

// NewBaseHandler creates a new base handler
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// ============================
// Response Methods
// ============================

// OK sends a 200 OK response
func (h *BaseHandler) OK(c *gin.Context, data any, message string) {
	response.OK(c, data, message)
}

// Created sends a 201 Created response
func (h *BaseHandler) Created(c *gin.Context, data any, message string) {
	response.Created(c, data, message)
}

// NoContent sends a 204 No Content response
func (h *BaseHandler) NoContent(c *gin.Context) {
	response.NoContent(c)
}

// Deleted sends a success response for deletion
func (h *BaseHandler) Deleted(c *gin.Context, message string) {
	response.Deleted(c, message)
}

// Updated sends a success response for update
func (h *BaseHandler) Updated(c *gin.Context, data any, message string) {
	response.Updated(c, data, message)
}

// Paginated sends a paginated response
func (h *BaseHandler) Paginated(c *gin.Context, items any, page, limit int, total int64) {
	response.Paginated(c, items, page, limit, total)
}

// ============================
// Error Methods
// ============================

// Error sends an error response based on error type
func (h *BaseHandler) Error(c *gin.Context, err error) {
	response.Error(c, err)
}

// BadRequest sends a 400 error
func (h *BaseHandler) BadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

// Unauthorized sends a 401 error
func (h *BaseHandler) Unauthorized(c *gin.Context, message string) {
	response.Unauthorized(c, message)
}

// Forbidden sends a 403 error
func (h *BaseHandler) Forbidden(c *gin.Context, message string) {
	response.Forbidden(c, message)
}

// NotFound sends a 404 error
func (h *BaseHandler) NotFound(c *gin.Context, resource string) {
	response.NotFound(c, resource)
}

// InternalError sends a 500 error
func (h *BaseHandler) InternalError(c *gin.Context, message string) {
	response.InternalError(c, message)
}

// ValidationError sends validation errors
func (h *BaseHandler) ValidationError(c *gin.Context, errors []apperror.FieldError) {
	response.ValidationError(c, errors)
}

// ============================
// Context Methods
// ============================

// GetUserID gets the authenticated user ID
func (h *BaseHandler) GetUserID(c *gin.Context) (uint, bool) {
	return ctxutil.GetUserID(c)
}

// MustGetUserID gets the user ID or panics
func (h *BaseHandler) MustGetUserID(c *gin.Context) uint {
	return ctxutil.MustGetUserID(c)
}

// GetUserInfo gets the authenticated user info
func (h *BaseHandler) GetUserInfo(c *gin.Context) (*ctxutil.UserInfo, bool) {
	return ctxutil.GetUserInfo(c)
}

// IsAdmin checks if user is admin
func (h *BaseHandler) IsAdmin(c *gin.Context) bool {
	return ctxutil.IsAdmin(c)
}

// IsOwnerOrAdmin checks if user owns the resource or is admin
func (h *BaseHandler) IsOwnerOrAdmin(c *gin.Context, ownerID uint) bool {
	return ctxutil.IsOwnerOrAdmin(c, ownerID)
}

// ============================
// Query Parameter Methods
// ============================

// GetPagination extracts pagination params from request
func (h *BaseHandler) GetPagination(c *gin.Context) queryutil.PaginationParams {
	return queryutil.GetPagination(c)
}

// GetIntQuery gets an integer query param
func (h *BaseHandler) GetIntQuery(c *gin.Context, key string, defaultValue int) int {
	return queryutil.GetIntQuery(c, key, defaultValue)
}

// GetUintQuery gets a uint query param
func (h *BaseHandler) GetUintQuery(c *gin.Context, key string, defaultValue uint) uint {
	return queryutil.GetUintQuery(c, key, defaultValue)
}

// GetBoolQuery gets a boolean query param
func (h *BaseHandler) GetBoolQuery(c *gin.Context, key string, defaultValue bool) bool {
	return queryutil.GetBoolQuery(c, key, defaultValue)
}

// GetStringQuery gets a string query param
func (h *BaseHandler) GetStringQuery(c *gin.Context, key, defaultValue string) string {
	return queryutil.GetStringQuery(c, key, defaultValue)
}

// GetOptionalUint gets an optional uint query param
func (h *BaseHandler) GetOptionalUint(c *gin.Context, key string) *uint {
	return queryutil.GetOptionalUint(c, key)
}

// ============================
// Path Parameter Methods
// ============================

// GetUintParam gets a uint path parameter
func (h *BaseHandler) GetUintParam(c *gin.Context, key string) (uint, bool) {
	return queryutil.GetUintParam(c, key)
}

// MustGetUintParam gets a uint path param (returns 0 on error)
func (h *BaseHandler) MustGetUintParam(c *gin.Context, key string) uint {
	return queryutil.MustGetUintParam(c, key)
}

// GetIntParam gets an int path parameter
func (h *BaseHandler) GetIntParam(c *gin.Context, key string) (int, bool) {
	return queryutil.GetIntParam(c, key)
}

// ============================
// Binding Helpers
// ============================

// BindJSON binds JSON request body and sends error response if invalid
// Returns true if binding was successful
func (h *BaseHandler) BindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		h.BadRequest(c, err.Error())
		return false
	}
	return true
}

// BindQuery binds query parameters and sends error response if invalid
func (h *BaseHandler) BindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		h.BadRequest(c, err.Error())
		return false
	}
	return true
}

// ============================
// Authorization Helpers
// ============================

// RequireOwnerOrAdmin checks ownership and sends 403 if unauthorized
// Returns true if authorized
func (h *BaseHandler) RequireOwnerOrAdmin(c *gin.Context, ownerID uint) bool {
	if !h.IsOwnerOrAdmin(c, ownerID) {
		h.Forbidden(c, "You are not authorized to perform this action")
		return false
	}
	return true
}

// RequireAdmin checks admin role and sends 403 if unauthorized
func (h *BaseHandler) RequireAdmin(c *gin.Context) bool {
	if !h.IsAdmin(c) {
		h.Forbidden(c, "Admin access required")
		return false
	}
	return true
}

// ============================
// Service Error Handling
// ============================

// HandleServiceError processes service layer errors
func (h *BaseHandler) HandleServiceError(c *gin.Context, err error, notFoundResource string) {
	if err == nil {
		return
	}

	// Check for common error messages
	errMsg := err.Error()

	switch errMsg {
	case "not found", "record not found":
		h.NotFound(c, notFoundResource)
	case "unauthorized":
		h.Forbidden(c, "You are not authorized to perform this action")
	case "forbidden":
		h.Forbidden(c, "Access denied")
	default:
		// Check if it's an AppError
		if appErr, ok := apperror.IsAppError(err); ok {
			h.Error(c, appErr)
			return
		}
		// Default to internal error for unknown errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
	}
}
