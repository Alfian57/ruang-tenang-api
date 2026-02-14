package service

import "errors"

var (
	ErrLevelExists = errors.New("level already exists")
	ErrNotFound    = errors.New("resource not found")
	ErrForbidden   = errors.New("access forbidden")
)

// ServiceError represents a structured service error
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// IsServiceError checks if an error is a ServiceError
func IsServiceError(err error) (*ServiceError, bool) {
	if se, ok := err.(*ServiceError); ok {
		return se, true
	}
	return nil, false
}
