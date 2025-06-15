package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderRequestID = "X-Request-ID"
	ContextRequestIDKey = "requestID"
)

// RequestID is a middleware that injects a request ID into the context and response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate a new one
		requestID := c.Request.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Set request ID in context
		c.Set(ContextRequestIDKey, requestID)

		// Set request ID in response header
		c.Header(HeaderRequestID, requestID)

		c.Next()
	}
}
