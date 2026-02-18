package slug

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var (
	regexpNonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]+`)
	regexpMultipleHyphens = regexp.MustCompile(`-{2,}`)
)

// Generate creates a URL-safe slug from a given string.
func Generate(input string) string {
	s := strings.ToLower(input)

	// Replace non-alphanumeric characters with hyphens
	s = regexpNonAlphanumeric.ReplaceAllString(s, "-")

	// Replace multiple consecutive hyphens with single hyphen
	s = regexpMultipleHyphens.ReplaceAllString(s, "-")

	// Trim leading/trailing hyphens
	s = strings.Trim(s, "-")

	// Truncate to reasonable length
	if len(s) > 200 {
		s = s[:200]
		s = strings.TrimRight(s, "-")
	}

	if s == "" {
		s = "untitled"
	}

	return s
}

// GenerateUnique creates a slug with a random suffix to ensure uniqueness.
func GenerateUnique(input string) string {
	base := Generate(input)
	suffix := randomSuffix()
	return fmt.Sprintf("%s-%s", base, suffix)
}

// randomSuffix generates a short random alphanumeric string
func randomSuffix() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
