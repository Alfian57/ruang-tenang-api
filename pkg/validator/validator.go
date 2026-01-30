package validator

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Alfian57/ruang-tenang-api/pkg/apperror"
)

// Validation patterns
var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	slugPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	phonePattern    = regexp.MustCompile(`^[+]?[0-9]{10,15}$`)
	urlPattern      = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)

// ValidationResult holds validation state
type ValidationResult struct {
	errors []apperror.FieldError
}

// NewValidation creates a new validation result
func NewValidation() *ValidationResult {
	return &ValidationResult{
		errors: make([]apperror.FieldError, 0),
	}
}

// AddError adds a field error
func (v *ValidationResult) AddError(field, message string) *ValidationResult {
	v.errors = append(v.errors, apperror.FieldError{
		Field:   field,
		Message: message,
	})
	return v
}

// HasErrors checks if there are any validation errors
func (v *ValidationResult) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *ValidationResult) Errors() []apperror.FieldError {
	return v.errors
}

// ToAppError converts validation result to AppError if there are errors
func (v *ValidationResult) ToAppError() *apperror.AppError {
	if !v.HasErrors() {
		return nil
	}
	return apperror.ValidationWithDetails("Validation failed", v.errors)
}

// ============================
// Chainable Validation Methods
// ============================

// Required validates that a field is not empty
func (v *ValidationResult) Required(field, value, message string) *ValidationResult {
	if strings.TrimSpace(value) == "" {
		if message == "" {
			message = field + " is required"
		}
		v.AddError(field, message)
	}
	return v
}

// RequiredInt validates that an integer field is set (non-zero)
func (v *ValidationResult) RequiredInt(field string, value int, message string) *ValidationResult {
	if value == 0 {
		if message == "" {
			message = field + " is required"
		}
		v.AddError(field, message)
	}
	return v
}

// RequiredUint validates that a uint field is set (non-zero)
func (v *ValidationResult) RequiredUint(field string, value uint, message string) *ValidationResult {
	if value == 0 {
		if message == "" {
			message = field + " is required"
		}
		v.AddError(field, message)
	}
	return v
}

// MinLength validates minimum string length
func (v *ValidationResult) MinLength(field, value string, min int, message string) *ValidationResult {
	if utf8.RuneCountInString(value) < min {
		if message == "" {
			message = field + " must be at least " + string(rune(min+'0')) + " characters"
		}
		v.AddError(field, message)
	}
	return v
}

// MaxLength validates maximum string length
func (v *ValidationResult) MaxLength(field, value string, max int, message string) *ValidationResult {
	if utf8.RuneCountInString(value) > max {
		if message == "" {
			message = field + " must not exceed " + string(rune(max+'0')) + " characters"
		}
		v.AddError(field, message)
	}
	return v
}

// Email validates email format
func (v *ValidationResult) Email(field, value, message string) *ValidationResult {
	if value == "" {
		return v // Skip if empty (use Required for mandatory check)
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		if message == "" {
			message = "Invalid email format"
		}
		v.AddError(field, message)
	}
	return v
}

// URL validates URL format
func (v *ValidationResult) URL(field, value, message string) *ValidationResult {
	if value == "" {
		return v
	}
	if !urlPattern.MatchString(value) {
		if message == "" {
			message = "Invalid URL format"
		}
		v.AddError(field, message)
	}
	return v
}

// Phone validates phone number format
func (v *ValidationResult) Phone(field, value, message string) *ValidationResult {
	if value == "" {
		return v
	}
	cleaned := strings.ReplaceAll(strings.ReplaceAll(value, " ", ""), "-", "")
	if !phonePattern.MatchString(cleaned) {
		if message == "" {
			message = "Invalid phone number format"
		}
		v.AddError(field, message)
	}
	return v
}

// Username validates username format (alphanumeric, underscore, dash)
func (v *ValidationResult) Username(field, value, message string) *ValidationResult {
	if value == "" {
		return v
	}
	if !usernamePattern.MatchString(value) {
		if message == "" {
			message = "Username can only contain letters, numbers, underscores, and dashes"
		}
		v.AddError(field, message)
	}
	return v
}

// Slug validates slug format
func (v *ValidationResult) Slug(field, value, message string) *ValidationResult {
	if value == "" {
		return v
	}
	if !slugPattern.MatchString(value) {
		if message == "" {
			message = "Invalid slug format"
		}
		v.AddError(field, message)
	}
	return v
}

// MinValue validates minimum numeric value
func (v *ValidationResult) MinValue(field string, value, min int, message string) *ValidationResult {
	if value < min {
		if message == "" {
			message = field + " must be at least " + string(rune(min+'0'))
		}
		v.AddError(field, message)
	}
	return v
}

// MaxValue validates maximum numeric value
func (v *ValidationResult) MaxValue(field string, value, max int, message string) *ValidationResult {
	if value > max {
		if message == "" {
			message = field + " must not exceed " + string(rune(max+'0'))
		}
		v.AddError(field, message)
	}
	return v
}

// InSlice validates that a value is in a list of allowed values
func (v *ValidationResult) InSlice(field, value string, allowed []string, message string) *ValidationResult {
	if value == "" {
		return v
	}
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	if message == "" {
		message = field + " must be one of: " + strings.Join(allowed, ", ")
	}
	v.AddError(field, message)
	return v
}

// Match validates that a value matches a regex pattern
func (v *ValidationResult) Match(field, value string, pattern *regexp.Regexp, message string) *ValidationResult {
	if value == "" {
		return v
	}
	if !pattern.MatchString(value) {
		if message == "" {
			message = field + " has invalid format"
		}
		v.AddError(field, message)
	}
	return v
}

// Custom allows custom validation logic
func (v *ValidationResult) Custom(valid bool, field, message string) *ValidationResult {
	if !valid {
		v.AddError(field, message)
	}
	return v
}

// ============================
// Standalone Validation Functions
// ============================

// IsValidEmail checks if email is valid
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidURL checks if URL is valid
func IsValidURL(url string) bool {
	return urlPattern.MatchString(url)
}

// IsValidUsername checks if username is valid
func IsValidUsername(username string) bool {
	return usernamePattern.MatchString(username)
}

// IsValidSlug checks if slug is valid
func IsValidSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

// IsValidPhone checks if phone number is valid
func IsValidPhone(phone string) bool {
	cleaned := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
	return phonePattern.MatchString(cleaned)
}

// Sanitize trims whitespace and normalizes strings
func Sanitize(s string) string {
	return strings.TrimSpace(s)
}

// SanitizeMultiline sanitizes multiline text while preserving newlines
func SanitizeMultiline(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}
