package queryutil

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// DefaultPagination values
const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

// GetPagination extracts pagination parameters from query string
// Supports both page/limit and limit/offset patterns
func GetPagination(c *gin.Context) PaginationParams {
	page := GetIntQuery(c, "page", DefaultPage)
	limit := GetIntQuery(c, "limit", DefaultLimit)
	offset := GetIntQuery(c, "offset", -1)

	// Validate limits
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Calculate offset from page if offset not explicitly provided
	if offset < 0 {
		offset = (page - 1) * limit
	} else {
		// If offset provided, calculate page from it
		page = (offset / limit) + 1
	}

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// GetIntQuery gets an integer query parameter with default value
func GetIntQuery(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetInt64Query gets an int64 query parameter with default value
func GetInt64Query(c *gin.Context, key string, defaultValue int64) int64 {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetUintQuery gets a uint query parameter with default value
func GetUintQuery(c *gin.Context, key string, defaultValue uint) uint {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return uint(value)
}

// GetBoolQuery gets a boolean query parameter with default value
func GetBoolQuery(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetStringQuery gets a string query parameter with default value
func GetStringQuery(c *gin.Context, key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetOptionalUint gets an optional uint query parameter
// Returns nil if not provided
func GetOptionalUint(c *gin.Context, key string) *uint {
	valueStr := c.Query(key)
	if valueStr == "" {
		return nil
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return nil
	}
	v := uint(value)
	return &v
}

// GetOptionalInt gets an optional int query parameter
// Returns nil if not provided
func GetOptionalInt(c *gin.Context, key string) *int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return nil
	}
	return &value
}

// GetOptionalString gets an optional string query parameter
// Returns nil if not provided
func GetOptionalString(c *gin.Context, key string) *string {
	value := c.Query(key)
	if value == "" {
		return nil
	}
	return &value
}

// SortParams holds sorting parameters
type SortParams struct {
	Field string
	Order string // "asc" or "desc"
}

// GetSort extracts sorting parameters from query string
func GetSort(c *gin.Context, defaultField, defaultOrder string, allowedFields []string) SortParams {
	field := GetStringQuery(c, "sort", defaultField)
	order := GetStringQuery(c, "order", defaultOrder)

	// Validate sort field
	isAllowed := false
	for _, allowed := range allowedFields {
		if field == allowed {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		field = defaultField
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = defaultOrder
	}

	return SortParams{
		Field: field,
		Order: order,
	}
}

// GetIntParam gets an integer path parameter with error handling
func GetIntParam(c *gin.Context, key string) (int, bool) {
	valueStr := c.Param(key)
	if valueStr == "" {
		return 0, false
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, false
	}
	return value, true
}

// GetUintParam gets a uint path parameter with error handling
func GetUintParam(c *gin.Context, key string) (uint, bool) {
	valueStr := c.Param(key)
	if valueStr == "" {
		return 0, false
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(value), true
}

// MustGetUintParam gets a uint path parameter, returns 0 on error
// Use when you want simpler code and will validate later
func MustGetUintParam(c *gin.Context, key string) uint {
	value, _ := GetUintParam(c, key)
	return value
}

// MustGetIntParam gets an int path parameter, returns 0 on error
func MustGetIntParam(c *gin.Context, key string) int {
	value, _ := GetIntParam(c, key)
	return value
}
