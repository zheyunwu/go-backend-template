package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"context" // Updated to standard library context
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// loggerKey is the key for storing a logger in the context.
const loggerKey contextKey = "logger"

// Log level constants
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Config holds logger configuration.
type Config struct {
	Level      slog.Level // Log level (Debug, Info, Warn, Error)
	JSONFormat bool       // Whether to output logs in JSON format
	Output     io.Writer  // Output destination for logs (e.g., os.Stdout)
}

// DefaultConfig returns the default logger configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:      LevelInfo,
		JSONFormat: false,
		Output:     os.Stdout,
	}
}

// Init initializes the global slog logger.
func Init(cfg *Config) {
	if cfg == nil {
		cfg = DefaultConfig()
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

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// FromContext retrieves the logger from the context, or returns the default logger if not found.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithContext stores the logger in the context.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
