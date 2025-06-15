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

// ProductHandler handles user API requests related to products.
type ProductHandler struct {
	ProductService     services.ProductService
	InteractionService services.UserInteractionService
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(
	productService services.ProductService,
	interactionService services.UserInteractionService,
) *ProductHandler {
	return &ProductHandler{
		ProductService:     productService,
		InteractionService: interactionService,
	}
}

// ListProducts retrieves a list of products.
// Supports Query Parameters:
// - is_liked: whether to fetch products liked by the user.
// - is_favorited: whether to fetch products favorited by the user.
// - is_reviewed: whether to fetch products reviewed by the user (placeholder for future use).
func (h *ProductHandler) ListProducts(ctx *gin.Context) {
	// Get parsed Query Parameters from context.
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		slog.Warn("Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// Get custom Query Parameters.
	isLiked, _ := strconv.ParseBool(ctx.Query("is_liked"))
	isFavorited, _ := strconv.ParseBool(ctx.Query("is_favorited"))

	// Get current authenticated user (if logged in).
	var userID uint
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if ok {
		userID = authenticatedUser.ID
	}

	var products []models.Product
	var pagination *response.Pagination
	var err error

	// Determine query method based on filter parameters.
	if isLiked && userID > 0 {
		// Get products liked by the user.
		products, pagination, err = h.InteractionService.ListUserLikedProducts(ctx.Request.Context(), userID, queryParams) // Pass context
	} else if isFavorited && userID > 0 {
		// Get products favorited by the user.
		products, pagination, err = h.InteractionService.ListUserFavoritedProducts(ctx.Request.Context(), userID, queryParams) // Pass context
	} else {
		// Get all products.
		products, pagination, err = h.ProductService.ListProducts(ctx.Request.Context(), queryParams) // Pass context
	}

	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Convert to user DTO.
	userProducts := make([]dto.UserProductDTO, 0, len(products))
	var productIDs []uint
	if userID > 0 && len(products) > 0 {
		for _, p := range products {
			productIDs = append(productIDs, p.ID)
		}
	}

	interactionStatusMap := make(map[uint]dto.UserInteractionStatus)
	if userID > 0 && len(productIDs) > 0 {
		var interactionErr error
		interactionStatusMap, interactionErr = h.InteractionService.GetUserProductInteractionStatus(ctx.Request.Context(), userID, productIDs)
		if interactionErr != nil {
			// Log the error but proceed, as interaction status is not critical for listing products
			slog.ErrorContext(ctx.Request.Context(), "Failed to get user product interaction status", "userID", userID, "error", interactionErr)
		}
	}

	for i := range products {
		productDTO := dto.ToUserProductDTO(&products[i])
		if productDTO != nil {
			if userID > 0 {
				if status, ok := interactionStatusMap[productDTO.ID]; ok {
					productDTO.IsLiked = status.IsLiked
					productDTO.IsFavorited = status.IsFavorited
				}
			}
			userProducts = append(userProducts, *productDTO)
		}
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProducts, "", *pagination))
}

// GetProduct retrieves details for a single product.
func (h *ProductHandler) GetProduct(ctx *gin.Context) {
	// Get product ID from path parameters.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Call service layer to get the product.
	product, err := h.ProductService.GetProduct(ctx.Request.Context(), uint(id)) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Convert to user DTO.
	userProduct := dto.ToUserProductDTO(product)

	// Get current authenticated user (if logged in).
	authenticatedUser, exists := handler_utils.GetAuthenticatedUser(ctx)
	if exists && userProduct != nil && authenticatedUser != nil && authenticatedUser.ID > 0 {
		// Check if liked.
		isLiked, errLike := h.InteractionService.IsLiked(ctx.Request.Context(), authenticatedUser.ID, userProduct.ID) // Pass context
		if errLike == nil {
			userProduct.IsLiked = isLiked
		}

		// Check if favorited.
		isFavorited, errFav := h.InteractionService.IsFavorited(ctx.Request.Context(), authenticatedUser.ID, userProduct.ID) // Pass context
		if errFav == nil {
			userProduct.IsFavorited = isFavorited
		}
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProduct, ""))
}
