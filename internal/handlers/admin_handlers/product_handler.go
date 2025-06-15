package admin_handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/internal/utils" // Added validator utility
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

var customValidator = utils.NewCustomValidator() // Create a validator instance

type ProductHandler struct {
	ProductService services.ProductService
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(productService services.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductService: productService,
	}
}

// ListProducts retrieves a list of products.
func (h *ProductHandler) ListProducts(ctx *gin.Context) {
	// Get parsed query parameters from context.
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		slog.Warn("Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// Get product list.
	products, pagination, err := h.ProductService.ListProducts(ctx.Request.Context(), queryParams) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(products, "", *pagination))
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

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(product, ""))
}

// CreateProduct creates a new product.
func (h *ProductHandler) CreateProduct(ctx *gin.Context) {
	// Parse request body to DTO.
	var createReq dto.CreateProductRequest
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		slog.Warn("Invalid product creation request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request parameters: "+err.Error())) // "请求参数错误: " -> "Invalid request parameters: "
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&createReq); validationErrs != nil {
		slog.Warn("Validation failed for CreateProduct", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Convert DTO to product model.
	product := createReq.ToModel()

	// Convert image URLs to ProductImage models.
	var images []models.ProductImage
	for _, url := range createReq.ImageURLs {
		if url != "" {
			images = append(images, models.ProductImage{
				ImageURL: url,
			})
		}
	}

	// Call service layer's creation method, passing all model objects.
	createdProductID, err := h.ProductService.CreateProduct( // Pass context
		ctx.Request.Context(),
		product,
		images,
		createReq.CategoryIDs,
	)

	// Handle error - now only one error is returned due to transactions.
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 201 Created.
	slog.Info("Product created successfully", "productId", createdProductID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdProductID}, ""))
}

// UpdateProduct updates product information.
func (h *ProductHandler) UpdateProduct(ctx *gin.Context) {
	// Parse product ID.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Parse request body to DTO.
	var updateReq dto.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		slog.Warn("Invalid product update request", "productId", id, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request parameters: "+err.Error())) // "请求参数错误: " -> "Invalid request parameters: "
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&updateReq); validationErrs != nil {
		slog.Warn("Validation failed for UpdateProduct", "productId", id, "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Convert DTO to update map.
	updates := updateReq.ToMap()

	// Convert image URLs to ProductImage models.
	var images []models.ProductImage
	if updateReq.ImageURLs != nil { // Only process if image array is included in the request.
		if len(updateReq.ImageURLs) == 0 {
			// If ImageURLs length is 0, set images to an empty array, but not nil.
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

	// If there's nothing to update, return success directly.
	if len(updates) == 0 && updateReq.ImageURLs == nil && updateReq.CategoryIDs == nil {
		slog.Info("No fields to update", "productId", id)
		ctx.JSON(http.StatusNoContent, nil)
		return
	}

	// Call service layer's update method, passing all model objects.
	err = h.ProductService.UpdateProduct( // Pass context
		ctx.Request.Context(),
		uint(id),
		updates,
		images,
		updateReq.CategoryIDs,
	)

	// Handle error - now only one error is returned due to transactions.
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 204 No Content.
	slog.Info("Product updated successfully", "productId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

// DeleteProduct deletes a product.
func (h *ProductHandler) DeleteProduct(ctx *gin.Context) {
	// Parse product ID.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Call service layer to delete the product.
	err = h.ProductService.DeleteProduct(ctx.Request.Context(), uint(id)) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 204 No Content.
	slog.Info("Product deleted", "productId", id)
	ctx.JSON(http.StatusNoContent, nil)
}
