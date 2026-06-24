package validator

import (
	"testing"
)

// TestMinMaxLengthMessageFormatting guards against the previous
// string(rune(min+'0')) bug, which produced a single wrong rune (e.g. min=11
// rendered as ';') instead of the decimal number in validation messages.
func TestMinMaxLengthMessageFormatting(t *testing.T) {
	tests := []struct {
		name string
		fn   func() *ValidationResult
		want string
	}{
		{
			name: "min length 11 renders digits",
			fn:   func() *ValidationResult { return NewValidation().MinLength("field", "ab", 11, "") },
			want: "field must be at least 11 characters",
		},
		{
			name: "max length 5 renders digits",
			fn:   func() *ValidationResult { return NewValidation().MaxLength("field", "abcdefgh", 5, "") },
			want: "field must not exceed 5 characters",
		},
		{
			name: "min value 10 renders digits",
			fn:   func() *ValidationResult { return NewValidation().MinValue("field", 5, 10, "") },
			want: "field must be at least 10",
		},
		{
			name: "max value 50 renders digits",
			fn:   func() *ValidationResult { return NewValidation().MaxValue("field", 99, 50, "") },
			want: "field must not exceed 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.fn()
			if !v.HasErrors() {
				t.Fatalf("expected an error, got none")
			}
			got := v.Errors()[0].Message
			if got != tt.want {
				t.Fatalf("message = %q, want %q", got, tt.want)
			}
		})
	}
}
