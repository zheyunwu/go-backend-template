package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/logger"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// CategoryHandler handles HTTP requests related to categories.
type CategoryHandler struct {
	CategoryService services.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: categoryService,
	}
}

// ListCategories retrieves a list of categories.
func (h *CategoryHandler) ListCategories(ctx *gin.Context) {
	// Get parsed query parameters from context.
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		logger.Warn(ctx, "Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// Get category list.
	categories, pagination, err := h.CategoryService.ListCategories(ctx.Request.Context(), queryParams) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(categories, "", *pagination))
}

// GetCategoryTree retrieves the category tree structure.
func (h *CategoryHandler) GetCategoryTree(ctx *gin.Context) {
	// Get depth setting from query parameters.
	depthStr := ctx.DefaultQuery("depth", "0")
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		logger.Warn(ctx, "Invalid depth parameter", "depth", depthStr)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid depth parameter, must be a non-negative integer"))
		return
	}

	// Depth must be a non-negative integer.
	if depth < 0 {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Depth must be a non-negative integer"))
		return
	}

	// Get category tree.
	categoryTree, err := h.CategoryService.GetCategoryTree(ctx.Request.Context(), depth, true) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(categoryTree, ""))
}
