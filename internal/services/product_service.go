package services

import (
	"context" // Added for context
	"fmt"
	"log/slog"

	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"gorm.io/gorm"
)

// Global validator instance (if any, usually not needed in service layer directly)

// ProductService defines the interface for product-related business logic.
type ProductService interface {
	// Basic functionalities
	ListProducts(ctx context.Context, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error)
	GetProduct(ctx context.Context, id uint) (*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product, images []models.ProductImage, categoryIDs []uint) (uint, error)
	UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error
	DeleteProduct(ctx context.Context, id uint) error
}

// productService is the implementation of ProductService.
type productService struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
}

// NewProductService creates a new instance of ProductService.
func NewProductService(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// ListProducts retrieves a list of products.
func (s *productService) ListProducts(ctx context.Context, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error) {
	// Call the repository layer to get the list of products.
	productList, total, err := s.productRepo.ListProducts(ctx, params) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list products", "error", err) // Use slog.ErrorContext
		return nil, nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Return an empty array if there is no data.
	if len(productList) == 0 {
		productList = []models.Product{}
	}

	// Construct pagination information.
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return productList, pagination, nil
}

// GetProduct retrieves details for a single product.
func (s *productService) GetProduct(ctx context.Context, id uint) (*models.Product, error) {
	// Call the repository layer to get the product.
	product, err := s.productRepo.GetProduct(ctx, id) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to get product from repository", "productId", id, "error", err) // Use slog.ErrorContext
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// CreateProduct creates a product and its associated data within a transaction.
func (s *productService) CreateProduct(ctx context.Context, product *models.Product, images []models.ProductImage, categoryIDs []uint) (uint, error) {
	// Validate that the product name is not empty.
	if product.Name == "" {
		return 0, errors.ErrProductNameEmpty
	}

	// Check if all categories exist.
	for _, categoryID := range categoryIDs {
		_, err := s.categoryRepo.GetCategory(ctx, categoryID) // Pass context
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, errors.ErrCategoryNotFound
			}
			slog.ErrorContext(ctx, "Failed to check category for product creation", "categoryID", categoryID, "error", err) // Use slog.ErrorContext
			return 0, fmt.Errorf("failed to check category %d: %w", categoryID, err)
		}
	}

	// Create the product and its associated data within a transaction.
	if err := s.productRepo.CreateProductWithRelations(ctx, product, images, categoryIDs); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to create product with relations", // Use slog.ErrorContext
			"name", product.Name,
			"barcode", product.Barcode,
			"error", err)
		return 0, fmt.Errorf("failed to create product: %w", err)
	}

	return product.ID, nil
}

// UpdateProduct updates a product and its associated data within a transaction.
func (s *productService) UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, id); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product for update", "productID", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Validate that the product name is not empty if it's being updated.
	if name, ok := updates["name"].(string); ok && name == "" {
		return errors.ErrProductNameEmpty
	}

	// Check if all categories exist if they are being updated.
	for _, categoryID := range categoryIDs {
		_, err := s.categoryRepo.GetCategory(ctx, categoryID) // Pass context
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.ErrCategoryNotFound
			}
			slog.ErrorContext(ctx, "Failed to check category for product update", "categoryID", categoryID, "error", err) // Use slog.ErrorContext
			return fmt.Errorf("failed to check category %d: %w", categoryID, err)
		}
	}

	// Update the product and its associated data within a transaction.
	if err := s.productRepo.UpdateProductWithRelations(ctx, id, updates, images, categoryIDs); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to update product with relations", "productId", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct deletes a product.
func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, id); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product for delete", "productID", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Call the repository layer to delete the product.
	if err := s.productRepo.DeleteProduct(ctx, id); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to delete product", "productId", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
