package middlewares

import (
	"fmt"
	"log/slog" // Keep for direct use if logger from context is not available initially
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/pkg/logger" // Import your logger package
	"github.com/go-backend-template/pkg/response"
)

/*
ErrorHandler middleware responsibilities:
Global Error Handling: Acts as the last line of defense for the application.
Catch Unhandled Panics: Prevents the entire application from crashing due to unhandled exceptions.
Focus: Unexpected critical errors and system-level exceptions.
Scope: Global protection mechanism for all requests.
Handling Method: Logs stack traces, recovers normal application operation, provides generic error responses.
*/
func ErrorHandler() gin.HandlerFunc {
	// Read the current environment.
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // Default to "dev" environment.
	}

	return func(c *gin.Context) {
		// Use defer recover to catch any panics.
		defer func() {
			if err := recover(); err != nil {
				// Retrieve logger from context
				log := logger.FromContext(c.Request.Context())

				// Log the error and stack trace.
				stackTrace := string(debug.Stack())
				log.Error("Panic recovered",
					"error", err,
					"url", c.Request.URL.Path,
					"stackTrace", stackTrace,
					"env", env,
				)

				// Construct the error response.
				errorMessage := "Internal Server Error"
				if env == "dev" {
					errorMessage = fmt.Sprintf("Panic: %v", err)
				}

				// Return a 500 response.
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					response.NewErrorResponse(errorMessage))
			}
		}()

		// Process the request.
		c.Next()

		// Check for any unhandled errors.
		if len(c.Errors) > 0 {
			// Retrieve logger from context
			log := logger.FromContext(c.Request.Context())

			// Get the last error.
			err := c.Errors.Last()
			log.Error("Unhandled error",
				"error", err.Error(),
				"url", c.Request.URL.Path,
				"env", env,
			)

			// If the response has not been written yet.
			if !c.Writer.Written() {
				errMsg := "Internal Server Error"
				if env == "dev" {
					errMsg = err.Error() // Display detailed error in development environment.
				}
				c.JSON(http.StatusInternalServerError,
					response.NewErrorResponse(errMsg))
			}
		}
	}
}
