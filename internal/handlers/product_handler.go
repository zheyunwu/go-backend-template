package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// ProductHandler 处理产品相关的用户API
type ProductHandler struct {
	ProductService     services.ProductService
	InteractionService services.UserInteractionService
}

// NewProductHandler 创建产品处理器
func NewProductHandler(
	productService services.ProductService,
	interactionService services.UserInteractionService,
) *ProductHandler {
	return &ProductHandler{
		ProductService:     productService,
		InteractionService: interactionService,
	}
}

// ListProducts 获取产品列表
// 支持Query Parameters:
// - is_liked: 是否获取用户点赞的产品
// - is_favorited: 是否获取用户收藏的产品
// - is_reviewed: 是否获取用户已评价的产品
func (h *ProductHandler) ListProducts(ctx *gin.Context) {
	// 从上下文中获取已解析的Query Parameters
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		slog.Warn("Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// 获取定制Query Parameters
	isLiked, _ := strconv.ParseBool(ctx.Query("is_liked"))
	isFavorited, _ := strconv.ParseBool(ctx.Query("is_favorited"))

	// 获取当前authenticatedUser（如果已登录）
	var userID uint
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if ok {
		userID = authenticatedUser.ID
	}

	var products []models.Product
	var pagination *response.Pagination
	var err error

	// 根据筛选参数决定查询方式
	if isLiked && userID > 0 {
		// 获取用户点赞的产品
		products, pagination, err = h.InteractionService.ListUserLikedProducts(userID, queryParams)
	} else if isFavorited && userID > 0 {
		// 获取用户收藏的产品
		products, pagination, err = h.InteractionService.ListUserFavoritedProducts(userID, queryParams)
	} else {
		// 获取所有产品
		products, pagination, err = h.ProductService.ListProducts(queryParams)
	}

	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 转换为用户DTO
	userProducts := make([]dto.UserProductDTO, 0, len(products))
	for i := range products {
		dto := dto.ToUserProductDTO(&products[i])
		if dto != nil {
			// 如果用户已登录，检查收藏和点赞状态
			if userID > 0 {
				// 查询点赞状态
				isLiked, errLike := h.InteractionService.IsLiked(userID, dto.ID)
				if errLike == nil {
					dto.IsLiked = isLiked
				}

				// 查询收藏状态
				isFavorited, errFav := h.InteractionService.IsFavorited(userID, dto.ID)
				if errFav == nil {
					dto.IsFavorited = isFavorited
				}
			}
			userProducts = append(userProducts, *dto)
		}
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProducts, "", *pagination))
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

	// 转换为用户DTO
	userProduct := dto.ToUserProductDTO(product)

	// 获取当前authenticatedUser（如果已登录）
	authenticatedUser, exists := handler_utils.GetAuthenticatedUser(ctx)
	if exists && userProduct != nil && authenticatedUser != nil && authenticatedUser.ID > 0 {
		// 检查是否已点赞
		isLiked, errLike := h.InteractionService.IsLiked(authenticatedUser.ID, userProduct.ID)
		if errLike == nil {
			userProduct.IsLiked = isLiked
		}

		// 检查是否已收藏
		isFavorited, errFav := h.InteractionService.IsFavorited(authenticatedUser.ID, userProduct.ID)
		if errFav == nil {
			userProduct.IsFavorited = isFavorited
		}
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProduct, ""))
}
