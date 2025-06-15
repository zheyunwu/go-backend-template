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
	Filter map[string]interface{} // Changed to interface{} to support arrays and nested structures
	Sort   string
	Page   int
	Limit  int
}

// ParseQueryParams parses common query parameters for list APIs.
// Example: /products?search=detergent&filter={"categories":[1,2]}&sort=updated_at:desc&page=1&limit=10
func ParseQueryParams(c *gin.Context) (*QueryParams, error) {
	// Initialize query parameters.
	q := &QueryParams{
		Filter: make(map[string]interface{}),
	}

	// 1. Parse 'search' parameter.
	q.Search = c.Query("search")

	// 2. Parse 'filter' parameter.
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

	// 3. Parse 'sort' parameter.
	sort := c.Query("sort")
	if sort != "" {
		// 'sort' can be in two forms: sort=title:desc or sort=title:asc (can be shortened to sort=title)
		parts := strings.Split(sort, ":")
		if len(parts) == 2 {
			q.Sort = parts[0] + " " + strings.ToUpper(parts[1])
		} else {
			q.Sort = sort
		}
	}

	// 4. Parse 'page' parameter.
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		slog.Warn("Invalid page parameter", "page", pageStr, "error", err)
		return nil, err
	}
	// Validate and set page number (page number less than 1 is not allowed).
	if page < 1 {
		slog.Debug("Invalid page value, using default", "requestedPage", page)
		page = 1
	}
	q.Page = page

	// 5. Parse 'limit' parameter.
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		slog.Warn("Invalid limit parameter", "limit", limitStr, "error", err)
		return nil, err
	}
	// Validate and set items per page (limit range between 1-100).
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
