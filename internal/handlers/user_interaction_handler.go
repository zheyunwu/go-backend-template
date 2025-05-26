package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/response"
)

// UserInteractionHandler 用户交互处理器
type UserInteractionHandler struct {
	InteractionService services.UserInteractionService
}

// NewUserInteractionHandler 创建用户交互处理器
func NewUserInteractionHandler(interactionService services.UserInteractionService) *UserInteractionHandler {
	return &UserInteractionHandler{
		InteractionService: interactionService,
	}
}

/*
点赞
*/

// ToggleLike 切换点赞状态（点赞或取消点赞）
func (h *UserInteractionHandler) ToggleLike(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 获取产品ID
	productID, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 解析请求体，判断是点赞还是取消点赞
	var req struct {
		IsLiked bool `json:"is_liked"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("无效的请求参数"))
		return
	}

	var result error
	var actionMsg string

	if req.IsLiked {
		// 执行点赞操作
		result = h.InteractionService.AddLike(authenticatedUser.ID, uint(productID))
		actionMsg = "点赞成功"
	} else {
		// 执行取消点赞操作
		result = h.InteractionService.RemoveLike(authenticatedUser.ID, uint(productID))
		actionMsg = "取消点赞成功"
	}

	if result != nil {
		handler_utils.HandleError(ctx, result)
		return
	}

	// 返回成功
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(
		gin.H{"is_like": req.IsLiked},
		actionMsg,
	))
}

// ListUserLikes 获取用户点赞的产品列表
// func (h *UserInteractionHandler) ListUserLikes(ctx *gin.Context) {
// // 获取当前authenticatedUser
// authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
// if !ok {
// 	  return
// }

// 	// 获取查询参数
// 	params, _ := ctx.Get("queryParams")
// 	queryParams, ok := params.(*pkg.QueryParams)
// 	if !ok {
// 		slog.Warn("Invalid query parameters type", "params", params)
// 		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
// 		return
// 	}

// 	// 获取点赞列表
// 	products, pagination, err := h.InteractionService.ListUserLikedProducts(authenticatedUser.ID, queryParams)
// 	if err != nil {
// 		handler_utils.HandleError(ctx, err)
// 		return
// 	}

// 	// 转换为用户DTO
// 	userProducts := dto.ToUserProductDTOList(products)

// 	// 返回结果
// 	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProducts, "", *pagination))
// }

/*
收藏
*/

// ToggleFavorite 切换收藏状态（收藏或取消收藏）
func (h *UserInteractionHandler) ToggleFavorite(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 获取产品ID
	productID, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 解析请求体，判断是收藏还是取消收藏
	var req struct {
		IsFavorited bool `json:"is_favorited"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("无效的请求参数"))
		return
	}

	var result error
	var actionMsg string

	if req.IsFavorited {
		// 执行收藏操作
		result = h.InteractionService.AddFavorite(authenticatedUser.ID, uint(productID))
		actionMsg = "收藏成功"
	} else {
		// 执行取消收藏操作
		result = h.InteractionService.RemoveFavorite(authenticatedUser.ID, uint(productID))
		actionMsg = "已取消收藏"
	}

	if result != nil {
		handler_utils.HandleError(ctx, result)
		return
	}

	// 返回成功
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(
		gin.H{"is_favorited": req.IsFavorited},
		actionMsg,
	))
}

// ListUserFavorites 获取用户收藏的产品列表
// func (h *UserInteractionHandler) ListUserFavorites(ctx *gin.Context) {
// // 获取当前authenticatedUser
// authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
// if !ok {
// 	return
// }

// 	// 获取查询参数
// 	params, _ := ctx.Get("queryParams")
// 	queryParams, ok := params.(*pkg.QueryParams)
// 	if !ok {
// 		slog.Warn("Invalid query parameters type", "params", params)
// 		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
// 		return
// 	}

// 	// 获取收藏列表
// 	products, pagination, err := h.InteractionService.ListUserFavoritedProducts(authenticatedUser.ID, queryParams)
// 	if err != nil {
// 		handler_utils.HandleError(ctx, err)
// 		return
// 	}

// 	// 转换为用户DTO
// 	userProducts := dto.ToUserProductDTOList(products)

// 	// 返回结果
// 	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProducts, "", *pagination))
// }

/*
统计
*/

// GetProductStats 获取产品统计数据（收藏数、点赞数）
func (h *UserInteractionHandler) GetProductStats(ctx *gin.Context) {
	// 获取产品ID
	productID, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 获取收藏数
	favoriteCount, err := h.InteractionService.GetProductFavoriteCount(uint(productID))
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 获取点赞数
	likeCount, err := h.InteractionService.GetProductLikeCount(uint(productID))
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回统计数据
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(
		gin.H{
			"favorite_count": favoriteCount,
			"like_count":     likeCount,
		},
		"",
	))
}
