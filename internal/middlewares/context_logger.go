package middlewares

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/pkg/logger" // Import your logger package
)

// ContextLogger is a middleware that sets up a request-scoped logger with request_id.
func ContextLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get(ContextRequestIDKey) // Retrieve requestID from context (set by RequestID middleware)

		// Create a new logger instance with the request_id
		// Use the default logger as a base
		requestScopedLogger := slog.Default().With("request_id", requestID)

		// Store the new logger in the context
		// Assumes logger.WithContext is available from your logger package
		// and logger.loggerKey is the key used for storing/retrieving the logger.
		ctxWithLogger := logger.WithContext(c.Request.Context(), requestScopedLogger)
		c.Request = c.Request.WithContext(ctxWithLogger)

		c.Next()
	}
}
