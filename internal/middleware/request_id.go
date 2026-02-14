package middleware

import (
	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware generates a unique request ID for each request.
// If the client sends an X-Request-ID header, it will be reused.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctxutil.SetRequestID(c, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}
