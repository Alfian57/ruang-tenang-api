package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware sets security-related HTTP headers to mitigate
// common web vulnerabilities like XSS, clickjacking, and MIME-type sniffing.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Header("Content-Security-Policy", buildCSP())

		c.Next()
	}
}

// buildCSP constructs the Content-Security-Policy header value
func buildCSP() string {
	directives := []string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
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
