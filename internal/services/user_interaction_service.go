package services

import (
	"context" // Added for context
	"fmt"
	"log/slog"

	"github.com/go-backend-template/internal/dto" // Added for UserInteractionStatus
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"gorm.io/gorm"
)

// UserInteractionService defines the interface for user interaction services.
type UserInteractionService interface {
	// Like related methods
	AddLike(ctx context.Context, userID, productID uint) error
	RemoveLike(ctx context.Context, userID, productID uint) error
	IsLiked(ctx context.Context, userID, productID uint) (bool, error)
	ListUserLikedProducts(ctx context.Context, userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error)

	// Favorite related methods
	AddFavorite(ctx context.Context, userID, productID uint) error
	RemoveFavorite(ctx context.Context, userID, productID uint) error
	IsFavorited(ctx context.Context, userID, productID uint) (bool, error)
	ListUserFavoritedProducts(ctx context.Context, userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error)

	// Statistics
	GetProductLikeCount(ctx context.Context, productID uint) (int, error)
	GetProductFavoriteCount(ctx context.Context, productID uint) (int, error)

	// Get interaction status for multiple products in bulk
	GetUserProductInteractionStatus(ctx context.Context, userID uint, productIDs []uint) (map[uint]dto.UserInteractionStatus, error)
}

// userInteractionService is the implementation of UserInteractionService.
type userInteractionService struct {
	interactionRepo repositories.UserInteractionRepository
	productRepo     repositories.ProductRepository
}

// NewUserInteractionService creates a new instance of UserInteractionService.
func NewUserInteractionService(
	interactionRepo repositories.UserInteractionRepository,
	productRepo repositories.ProductRepository,
) UserInteractionService {
	return &userInteractionService{
		interactionRepo: interactionRepo,
		productRepo:     productRepo,
	}
}

// AddLike adds a like for a product by a user.
func (s *userInteractionService) AddLike(ctx context.Context, userID, productID uint) error {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before adding like", "productID", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Add the like.
	if err := s.interactionRepo.AddLike(ctx, userID, productID); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to add like", "userId", userID, "productId", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to add like: %w", err)
	}

	return nil
}

// RemoveLike removes a like for a product by a user.
func (s *userInteractionService) RemoveLike(ctx context.Context, userID, productID uint) error {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before removing like", "productID", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Remove the like.
	if err := s.interactionRepo.RemoveLike(ctx, userID, productID); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to remove like", "userId", userID, "productId", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to remove like: %w", err)
	}

	return nil
}

// IsLiked checks if a product is liked by a user.
func (s *userInteractionService) IsLiked(ctx context.Context, userID, productID uint) (bool, error) {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			// It's not an error if the product doesn't exist, just means the user hasn't liked it.
			// However, to maintain consistency with other methods, we can return ErrProductNotFound.
			// Or, decide if this check is strictly necessary for "IsLiked".
			// For now, returning error to be consistent.
			slog.WarnContext(ctx, "Product not found when checking if liked", "productID", productID, "userID", userID)
			return false, errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before checking like status", "productID", productID, "error", err) // Use slog.ErrorContext
		return false, fmt.Errorf("failed to check product: %w", err)
	}

	// Check if liked.
	isLiked, err := s.interactionRepo.IsLiked(ctx, userID, productID) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check like status", "userId", userID, "productId", productID, "error", err) // Use slog.ErrorContext
		return false, fmt.Errorf("failed to check like status: %w", err)
	}

	return isLiked, nil
}

// ListUserLikedProducts retrieves products liked by a user.
func (s *userInteractionService) ListUserLikedProducts(ctx context.Context, userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error) {
	// Get the list of liked products.
	products, total, err := s.interactionRepo.ListUserInteractedProducts(ctx, userID, params, "like") // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list user likes", "userId", userID, "error", err) // Use slog.ErrorContext
		return nil, nil, fmt.Errorf("failed to list user likes: %w", err)
	}

	// Construct pagination information.
	pagination := &response.Pagination{
		TotalCount:  total,
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (total + params.Limit - 1) / params.Limit,
	}

	// Return an empty array if there is no data, instead of nil.
	if products == nil {
		products = []models.Product{}
	}

	return products, pagination, nil
}

