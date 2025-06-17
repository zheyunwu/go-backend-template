package repositories

import (
	"context" // Added for context
	"strconv"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

// ProductRepository defines the interface for product data access operations.
type ProductRepository interface {
	// General CRUD queries
	ListProducts(ctx context.Context, params *query_params.QueryParams) ([]models.Product, int, error)
	GetProduct(ctx context.Context, id uint) (*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteProduct(ctx context.Context, id uint) error

	// Transaction support
	CreateProductWithRelations(ctx context.Context, product *models.Product, images []models.ProductImage, categoryIDs []uint) error
	UpdateProductWithRelations(ctx context.Context, id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error
}

type productRepository struct {
	db           *gorm.DB
	categoryRepo CategoryRepository
}

func NewProductRepository(db *gorm.DB, categoryRepo CategoryRepository) ProductRepository {
	return &productRepository{
		db:           db,
		categoryRepo: categoryRepo,
	}
}

/*
5 general CRUD queries
*/

// ListProducts retrieves a list of products based on query parameters.
func (r *productRepository) ListProducts(ctx context.Context, params *query_params.QueryParams) ([]models.Product, int, error) {
	var products []models.Product
	var totalCount int64

	// Create query.
	query := r.db.WithContext(ctx).Model(&models.Product{}) // Add WithContext

	// Handle search.
	if params.Search != "" {
		query = query.Where("products.name LIKE ? OR products.barcode LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Handle filters.
	if params.Filter != nil {
		for key, value := range params.Filter {
			// Special handling for 'categories' filter.
			if key == "categories" {
				if categoryIDs, ok := value.([]interface{}); ok && len(categoryIDs) > 0 {
					// Create a slice to store numeric IDs.
					var categoryIDsUint []uint

					// Convert interface{} to uint IDs.
					for _, id := range categoryIDs {
						switch v := id.(type) {
						case float64:
							// JSON numbers are parsed as float64 by default.
							categoryIDsUint = append(categoryIDsUint, uint(v))
						case int:
							categoryIDsUint = append(categoryIDsUint, uint(v))
						case uint:
							categoryIDsUint = append(categoryIDsUint, v)
						case string:
							// If it's a string representation of a number, try to convert.
							if numID, err := strconv.ParseUint(v, 10, 64); err == nil {
								categoryIDsUint = append(categoryIDsUint, uint(numID))
							}
						}
					}

					// Expand category ID list to include all children of the specified categories.
					expandedCategoryIDs, err := r.categoryRepo.ExpandCategoryIDsWithChildren(ctx, categoryIDsUint) // Pass context
					if err != nil {
						return nil, 0, err
					}

					// Filter products using the expanded category ID list.
					if len(expandedCategoryIDs) > 0 {
						// Using EXISTS subquery, which is more efficient than JOIN and IN for this case.
						query = query.Where("EXISTS (SELECT 1 FROM product_categories pc WHERE pc.product_id = products.id AND pc.category_id IN ?)", expandedCategoryIDs)
					}
				}
			} else {
				// Handle regular filter conditions.
				query = query.Where("products."+key+" = ?", value)
			}
		}
	}

	// Handle sorting.
	if params.Sort != "" {
		query = query.Order("products." + params.Sort)
	} else {
		query = query.Order("products.updated_at DESC") // Default sort by update time descending.
	}

	// Get total count of records.
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination.
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Preload associated data.
	query = query.Preload("Images").Preload("Categories")

	// Execute query.
	err = query.Find(&products).Error
	return products, int(totalCount), err
}

// GetProduct retrieves a single product by ID, with preloaded associations.
func (r *productRepository) GetProduct(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	// Preload associated data.
	err := r.db.WithContext(ctx).Preload("Images").Preload("Categories").First(&product, id).Error // Add WithContext
	return &product, err
}

// CreateProduct creates a new product.
func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error // Add WithContext
}

// UpdateProduct updates an existing product.
func (r *productRepository) UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error // Add WithContext
}

// DeleteProduct deletes a product (soft delete if DeletedAt field exists in the model).
func (r *productRepository) DeleteProduct(ctx context.Context, id uint) error {
	// GORM automatically performs a soft delete if the model has a gorm.DeletedAt field.
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error // Add WithContext
}

/*
Transaction support
*/

// CreateProductWithRelations creates a product and all its associated data within a single transaction.
func (r *productRepository) CreateProductWithRelations(ctx context.Context, product *models.Product, images []models.ProductImage, categoryIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { // Add WithContext to db for Transaction
		// 1. Create basic product information.
		if err := tx.Create(product).Error; err != nil {
			return err
		}

		// 2. Add product images.
		for i := range images {
			images[i].ProductID = product.ID // Set product ID.
			if err := tx.Create(&images[i]).Error; err != nil {
				return err
			}
		}

		// 3. Add product category associations.
		for _, categoryID := range categoryIDs {
			productCategory := models.ProductCategory{
				ProductID:  product.ID,
				CategoryID: categoryID,
			}
			if err := tx.Create(&productCategory).Error; err != nil {
				return err
			}
		}

		// All operations successful.
		return nil
	})
}

// UpdateProductWithRelations intelligently updates a product and its associated data within a single transaction.
func (r *productRepository) UpdateProductWithRelations(ctx context.Context, id uint, updates map[string]interface{}, newImages []models.ProductImage, categoryIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { // Add WithContext to db for Transaction
		// 1. Update basic product information.
		if len(updates) > 0 {
			if err := tx.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}

		// 2. Intelligently update images - only if a new list of images is provided.
		if newImages != nil {
			// 2.1 Get all current images.
			var existingImages []models.ProductImage
			if err := tx.Where("product_id = ?", id).Find(&existingImages).Error; err != nil {
				return err
			}

			// 2.2 Create a URL map for quick lookup.
			existingImageMap := make(map[string]uint) // URL -> ID
			for _, img := range existingImages {
				existingImageMap[img.ImageURL] = img.ID
			}

			newImageMap := make(map[string]bool) // Record new image URLs.
			for _, img := range newImages {
				newImageMap[img.ImageURL] = true
			}

			// 2.3 Delete images that are no longer needed - use Unscoped() for permanent deletion.
			for url, imgID := range existingImageMap {
				if !newImageMap[url] {
					if err := tx.Unscoped().Delete(&models.ProductImage{}, imgID).Error; err != nil {
						return err
					}
				}
			}

			// 2.4 Add only new images.
			for _, img := range newImages {
				if _, exists := existingImageMap[img.ImageURL]; !exists {
					img.ProductID = id
					if err := tx.Create(&img).Error; err != nil {
						return err
					}
				}
			}
		}

		// 3. Update category associations - only if a new list of category IDs is provided.
		if categoryIDs != nil {
			// 3.1 Get current category associations.
			var existingCategories []models.ProductCategory
			if err := tx.Where("product_id = ?", id).Find(&existingCategories).Error; err != nil {
				return err
			}

			// 3.2 Create a map for quick lookup.
			existingCategoryMap := make(map[uint]bool)
			for _, pc := range existingCategories {
				existingCategoryMap[pc.CategoryID] = true
			}

			newCategoryMap := make(map[uint]bool)
			for _, catID := range categoryIDs {
				newCategoryMap[catID] = true
			}

			// 3.3 Delete category associations that are no longer needed - use Unscoped() for physical deletion.
			for _, pc := range existingCategories {
				if !newCategoryMap[pc.CategoryID] {
					if err := tx.Unscoped().Where("product_id = ? AND category_id = ?",
						id, pc.CategoryID).Delete(&models.ProductCategory{}).Error; err != nil {
						return err
					}
				}
			}

			// 3.4 Add only new category associations.
			for _, catID := range categoryIDs {
				if !existingCategoryMap[catID] {
					productCategory := models.ProductCategory{
						ProductID:  id,
						CategoryID: catID,
					}
					if err := tx.Create(&productCategory).Error; err != nil {
						return err
					}
				}
			}
		}

		// All operations successful.
		return nil
	})
}
