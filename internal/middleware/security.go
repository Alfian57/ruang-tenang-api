package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware sets security-related HTTP headers to mitigate
// common web vulnerabilities like XSS, clickjacking, and MIME-type sniffing.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate nonce for script-src
		nonce := generateNonce()
		c.Set("csp-nonce", nonce)

		// Prevent MIME-type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS filter in older browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy - restrict browser features
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Content Security Policy
		c.Header("Content-Security-Policy", buildCSP(nonce))

		c.Next()
	}
}

// buildCSP constructs the Content-Security-Policy header value
func buildCSP(nonce string) string {
	directives := []string{
		"default-src 'self'",
		"script-src 'self' 'nonce-" + nonce + "' 'strict-dynamic'",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"font-src 'self' https://fonts.gstatic.com",
		"img-src 'self' data: blob: https:",
		"connect-src 'self' https:",
		"media-src 'self' https:",
		"frame-ancestors 'none'",
	}

	csp := ""
	for i, directive := range directives {
		if i > 0 {
			csp += "; "
		}
		csp += directive
	}
	return csp
}

// generateNonce generates a cryptographically secure nonce
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
