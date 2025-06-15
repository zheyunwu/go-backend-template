package middlewares

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/logger"
	"github.com/google/uuid"
)

const (
	HeaderRequestID     = "X-Request-ID"
	ContextRequestIDKey = "requestID"
)

// LoggingMiddleware is a Gin middleware that logs incoming requests and their responses.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or get request ID
		requestID := c.Request.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Set request ID in gin context for other middlewares
		c.Set(ContextRequestIDKey, requestID)
		// Set request ID in response header
		c.Header(HeaderRequestID, requestID)

		// Create request-scoped logger with request ID and store in context
		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Request start time and basic info
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Log request start
		logger.Debug(ctx, "Request started",
			"method", method,
			"path", path,
			"client_ip", clientIP,
		)

		// Process request
		c.Next()

		// Calculate request duration and get response info
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// Add user ID to logger if authenticated
		if user, exists := c.Get("authenticatedUser"); exists {
			if userModel, ok := user.(models.User); ok {
				ctx = logger.WithUserID(ctx, fmt.Sprintf("%d", userModel.ID))
			}
		}

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		// Log request completion
		requestLogger := logger.FromContext(ctx)
		requestLogger.LogAttrs(ctx, logLevel, "Request completed",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status_code", statusCode),
			slog.Duration("duration", duration),
			slog.String("client_ip", clientIP),
		)
	}
}
