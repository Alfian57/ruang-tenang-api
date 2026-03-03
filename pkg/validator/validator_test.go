package validator

import (
	"regexp"
	"testing"
)

func TestValidationChainAndToAppError(t *testing.T) {
	v := NewValidation()
	v.Required("email", "", "").Email("email", "not-an-email", "")

	if !v.HasErrors() {
		t.Fatal("expected validation errors")
	}
	if len(v.Errors()) < 2 {
		t.Fatalf("expected at least 2 errors, got %d", len(v.Errors()))
	}

	appErr := v.ToAppError()
	if appErr == nil {
		t.Fatal("expected app error from validation result")
	}
}

func TestStandaloneValidatorsAndSanitize(t *testing.T) {
	if !IsValidEmail("user@example.com") || IsValidEmail("bad-email") {
		t.Fatal("email validation mismatch")
	}
	if !IsValidURL("https://example.com") || IsValidURL("ftp://invalid") {
		t.Fatal("url validation mismatch")
	}
	if !IsValidUsername("user_name-1") || IsValidUsername("invalid name") {
		t.Fatal("username validation mismatch")
	}
	if !IsValidSlug("my-awesome-slug") || IsValidSlug("Bad Slug") {
		t.Fatal("slug validation mismatch")
	}
	if !IsValidPhone("+6281234567890") || IsValidPhone("abc") {
		t.Fatal("phone validation mismatch")
	}

	if got := Sanitize("  hi  "); got != "hi" {
		t.Fatalf("unexpected sanitize result: %q", got)
	}

	if got := SanitizeMultiline(" a \n b  "); got != "a\nb" {
		t.Fatalf("unexpected multiline sanitize result: %q", got)
	}
}

func TestAdditionalChainableValidationMethods(t *testing.T) {
	v := NewValidation()
	v.RequiredInt("age", 0, "")
	v.RequiredUint("count", 0, "")
	v.MinLength("title", "ab", 3, "")
	v.MaxLength("title", "abcdef", 3, "")
	v.URL("website", "invalid-url", "")
	v.Phone("phone", "abcd", "")
	v.Username("username", "bad name", "")
	v.Slug("slug", "Bad Slug", "")
	v.MinValue("score", 1, 5, "")
	v.MaxValue("score", 10, 5, "")
	v.InSlice("status", "other", []string{"draft", "published"}, "")
	v.Match("code", "abc", regexp.MustCompile(`^[0-9]+$`), "")
	v.Custom(false, "custom", "custom failed")

	if !v.HasErrors() {
		t.Fatal("expected validation errors")
	}
	if len(v.Errors()) < 13 {
		t.Fatalf("expected many validation errors, got %d", len(v.Errors()))
	}

	v2 := NewValidation()
	v2.RequiredInt("age", 10, "")
	v2.RequiredUint("count", 1, "")
	v2.MinLength("title", "abcd", 3, "")
	v2.MaxLength("title", "abc", 3, "")
	v2.URL("website", "https://example.com", "")
	v2.Phone("phone", "+62 812-3456-7890", "")
	v2.Username("username", "user_name", "")
	v2.Slug("slug", "my-slug", "")
	v2.MinValue("score", 10, 5, "")
	v2.MaxValue("score", 2, 5, "")
	v2.InSlice("status", "draft", []string{"draft", "published"}, "")
	v2.Match("code", "123", regexp.MustCompile(`^[0-9]+$`), "")
	v2.Custom(true, "custom", "")

	if v2.HasErrors() {
		t.Fatalf("expected no validation errors, got %+v", v2.Errors())
	}
	if v2.ToAppError() != nil {
		t.Fatal("expected nil app error when no validation errors")
	}
}

func TestOptionalValidatorsSkipEmptyValue(t *testing.T) {
	v := NewValidation()
	v.Email("email", "", "")
	v.URL("website", "", "")
	v.Phone("phone", "", "")
	v.Username("username", "", "")
	v.Slug("slug", "", "")
	v.InSlice("status", "", []string{"draft"}, "")
	v.Match("code", "", regexp.MustCompile(`^[0-9]+$`), "")

	if v.HasErrors() {
		t.Fatalf("expected no errors for empty optional values, got %+v", v.Errors())
	}
}

func TestValidatorsRespectCustomMessages(t *testing.T) {
	v := NewValidation()
	v.Email("email", "bad", "email custom")
	v.URL("url", "bad", "url custom")
	v.Phone("phone", "bad", "phone custom")
	v.Username("username", "bad name", "username custom")
	v.Slug("slug", "Bad Slug", "slug custom")
	v.InSlice("status", "other", []string{"a", "b"}, "slice custom")
	v.Match("code", "abc", regexp.MustCompile(`^[0-9]+$`), "match custom")

	errs := v.Errors()
	if len(errs) != 7 {
		t.Fatalf("expected 7 errors, got %d", len(errs))
	}
	expected := map[string]bool{
		"email custom":    false,
		"url custom":      false,
		"phone custom":    false,
		"username custom": false,
		"slug custom":     false,
		"slice custom":    false,
		"match custom":    false,
	}
	for _, fe := range errs {
		if _, ok := expected[fe.Message]; ok {
			expected[fe.Message] = true
		}
	}
	for msg, seen := range expected {
		if !seen {
			t.Fatalf("expected message %q in errors", msg)
		}
	}
}
