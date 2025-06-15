package middlewares

import (
	"fmt"
	"log/slog" // Keep slog for level constants if needed, or remove if logger pkg has them
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/logger" // Import your logger package
)

/*
RequestLogger 中间件的职责：
记录 HTTP 请求的生命周期：捕获请求开始、完成和相关指标
针对每个 HTTP 请求执行：为每个进入应用的请求记录详细信息
计算指标：记录响应时间、状态码等请求特定的信息
关注点：HTTP 请求的跟踪和监控
*/
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求开始时间
		startTime := time.Now()

		// 请求路径
		path := c.Request.URL.Path

		// 请求方法
		method := c.Request.Method

		// 客户端IP
		clientIP := c.ClientIP()

		// Retrieve the logger from context. This logger should already have request_id.
		log := logger.FromContext(c.Request.Context())

		// 记录请求开始
		log.Debug("Request started",
			"method", method,
			"path", path,
			"clientIP", clientIP,
		)

		// 处理请求
		c.Next()

		// 请求结束时间
		endTime := time.Now()

		// 计算延迟时间
		latency := endTime.Sub(startTime)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 获取用户ID（如果存在）
		authenticatedUser, exists := c.Get("authenticatedUser")
		userIDValue := ""
		if exists {
			if user, ok := authenticatedUser.(models.User); ok {
				userIDValue = fmt.Sprintf("%d", user.ID)
			}
		}

		// 记录请求完成
		logLevel := slog.LevelInfo

		// 对于错误响应，使用警告级别
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		}

		// 对于服务器错误，使用错误级别
		if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		log.LogAttrs(c.Request.Context(), logLevel, "Request completed",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("statusCode", statusCode),
			slog.String("latency", latency.String()),
			slog.String("clientIP", clientIP),
			slog.String("userID", userIDValue),
		)
	}
}
