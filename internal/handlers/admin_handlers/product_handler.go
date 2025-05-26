package admin_handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

type ProductHandler struct {
	ProductService services.ProductService
}

func NewProductHandler(productService services.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductService: productService,
	}
}

// ListProducts 获取产品列表
func (h *ProductHandler) ListProducts(ctx *gin.Context) {
	// 从上下文中获取已解析的查询参数
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		slog.Warn("Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// 获取产品列表
	products, pagination, err := h.ProductService.ListProducts(queryParams)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(products, "", *pagination))
}

// GetProduct 获取单个产品详情
func (h *ProductHandler) GetProduct(ctx *gin.Context) {
	// 获取Path参数：product ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 调用 Service层 获取 Product
	product, err := h.ProductService.GetProduct(uint(id))
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(product, ""))
}

// CreateProduct 创建新产品
func (h *ProductHandler) CreateProduct(ctx *gin.Context) {
	// 解析请求体到DTO
	var createReq dto.CreateProductRequest
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		slog.Warn("Invalid product creation request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 转换DTO到产品模型
	product := createReq.ToModel()

	// 转换图片URLs到ProductImage模型
	var images []models.ProductImage
	for _, url := range createReq.ImageURLs {
		if url != "" {
			images = append(images, models.ProductImage{
				ImageURL: url,
			})
		}
	}

	// 调用Service层的创建方法，传递所有模型对象
	createdProductID, err := h.ProductService.CreateProduct(
		product,
		images,
		createReq.CategoryIDs,
	)

	// 处理错误 - 现在只有一个错误返回，因为使用了事务
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回201 Created
	slog.Info("Product created successfully", "productId", createdProductID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdProductID}, ""))
}

// UpdateProduct 更新产品信息
func (h *ProductHandler) UpdateProduct(ctx *gin.Context) {
	// 解析产品ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 解析请求体到DTO
	var updateReq dto.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		slog.Warn("Invalid product update request", "productId", id, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("请求参数错误: "+err.Error()))
		return
	}

	// 转换DTO到更新映射
	updates := updateReq.ToMap()

	// 转换图片URLs到ProductImage模型
	var images []models.ProductImage
	if updateReq.ImageURLs != nil { // 只有在请求中包含图片数组时才处理
		if len(updateReq.ImageURLs) == 0 {
			// 如果 ImageURLs 长度为0，则将images设置为空数组，但不能为nil。
			images = []models.ProductImage{}
		} else {
			for _, url := range updateReq.ImageURLs {
				if url != "" {
					images = append(images, models.ProductImage{
						ImageURL: url,
					})
				}
			}
		}
	}

	// 如果没有要更新的内容，直接返回成功
	if len(updates) == 0 && updateReq.ImageURLs == nil && updateReq.CategoryIDs == nil {
		slog.Info("No fields to update", "productId", id)
		ctx.JSON(http.StatusNoContent, nil)
		return
	}

	// 调用Service层的更新方法，传递所有模型对象
	err = h.ProductService.UpdateProduct(
		uint(id),
		updates,
		images,
		updateReq.CategoryIDs,
	)

	// 处理错误 - 现在只有一个错误返回，因为使用了事务
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("Product updated successfully", "productId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

// DeleteProduct 删除产品
func (h *ProductHandler) DeleteProduct(ctx *gin.Context) {
	// 解析产品ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 调用 Service层 删除 Product
	err = h.ProductService.DeleteProduct(uint(id))
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("Product deleted", "productId", id)
	ctx.JSON(http.StatusNoContent, nil)
}
