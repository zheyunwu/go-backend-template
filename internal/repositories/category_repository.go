package repositories

import (
	"context" // Added for context
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

// CategoryRepository defines the interface for category data access operations.
type CategoryRepository interface {
	// General CRUD queries
	ListCategories(ctx context.Context, params *query_params.QueryParams) ([]models.Category, int, error)
	GetCategory(ctx context.Context, id uint) (*models.Category, error)
	CreateCategory(ctx context.Context, category *models.Category) error
	UpdateCategory(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteCategory(ctx context.Context, id uint) error

	// Custom queries
	GetCategoryTree(ctx context.Context, depth int, enabledOnly bool) ([]models.Category, error)
	GetChildCategories(ctx context.Context, parentID uint) ([]models.Category, error)
	GetCategoryProducts(ctx context.Context, categoryID uint, params *query_params.QueryParams) ([]models.Product, int, error)

	// Utility methods for other repositories
	ExpandCategoryIDsWithChildren(ctx context.Context, categoryIDs []uint) ([]uint, error)
	GetAllChildCategoryIDs(ctx context.Context, parentID uint) ([]uint, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

/*
General CRUD queries
*/

// ListCategories retrieves a list of categories.
func (r *categoryRepository) ListCategories(ctx context.Context, params *query_params.QueryParams) ([]models.Category, int, error) {
	var categories []models.Category
	var totalCount int64

	// Create query.
	query := r.db.WithContext(ctx).Model(&models.Category{}) // Add WithContext

	// Handle search.
	if params.Search != "" {
		query = query.Where("name LIKE ? OR name_zh LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Handle filters.
	if params.Filter != nil {
		for key, value := range params.Filter {
			query = query.Where(key+" = ?", value)
		}
	}

	// Handle sorting.
	if params.Sort != "" {
		query = query.Order(params.Sort)
	} else {
		query = query.Order("name ASC") // Default sort by name ascending.
	}

	// Get total count.
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination.
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Execute query.
	err = query.Find(&categories).Error
	return categories, int(totalCount), err
}

// GetCategory retrieves a single category by ID.
func (r *categoryRepository) GetCategory(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	// Preload child categories if needed.
	err := r.db.WithContext(ctx).First(&category, id).Error // Add WithContext
	return &category, err
}

// CreateCategory creates a new category.
func (r *categoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error // Add WithContext
}

// UpdateCategory updates an existing category.
func (r *categoryRepository) UpdateCategory(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Category{}).Where("id = ?", id).Updates(updates).Error // Add WithContext
}

// DeleteCategory deletes a category (soft delete).
func (r *categoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error // Add WithContext
}

/*
Custom queries
*/

// GetCategoryTree retrieves the category tree structure.
func (r *categoryRepository) GetCategoryTree(ctx context.Context, depth int, enabledOnly bool) ([]models.Category, error) {
	// First, get all root categories (those without a parent).
	var rootCategories []models.Category
	query := r.db.WithContext(ctx).Where("parent_id IS NULL") // Add WithContext

	// The enabledOnly parameter determines whether to show only enabled categories.
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	if err := query.Find(&rootCategories).Error; err != nil {
		return nil, err
	}

	// Recursively get child categories for each root category, passing current depth as 1.
	for i := range rootCategories {
		if err := r.loadChildCategoriesWithDepth(ctx, &rootCategories[i], 1, depth, enabledOnly); err != nil { // Pass context
			return nil, err
		}
	}

	return rootCategories, nil
}

// loadChildCategoriesWithDepth recursively loads child categories with a depth limit.
func (r *categoryRepository) loadChildCategoriesWithDepth(ctx context.Context, category *models.Category, currentDepth, maxDepth int, enabledOnly bool) error {
	// If the maximum depth is reached or if depth is unlimited (maxDepth <= 0), do not load further children.
	if maxDepth > 0 && currentDepth >= maxDepth {
		category.Children = []models.Category{} // Set to an empty slice instead of nil.
		return nil
	}

	var children []models.Category
	query := r.db.WithContext(ctx).Where("parent_id = ?", category.ID) // Add WithContext

	// Based on enabledOnly parameter, decide whether to show only enabled categories.
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	if err := query.Find(&children).Error; err != nil {
		return err
	}

	// Continue to recursively load children for each child category.
	for i := range children {
		if err := r.loadChildCategoriesWithDepth(ctx, &children[i], currentDepth+1, maxDepth, enabledOnly); err != nil { // Pass context
			return err
		}
	}

	// Set child categories.
	category.Children = children
	return nil
}

// GetChildCategories retrieves direct child categories.
func (r *categoryRepository) GetChildCategories(ctx context.Context, parentID uint) ([]models.Category, error) {
	var children []models.Category
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&children).Error // Add WithContext
	return children, err
}

// GetCategoryProducts retrieves products under a category.
func (r *categoryRepository) GetCategoryProducts(ctx context.Context, categoryID uint, params *query_params.QueryParams) ([]models.Product, int, error) {
	var products []models.Product
	var totalCount int64

	// Query products for a specific category using a join table.
	query := r.db.WithContext(ctx).Table("products"). // Add WithContext
		Joins("JOIN product_categories ON products.id = product_categories.product_id").
		Where("product_categories.category_id = ?", categoryID)

	// Handle search.
	if params.Search != "" {
		query = query.Where("products.name LIKE ? OR products.barcode LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Handle filters.
	if params.Filter != nil {
		for key, value := range params.Filter {
			query = query.Where("products."+key+" = ?", value)
		}
	}

	// Handle sorting.
	if params.Sort != "" {
		query = query.Order("products." + params.Sort)
	} else {
		query = query.Order("products.updated_at DESC") // Default sort by update time descending.
	}

	// Get total count.
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination.
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Preload associated data.
	query = query.Preload("Images").Preload("Retailers").Preload("Categories")

	// Execute query.
	err = query.Find(&products).Error
	return products, int(totalCount), err
}

/*
Utility methods
*/

// ExpandCategoryIDsWithChildren retrieves IDs of all specified categories and their children.
func (r *categoryRepository) ExpandCategoryIDsWithChildren(ctx context.Context, categoryIDs []uint) ([]uint, error) {
	// If no category IDs, return an empty result.
	if len(categoryIDs) == 0 {
		return []uint{}, nil
	}

	// Create a result set, initially containing the original category IDs.
	expandedIDs := make(map[uint]bool)
	for _, id := range categoryIDs {
		expandedIDs[id] = true
	}

	// Recursively find all child categories.
	for _, id := range categoryIDs {
		childIDs, err := r.GetAllChildCategoryIDs(ctx, id) // Pass context
		if err != nil {
			return nil, err
		}

		// Add child category IDs to the result set.
		for _, childID := range childIDs {
			expandedIDs[childID] = true
		}
	}

	// Convert map to slice.
	result := make([]uint, 0, len(expandedIDs))
	for id := range expandedIDs {
		result = append(result, id)
	}

	return result, nil
}

// GetAllChildCategoryIDs recursively retrieves all child category IDs for a given parent.
func (r *categoryRepository) GetAllChildCategoryIDs(ctx context.Context, parentID uint) ([]uint, error) {
	var childIDs []uint
	var childCategories []models.Category

	// Query direct child categories.
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&childCategories).Error; err != nil { // Add WithContext
		return nil, err
	}

	// If no child categories, return an empty result.
	if len(childCategories) == 0 {
		return []uint{}, nil
	}

	// Collect child category IDs.
	for _, child := range childCategories {
		childIDs = append(childIDs, child.ID)

		// Recursively get grandchildren IDs.
		grandChildIDs, err := r.GetAllChildCategoryIDs(ctx, child.ID) // Pass context
		if err != nil {
			return nil, err
		}

		// Add to results.
		childIDs = append(childIDs, grandChildIDs...)
	}

	return childIDs, nil
}
