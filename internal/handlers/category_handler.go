package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// CategoryHandler 处理与分类相关的HTTP请求
type CategoryHandler struct {
	CategoryService services.CategoryService
}

// NewCategoryHandler 创建一个新的CategoryHandler
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: categoryService,
	}
}

// ListCategories 获取分类列表
func (h *CategoryHandler) ListCategories(ctx *gin.Context) {
	// 从上下文中获取已解析的查询参数
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		slog.Warn("Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// 获取分类列表
	categories, pagination, err := h.CategoryService.ListCategories(queryParams)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(categories, "", *pagination))
}

// GetCategoryTree 获取分类树结构
func (h *CategoryHandler) GetCategoryTree(ctx *gin.Context) {
	// 从查询参数中获取深度设置
	depthStr := ctx.DefaultQuery("depth", "0")
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		slog.Warn("Invalid depth parameter", "depth", depthStr)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid depth parameter, must be a non-negative integer"))
		return
	}

	// 深度必须是非负整数
	if depth < 0 {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Depth must be a non-negative integer"))
		return
	}

	// 获取分类树
	categoryTree, err := h.CategoryService.GetCategoryTree(depth, true)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(categoryTree, ""))
}
