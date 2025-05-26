package logger

import (
	"io"
	"log/slog"
	"os"
	"time"
)

// 日志级别常量
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Config 日志配置
type Config struct {
	Level      slog.Level
	JSONFormat bool
	Output     io.Writer
}

// DefaultConfig 返回默认日志配置
func DefaultConfig() *Config {
	return &Config{
		Level:      LevelInfo,
		JSONFormat: false,
		Output:     os.Stdout,
	}
}

// Init 初始化全局 slog logger
func Init(cfg *Config) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 格式化时间字段
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
