package repositories

import (
	"context" // Added for context
	"strconv"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

// UserInteractionRepository defines the interface for user interaction data access operations.
type UserInteractionRepository interface {
	// Like related methods
	AddLike(ctx context.Context, userID, productID uint) error
	RemoveLike(ctx context.Context, userID, productID uint) error
	IsLiked(ctx context.Context, userID, productID uint) (bool, error)

	// Favorite related methods
	AddFavorite(ctx context.Context, userID, productID uint) error
	RemoveFavorite(ctx context.Context, userID, productID uint) error
	IsFavorited(ctx context.Context, userID, productID uint) (bool, error)

	// General list method for interacted products
	ListUserInteractedProducts(ctx context.Context, userID uint, params *query_params.QueryParams, interactionType string) ([]models.Product, int, error)

	// Statistics
	GetProductLikeCount(ctx context.Context, productID uint) (int, error)
	GetProductFavoriteCount(ctx context.Context, productID uint) (int, error)

	// Bulk queries
	GetLikedProductIDs(ctx context.Context, userID uint, productIDs []uint) (map[uint]bool, error)
	GetFavoritedProductIDs(ctx context.Context, userID uint, productIDs []uint) (map[uint]bool, error)
}

// userInteractionRepository is the implementation for user interaction data access.
type userInteractionRepository struct {
	db           *gorm.DB
	categoryRepo CategoryRepository
}

// NewUserInteractionRepository creates a new instance of UserInteractionRepository.
func NewUserInteractionRepository(db *gorm.DB, categoryRepo CategoryRepository) UserInteractionRepository {
	return &userInteractionRepository{
		db:           db,
		categoryRepo: categoryRepo,
	}
}

// AddLike adds a like for a product by a user.
func (r *userInteractionRepository) AddLike(ctx context.Context, userID, productID uint) error {
	like := models.UserProductLike{
		UserID:    userID,
		ProductID: productID,
	}
	// Use FirstOrCreate to avoid duplicate likes.
	return r.db.WithContext(ctx).Where(models.UserProductLike{UserID: userID, ProductID: productID}). // Add WithContext
		FirstOrCreate(&like).Error
}

// RemoveLike removes a like for a product by a user.
func (r *userInteractionRepository) RemoveLike(ctx context.Context, userID, productID uint) error {
	// Delete using primary key conditions.
	return r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID). // Add WithContext
		Delete(&models.UserProductLike{}).Error
}

// IsLiked checks if a product is liked by a user.
func (r *userInteractionRepository) IsLiked(ctx context.Context, userID, productID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserProductLike{}). // Add WithContext
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// AddFavorite adds a favorite for a product by a user.
func (r *userInteractionRepository) AddFavorite(ctx context.Context, userID, productID uint) error {
	favorite := models.UserProductFavorite{
		UserID:    userID,
		ProductID: productID,
	}
	// Use FirstOrCreate to avoid duplicate favorites.
	return r.db.WithContext(ctx).Where(models.UserProductFavorite{UserID: userID, ProductID: productID}). // Add WithContext
		FirstOrCreate(&favorite).Error
}

// RemoveFavorite removes a favorite for a product by a user.
func (r *userInteractionRepository) RemoveFavorite(ctx context.Context, userID, productID uint) error {
	// Delete using primary key conditions.
	return r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID). // Add WithContext
		Delete(&models.UserProductFavorite{}).Error
}

// IsFavorited checks if a product is favorited by a user.
func (r *userInteractionRepository) IsFavorited(ctx context.Context, userID, productID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserProductFavorite{}). // Add WithContext
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// ListUserInteractedProducts retrieves products liked or favorited by a user.
// interactionType can be "like" or "favorite".
func (r *userInteractionRepository) ListUserInteractedProducts(ctx context.Context, userID uint, params *query_params.QueryParams, interactionType string) ([]models.Product, int, error) {
	var products []models.Product
	var total int64
	var tableName, orderField string

	// Determine table name and sort field based on interaction type.
	switch interactionType {
	case "like":
		tableName = "user_product_likes"
		orderField = "user_product_likes.created_at"
	case "favorite":
		tableName = "user_product_favorites"
		orderField = "user_product_favorites.created_at"
	default:
		return nil, 0, nil // Or return an error for invalid interactionType
	}

	// Create query.
	query := r.db.WithContext(ctx).Table("products"). // Add WithContext
		Joins("JOIN "+tableName+" ON products.id = "+tableName+".product_id").
		Where(tableName+".user_id = ? AND products.deleted_at IS NULL", userID)

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
		query = query.Order(orderField + " DESC") // Default sort by interaction time descending.
	}

	// Get total count of records.
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination.
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Preload associated data.
	query = query.Preload("Images").Preload("Categories")

	err = query.Find(&products).Error

	return products, int(total), err
}

// GetProductLikeCount retrieves the like count for a product.
func (r *userInteractionRepository) GetProductLikeCount(ctx context.Context, productID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserProductLike{}). // Add WithContext
		Where("product_id = ?", productID).
		Count(&count).Error
	return int(count), err
}

// GetLikedProductIDs retrieves a map of product IDs liked by a user from a given list of product IDs.
func (r *userInteractionRepository) GetLikedProductIDs(ctx context.Context, userID uint, productIDs []uint) (map[uint]bool, error) {
	if len(productIDs) == 0 {
		return map[uint]bool{}, nil
	}
	var likedProducts []models.UserProductLike
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id IN ?", userID, productIDs).
		Find(&likedProducts).Error
	if err != nil {
		return nil, err
	}

	likedMap := make(map[uint]bool)
	for _, p := range likedProducts {
		likedMap[p.ProductID] = true
	}
	return likedMap, nil
}

// GetFavoritedProductIDs retrieves a map of product IDs favorited by a user from a given list of product IDs.
func (r *userInteractionRepository) GetFavoritedProductIDs(ctx context.Context, userID uint, productIDs []uint) (map[uint]bool, error) {
	if len(productIDs) == 0 {
		return map[uint]bool{}, nil
	}
	var favoritedProducts []models.UserProductFavorite
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id IN ?", userID, productIDs).
		Find(&favoritedProducts).Error
	if err != nil {
		return nil, err
	}

	favoritedMap := make(map[uint]bool)
	for _, p := range favoritedProducts {
		favoritedMap[p.ProductID] = true
	}
	return favoritedMap, nil
}

// GetProductFavoriteCount retrieves the favorite count for a product.
func (r *userInteractionRepository) GetProductFavoriteCount(ctx context.Context, productID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserProductFavorite{}). // Add WithContext
		Where("product_id = ?", productID).
		Count(&count).Error
	return int(count), err
}
