package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/response"
)

// UserInteractionHandler handles user interaction related HTTP requests.
type UserInteractionHandler struct {
	InteractionService services.UserInteractionService
}

// NewUserInteractionHandler creates a new UserInteractionHandler.
func NewUserInteractionHandler(interactionService services.UserInteractionService) *UserInteractionHandler {
	return &UserInteractionHandler{
		InteractionService: interactionService,
	}
}

/*
Likes
*/

// ToggleLike toggles the like status for a product (likes or unlikes).
func (h *UserInteractionHandler) ToggleLike(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Get product ID.
	productID, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Parse request body to determine if it's a like or unlike action.
	var req struct {
		IsLiked bool `json:"is_liked"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request parameters")) // "无效的请求参数" -> "Invalid request parameters"
		return
	}

	var result error
	var actionMsg string

	if req.IsLiked {
		// Perform like action.
		result = h.InteractionService.AddLike(ctx.Request.Context(), authenticatedUser.ID, uint(productID)) // Pass context
		actionMsg = "Successfully liked" // "点赞成功" -> "Successfully liked"
	} else {
		// Perform unlike action.
		result = h.InteractionService.RemoveLike(ctx.Request.Context(), authenticatedUser.ID, uint(productID)) // Pass context
		actionMsg = "Successfully unliked" // "取消点赞成功" -> "Successfully unliked"
	}

	if result != nil {
		handler_utils.HandleError(ctx, result)
		return
	}

	// Return success.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(
		gin.H{"is_like": req.IsLiked},
		actionMsg,
	))
}

// ListUserLikes retrieves a list of products liked by the user.
// func (h *UserInteractionHandler) ListUserLikes(ctx *gin.Context) {
// // Get current authenticated user.
// authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
// if !ok {
// 	  return
// }

// 	// Get query parameters.
// 	params, _ := ctx.Get("queryParams")
// 	queryParams, ok := params.(*pkg.QueryParams)
// 	if !ok {
// 		slog.Warn("Invalid query parameters type", "params", params)
// 		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
// 		return
// 	}

// 	// Get list of liked products.
// 	products, pagination, err := h.InteractionService.ListUserLikedProducts(authenticatedUser.ID, queryParams)
// 	if err != nil {
// 		handler_utils.HandleError(ctx, err)
// 		return
// 	}

// 	// Convert to user DTO.
// 	userProducts := dto.ToUserProductDTOList(products)

// 	// Return result.
// 	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProducts, "", *pagination))
// }

/*
Favorites
*/

// ToggleFavorite toggles the favorite status for a product (favorites or unfavorites).
func (h *UserInteractionHandler) ToggleFavorite(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Get product ID.
	productID, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Parse request body to determine if it's a favorite or unfavorite action.
	var req struct {
		IsFavorited bool `json:"is_favorited"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request parameters")) // "无效的请求参数" -> "Invalid request parameters"
		return
	}

	var result error
	var actionMsg string

	if req.IsFavorited {
		// Perform favorite action.
		result = h.InteractionService.AddFavorite(ctx.Request.Context(), authenticatedUser.ID, uint(productID)) // Pass context
		actionMsg = "Successfully favorited" // "收藏成功" -> "Successfully favorited"
	} else {
		// Perform unfavorite action.
		result = h.InteractionService.RemoveFavorite(ctx.Request.Context(), authenticatedUser.ID, uint(productID)) // Pass context
		actionMsg = "Successfully unfavorited" // "已取消收藏" -> "Successfully unfavorited"
	}

	if result != nil {
		handler_utils.HandleError(ctx, result)
		return
	}

	// Return success.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(
		gin.H{"is_favorited": req.IsFavorited},
		actionMsg,
	))
}

// ListUserFavorites retrieves a list of products favorited by the user.
// func (h *UserInteractionHandler) ListUserFavorites(ctx *gin.Context) {
// // Get current authenticated user.
// authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
// if !ok {
// 	return
// }

// 	// Get query parameters.
// 	params, _ := ctx.Get("queryParams")
// 	queryParams, ok := params.(*pkg.QueryParams)
// 	if !ok {
// 		slog.Warn("Invalid query parameters type", "params", params)
// 		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
// 		return
// 	}

// 	// Get list of favorited products.
// 	products, pagination, err := h.InteractionService.ListUserFavoritedProducts(authenticatedUser.ID, queryParams)
// 	if err != nil {
// 		handler_utils.HandleError(ctx, err)
// 		return
// 	}

// 	// Convert to user DTO.
// 	userProducts := dto.ToUserProductDTOList(products)

// 	// Return result.
// 	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProducts, "", *pagination))
// }

/*
Statistics
*/

// GetProductStats retrieves product statistics (favorite count, like count).
func (h *UserInteractionHandler) GetProductStats(ctx *gin.Context) {
	// Get product ID.
	productID, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Get favorite count.
	favoriteCount, err := h.InteractionService.GetProductFavoriteCount(ctx.Request.Context(), uint(productID)) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Get like count.
	likeCount, err := h.InteractionService.GetProductLikeCount(ctx.Request.Context(), uint(productID)) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return statistics.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(
		gin.H{
			"favorite_count": favoriteCount,
			"like_count":     likeCount,
		},
		"",
	))
}
