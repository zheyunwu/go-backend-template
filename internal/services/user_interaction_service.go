package services

import (
	"fmt"
	"log/slog"

	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"gorm.io/gorm"
)

// UserInteractionService 用户交互服务接口
type UserInteractionService interface {
	// 点赞相关
	AddLike(userID, productID uint) error
	RemoveLike(userID, productID uint) error
	IsLiked(userID, productID uint) (bool, error)
	ListUserLikedProducts(userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error)

	// 收藏相关
	AddFavorite(userID, productID uint) error
	RemoveFavorite(userID, productID uint) error
	IsFavorited(userID, productID uint) (bool, error)
	ListUserFavoritedProducts(userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error)

	// 统计功能
	GetProductLikeCount(productID uint) (int, error)
	GetProductFavoriteCount(productID uint) (int, error)
}

// userInteractionService 用户交互服务实现
type userInteractionService struct {
	interactionRepo repositories.UserInteractionRepository
	productRepo     repositories.ProductRepository
}

// NewUserInteractionService 创建用户交互服务实例
func NewUserInteractionService(
	interactionRepo repositories.UserInteractionRepository,
	productRepo repositories.ProductRepository,
) UserInteractionService {
	return &userInteractionService{
		interactionRepo: interactionRepo,
		productRepo:     productRepo,
	}
}

// AddLike 添加点赞
func (s *userInteractionService) AddLike(userID, productID uint) error {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// 添加点赞
	if err := s.interactionRepo.AddLike(userID, productID); err != nil {
		slog.Error("Failed to add like", "userId", userID, "productId", productID, "error", err)
		return fmt.Errorf("failed to add like: %w", err)
	}

	return nil
}

// RemoveLike 取消点赞
func (s *userInteractionService) RemoveLike(userID, productID uint) error {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// 取消点赞
	if err := s.interactionRepo.RemoveLike(userID, productID); err != nil {
		slog.Error("Failed to remove like", "userId", userID, "productId", productID, "error", err)
		return fmt.Errorf("failed to remove like: %w", err)
	}

	return nil
}

// IsLiked 检查是否已点赞
func (s *userInteractionService) IsLiked(userID, productID uint) (bool, error) {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errors.ErrProductNotFound
		}
		return false, fmt.Errorf("failed to check product: %w", err)
	}

	// 检查是否已点赞
	isLiked, err := s.interactionRepo.IsLiked(userID, productID)
	if err != nil {
		slog.Error("Failed to check like status", "userId", userID, "productId", productID, "error", err)
		return false, fmt.Errorf("failed to check like status: %w", err)
	}

	return isLiked, nil
}

// ListUserLikedProducts 获取用户点赞的产品
func (s *userInteractionService) ListUserLikedProducts(userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error) {
	// 获取用户点赞列表
	products, total, err := s.interactionRepo.ListUserInteractedProducts(userID, params, "like")
	if err != nil {
		slog.Error("Failed to list user likes", "userId", userID, "error", err)
		return nil, nil, fmt.Errorf("failed to list user likes: %w", err)
	}

	// 构建分页信息
	pagination := &response.Pagination{
		TotalCount:  total,
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (total + params.Limit - 1) / params.Limit,
	}

	// 如果没有数据，返回空数组而不是nil
	if products == nil {
		products = []models.Product{}
	}

	return products, pagination, nil
}

// AddFavorite 添加收藏
func (s *userInteractionService) AddFavorite(userID, productID uint) error {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// 添加收藏
	if err := s.interactionRepo.AddFavorite(userID, productID); err != nil {
		slog.Error("Failed to add favorite", "userId", userID, "productId", productID, "error", err)
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

// RemoveFavorite 取消收藏
func (s *userInteractionService) RemoveFavorite(userID, productID uint) error {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// 取消收藏
	if err := s.interactionRepo.RemoveFavorite(userID, productID); err != nil {
		slog.Error("Failed to remove favorite", "userId", userID, "productId", productID, "error", err)
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	return nil
}

// IsFavorited 检查是否已收藏
func (s *userInteractionService) IsFavorited(userID, productID uint) (bool, error) {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errors.ErrProductNotFound
		}
		return false, fmt.Errorf("failed to check product: %w", err)
	}

	// 检查是否已收藏
	isFavorited, err := s.interactionRepo.IsFavorited(userID, productID)
	if err != nil {
		slog.Error("Failed to check favorite status", "userId", userID, "productId", productID, "error", err)
		return false, fmt.Errorf("failed to check favorite status: %w", err)
	}

	return isFavorited, nil
}

// ListUserFavoritedProducts 获取用户所有收藏
func (s *userInteractionService) ListUserFavoritedProducts(userID uint, params *query_params.QueryParams) ([]models.Product, *response.Pagination, error) {
	// 获取用户收藏列表
	products, total, err := s.interactionRepo.ListUserInteractedProducts(userID, params, "favorite")
	if err != nil {
		slog.Error("Failed to list user favorites", "userId", userID, "error", err)
		return nil, nil, fmt.Errorf("failed to list user favorites: %w", err)
	}

	// 构建分页信息
	pagination := &response.Pagination{
		TotalCount:  total,
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (total + params.Limit - 1) / params.Limit,
	}

	// 如果没有数据，返回空数组而不是nil
	if products == nil {
		products = []models.Product{}
	}

	return products, pagination, nil
}

// GetProductLikeCount 获取产品的点赞数量
func (s *userInteractionService) GetProductLikeCount(productID uint) (int, error) {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.ErrProductNotFound
		}
		return 0, fmt.Errorf("failed to check product: %w", err)
	}

	// 获取点赞数量
	count, err := s.interactionRepo.GetProductLikeCount(productID)
	if err != nil {
		slog.Error("Failed to get product like count", "productId", productID, "error", err)
		return 0, fmt.Errorf("failed to get product like count: %w", err)
	}

	return count, nil
}

// GetProductFavoriteCount 获取产品的收藏数量
func (s *userInteractionService) GetProductFavoriteCount(productID uint) (int, error) {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(productID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.ErrProductNotFound
		}
		return 0, fmt.Errorf("failed to check product: %w", err)
	}

	// 获取收藏数量
	count, err := s.interactionRepo.GetProductFavoriteCount(productID)
	if err != nil {
		slog.Error("Failed to get product favorite count", "productId", productID, "error", err)
		return 0, fmt.Errorf("failed to get product favorite count: %w", err)
	}

	return count, nil
}