// AddFavorite adds a favorite for a product by a user.
func (s *userInteractionService) AddFavorite(ctx context.Context, userID, productID uint) error {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before adding favorite", "productID", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Add the favorite.
	if err := s.interactionRepo.AddFavorite(ctx, userID, productID); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to add favorite", "userId", userID, "productId", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

// RemoveFavorite removes a favorite for a product by a user.
func (s *userInteractionService) RemoveFavorite(ctx context.Context, userID, productID uint) error {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before removing favorite", "productID", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Remove the favorite.
	if err := s.interactionRepo.RemoveFavorite(ctx, userID, productID); err != nil { // Pass context
		slog.ErrorContext(ctx, "Failed to remove favorite", "userId", userID, "productId", productID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	return nil
}

// IsFavorited checks if a product is favorited by a user.
func (s *userInteractionService) IsFavorited(ctx context.Context, userID, productID uint) (bool, error) {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			slog.WarnContext(ctx, "Product not found when checking if favorited", "productID", productID, "userID", userID)
			return false, errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before checking favorite status", "productID", productID, "error", err) // Use slog.ErrorContext
		return false, fmt.Errorf("failed to check product: %w", err)
	}

	// Check if favorited.
	isFavorited, err := s.interactionRepo.IsFavorited(ctx, userID, productID) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check favorite status", "userId", userID, "productId", productID, "error", err) // Use slog.ErrorContext
		return false, fmt.Errorf("failed to check favorite status: %w", err)
	}

	return isFavorited, nil
}

// ListUserFavoritedProducts retrieves products favorited by a user.
func (s *userInteractionService) ListUserFavoritedProducts(ctx context.Context, userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error) {
	// Get the list of favorited products.
	products, total, err := s.interactionRepo.ListUserInteractedProducts(ctx, userID, params, "favorite") // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list user favorites", "userId", userID, "error", err) // Use slog.ErrorContext
		return nil, nil, fmt.Errorf("failed to list user favorites: %w", err)
	}

	// Construct pagination information.
	pagination := &response.Pagination{
		TotalCount:  total,
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (total + params.Limit - 1) / params.Limit,
	}

	// Return an empty array if there is no data, instead of nil.
	if products == nil {
		products = []models.Product{}
	}

	return products, pagination, nil
}

// GetProductLikeCount gets the like count for a product.
func (s *userInteractionService) GetProductLikeCount(ctx context.Context, productID uint) (int, error) {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			slog.WarnContext(ctx, "Product not found when getting like count", "productID", productID)
			return 0, errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before getting like count", "productID", productID, "error", err) // Use slog.ErrorContext
		return 0, fmt.Errorf("failed to check product: %w", err)
	}

	// Get the like count.
	count, err := s.interactionRepo.GetProductLikeCount(ctx, productID) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get product like count", "productId", productID, "error", err) // Use slog.ErrorContext
		return 0, fmt.Errorf("failed to get product like count: %w", err)
	}

	return count, nil
}

// GetUserProductInteractionStatus gets like and favorite status for a list of products by a user in bulk.
func (s *userInteractionService) GetUserProductInteractionStatus(ctx context.Context, userID uint, productIDs []uint) (map[uint]dto.UserInteractionStatus, error) {
	if len(productIDs) == 0 {
		return map[uint]dto.UserInteractionStatus{}, nil
	}

	// Get like statuses in bulk.
	likedMap, err := s.interactionRepo.GetLikedProductIDs(ctx, userID, productIDs)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get liked product IDs", "userID", userID, "error", err)
		return nil, fmt.Errorf("failed to get liked product IDs: %w", err)
	}

	// Get favorite statuses in bulk.
	favoritedMap, err := s.interactionRepo.GetFavoritedProductIDs(ctx, userID, productIDs)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get favorited product IDs", "userID", userID, "error", err)
		return nil, fmt.Errorf("failed to get favorited product IDs: %w", err)
	}

	// Merge results.
	interactionStatusMap := make(map[uint]dto.UserInteractionStatus)
	for _, productID := range productIDs {
		status := dto.UserInteractionStatus{
			IsLiked:     likedMap[productID],     // Defaults to false if productID is not in likedMap
			IsFavorited: favoritedMap[productID], // Defaults to false if productID is not in favoritedMap
		}
		interactionStatusMap[productID] = status
	}

	return interactionStatusMap, nil
}

// GetProductFavoriteCount gets the favorite count for a product.
func (s *userInteractionService) GetProductFavoriteCount(ctx context.Context, productID uint) (int, error) {
	// Check if the product exists.
	if _, err := s.productRepo.GetProduct(ctx, productID); err != nil { // Pass context
		if err == gorm.ErrRecordNotFound {
			slog.WarnContext(ctx, "Product not found when getting favorite count", "productID", productID)
			return 0, errors.ErrProductNotFound
		}
		slog.ErrorContext(ctx, "Failed to check product before getting favorite count", "productID", productID, "error", err) // Use slog.ErrorContext
		return 0, fmt.Errorf("failed to check product: %w", err)
	}

	// Get the favorite count.
	count, err := s.interactionRepo.GetProductFavoriteCount(ctx, productID) // Pass context
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get product favorite count", "productId", productID, "error", err) // Use slog.ErrorContext
		return 0, fmt.Errorf("failed to get product favorite count: %w", err)
	}

	return count, nil
}
