package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// QueryParamParser 解析并验证请求中的查询参数
func QueryParamParser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析查询参数
		params, err := query_params.ParseQueryParams(c)
		if err != nil {
			slog.Warn("Failed to parse query parameters", "path", c.Request.URL.Path, "error", err)
			c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters: "+err.Error()))
			c.Abort()
			return
		}

		// 将解析后的参数放入上下文
		c.Set("queryParams", params)
		c.Next()
	}
}
