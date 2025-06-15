package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// QueryParamParser parses and validates query parameters from the request.
func QueryParamParser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters.
		params, err := query_params.ParseQueryParams(c)
		if err != nil {
			slog.Warn("Failed to parse query parameters", "path", c.Request.URL.Path, "error", err)
			c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters: "+err.Error()))
			c.Abort()
			return
		}

		// Store the parsed parameters in the context.
		c.Set("queryParams", params)
		c.Next()
	}
}
