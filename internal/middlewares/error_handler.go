package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/pkg/logger"
	"github.com/go-backend-template/pkg/response"
)

// ErrorHandler middleware with improved logging using the new logger pattern
func ErrorHandler() gin.HandlerFunc {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				stackTrace := string(debug.Stack())

				// Use context-aware logging
				logger.Error(ctx, "Panic recovered",
					"error", err,
					"url", c.Request.URL.Path,
					"stack_trace", stackTrace,
					"env", env,
				)

				errorMessage := "Internal Server Error"
				if env == "dev" {
					errorMessage = fmt.Sprintf("Panic: %v", err)
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError,
					response.NewErrorResponse(errorMessage))
			}
		}()

		c.Next()

		// Handle unhandled errors
		if len(c.Errors) > 0 {
			ctx := c.Request.Context()
			err := c.Errors.Last()

			logger.Error(ctx, "Unhandled error",
				"error", err.Error(),
				"url", c.Request.URL.Path,
				"env", env,
			)

			if !c.Writer.Written() {
				errMsg := "Internal Server Error"
				if env == "dev" {
					errMsg = err.Error()
				}
				c.JSON(http.StatusInternalServerError,
					response.NewErrorResponse(errMsg))
			}
		}
	}
}
