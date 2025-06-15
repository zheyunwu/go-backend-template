package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// loggerKey is the key for storing a logger in the context.
const loggerKey contextKey = "logger"

// Config holds logger configuration.
type Config struct {
	Level      slog.Level // Log level (Debug, Info, Warn, Error)
	JSONFormat bool       // Whether to output logs in JSON format
	Output     io.Writer  // Output destination for logs (e.g., os.Stdout)
	Service    string     // Service name for structured logging
	Version    string     // Service version
}

// Global logger instance
var defaultLogger *slog.Logger

// Init initializes the global slog logger.
func Init(cfg *Config) {
	if cfg == nil {
		cfg = &Config{
			Level:      slog.LevelInfo,
			JSONFormat: false,
			Output:     os.Stdout,
			Service:    "go-backend-template",
			Version:    "1.0.0",
		}
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: true, // Include source file and line number in logs
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format the time field
			if a.Key == "time" {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	if cfg.JSONFormat {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	// Create base logger with service metadata
	baseLogger := slog.New(handler)
	defaultLogger = baseLogger.With(
		"service", cfg.Service,
		"version", cfg.Version,
	)
	
	slog.SetDefault(defaultLogger)
}

// FromContext retrieves the logger from the context, or returns the default logger if not found.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return defaultLogger
}

// WithContext stores the logger in the context.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// WithRequestID creates a new logger with request ID and stores it in context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := defaultLogger.With("request_id", requestID)
	return WithContext(ctx, logger)
}

// WithUserID adds user ID to the logger in context
func WithUserID(ctx context.Context, userID interface{}) context.Context {
	logger := FromContext(ctx).With("user_id", userID)
	return WithContext(ctx, logger)
}

// Convenience functions for common logging patterns
func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}
