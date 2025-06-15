package services

import (
	"context" // Added for context
	"fmt"
	"log/slog"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// CategoryService defines the interface for category-related business logic.
type CategoryService interface {
	// Basic query methods
	ListCategories(ctx context.Context, params *query_params.QueryParams) ([]models.Category, *response.Pagination, error)
	GetCategoryTree(ctx context.Context, depth int, enabledOnly bool) ([]models.Category, error)
	GetCategory(ctx context.Context, id uint) (*models.Category, error)

	// Management methods - can be implemented as needed later
	// CreateCategory(ctx context.Context, category *models.Category) (uint, error)
	// UpdateCategory(ctx context.Context, id uint, updates map[string]interface{}) error
	// DeleteCategory(ctx context.Context, id uint) error
}

// categoryService is the implementation of CategoryService.
type categoryService struct {
	categoryRepo repositories.CategoryRepository
}

// NewCategoryService creates a new instance of CategoryService.
func NewCategoryService(categoryRepo repositories.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

// ListCategories retrieves a list of categories.
func (s *categoryService) ListCategories(ctx context.Context, params *query_params.QueryParams) ([]models.Category, *response.Pagination, error) {
	// Call the repository layer to get the list of categories.
	categories, total, err := s.categoryRepo.ListCategories(ctx, params) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list categories", "error", err) // Use slog.ErrorContext
		return nil, nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// Return an empty array if there is no data.
	if len(categories) == 0 {
		categories = []models.Category{}
	}

	// Construct pagination information.
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return categories, pagination, nil
}

// GetCategoryTree retrieves the category tree structure.
func (s *categoryService) GetCategoryTree(ctx context.Context, depth int, enabledOnly bool) ([]models.Category, error) {
	// Call the repository layer to get the category tree, passing depth and whether to show all categories.
	categories, err := s.categoryRepo.GetCategoryTree(ctx, depth, enabledOnly) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get category tree", "error", err, "depth", depth, "enabledOnly", enabledOnly) // Use slog.ErrorContext
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}

	// Return an empty array if there is no data.
	if len(categories) == 0 {
		categories = []models.Category{}
	}

	return categories, nil
}

// GetCategory retrieves details for a single category.
func (s *categoryService) GetCategory(ctx context.Context, id uint) (*models.Category, error) {
	// Call the repository layer to get the category.
	category, err := s.categoryRepo.GetCategory(ctx, id) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get category", "categoryId", id, "error", err) // Use slog.ErrorContext
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}
