package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/pkg/response"
)

/*
ErrorHandler 中间件的职责：
全局错误处理：作为应用的最后一道防线
捕获未处理的 panic：防止因未捕获的异常而导致整个应用崩溃
关注点：非预期的严重错误和系统级异常
执行范围：针对所有请求的全局保护机制
处理方式：记录堆栈跟踪、恢复应用正常运行、提供通用错误响应
*/
func ErrorHandler() gin.HandlerFunc {
	// 读取当前环境
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	return func(c *gin.Context) {
		// 使用defer recover来捕获任何panic
		defer func() {
			if err := recover(); err != nil {
				// 记录错误和堆栈跟踪
				stackTrace := string(debug.Stack())
				slog.Error("Panic recovered",
					"error", err,
					"url", c.Request.URL.Path,
					"stackTrace", stackTrace,
					"env", env,
				)

				// 构建错误响应
				errorMessage := "Internal Server Error"
				if env == "dev" {
					errorMessage = fmt.Sprintf("Panic: %v", err)
				}

				// 返回500响应
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					response.NewErrorResponse(errorMessage))
			}
		}()

		// 处理请求
		c.Next()

		// 检查是否有未处理的错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last()
			slog.Error("Unhandled error",
				"error", err.Error(),
				"url", c.Request.URL.Path,
				"env", env,
			)

			// 如果响应还没有被写入
			if !c.Writer.Written() {
				errMsg := "Internal Server Error"
				if env == "dev" {
					errMsg = err.Error() // 在开发环境显示详细错误
				}
				c.JSON(http.StatusInternalServerError,
					response.NewErrorResponse(errMsg))
			}
		}
	}
}
