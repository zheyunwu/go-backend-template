package query_params

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Search string
	Filter map[string]interface{} // 修改为interface{}以支持数组和嵌套结构
	Sort   string
	Page   int
	Limit  int
}

// 解析List接口的通用查询参数
// 例如：/products?search=洗衣剂&filter={"categories":[1,2]}&sort=updated_at:desc&page=1&limit=10
func ParseQueryParams(c *gin.Context) (*QueryParams, error) {
	// 初始化查询参数
	q := &QueryParams{
		Filter: make(map[string]interface{}),
	}

	// 1. 解析 search 参数
	q.Search = c.Query("search")

	// 2. 解析 filter 参数
	filter := c.Query("filter")
	if filter != "" {
		var filterMap map[string]interface{}
		err := json.Unmarshal([]byte(filter), &filterMap)
		if err != nil {
			slog.Warn("Failed to parse filter parameter", "filter", filter, "error", err)
			return nil, err
		}
		q.Filter = filterMap
	}

	// 3. 解析 sort 参数
	sort := c.Query("sort")
	if sort != "" {
		// sort有两种情况：sort=title:desc 或 sort=title:asc (可简写为sort=title)
		parts := strings.Split(sort, ":")
		if len(parts) == 2 {
			q.Sort = parts[0] + " " + strings.ToUpper(parts[1])
		} else {
			q.Sort = sort
		}
	}

	// 4. 解析 page 参数
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		slog.Warn("Invalid page parameter", "page", pageStr, "error", err)
		return nil, err
	}
	// 验证并设置页码（不允许小于1的页码）
	if page < 1 {
		slog.Debug("Invalid page value, using default", "requestedPage", page)
		page = 1
	}
	q.Page = page

	// 5. 解析 limit 参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		slog.Warn("Invalid limit parameter", "limit", limitStr, "error", err)
		return nil, err
	}
	// 验证并设置每页数量（限制范围在1-100之间）
	if limit < 1 {
		slog.Debug("Limit too small, using default", "requestedLimit", limit)
		limit = 10
	} else if limit > 100 {
		slog.Debug("Limit too large, using maximum", "requestedLimit", limit)
		limit = 100
	}
	q.Limit = limit

	return q, nil
}
