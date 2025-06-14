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

// 全局验证器实例

// ProductService 定义产品相关的业务逻辑接口
type ProductService interface {
	// 基本功能
	ListProducts(params *query_params.QueryParams) ([]models.Product, *response.Pagination, error)
	GetProduct(id uint) (*models.Product, error)
	CreateProduct(product *models.Product, images []models.ProductImage, categoryIDs []uint) (uint, error)
	UpdateProduct(id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error
	DeleteProduct(id uint) error
}

// productService 产品服务实现
type productService struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
}

// NewProductService 创建一个产品服务实例
func NewProductService(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// ListProducts 获取产品列表
func (s *productService) ListProducts(params *query_params.QueryParams) ([]models.Product, *response.Pagination, error) {
	// 调用Repo层获取产品列表
	productList, total, err := s.productRepo.ListProducts(params)
	if err != nil {
		slog.Error("Failed to list products", "error", err)
		return nil, nil, fmt.Errorf("failed to list products: %w", err)
	}

	// 没有数据时返回空数组
	if len(productList) == 0 {
		productList = []models.Product{}
	}

	// 构建分页信息
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return productList, pagination, nil
}

// GetProduct 获取产品详情
func (s *productService) GetProduct(id uint) (*models.Product, error) {
	// 调用repo层获取产品
	product, err := s.productRepo.GetProduct(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrProductNotFound
		}
		slog.Error("Failed to get product from repository", "productId", id, "error", err)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// CreateProduct 使用事务创建产品及其关联数据
func (s *productService) CreateProduct(product *models.Product, images []models.ProductImage, categoryIDs []uint) (uint, error) {
	// 验证产品名称是否为空
	if product.Name == "" {
		return 0, errors.ErrProductNameEmpty
	}

	// 检查所有分类是否存在
	for _, categoryID := range categoryIDs {
		_, err := s.categoryRepo.GetCategory(categoryID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, errors.ErrCategoryNotFound
			}
			return 0, fmt.Errorf("failed to check category %d: %w", categoryID, err)
		}
	}

	// 使用事务创建产品及其关联数据
	if err := s.productRepo.CreateProductWithRelations(product, images, categoryIDs); err != nil {
		slog.Error("Failed to create product with relations",
			"name", product.Name,
			"barcode", product.Barcode,
			"error", err)
		return 0, fmt.Errorf("failed to create product: %w", err)
	}

	return product.ID, nil
}

// UpdateProduct 使用事务更新产品及其关联数据
func (s *productService) UpdateProduct(id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// 验证产品名称是否为空
	if name, ok := updates["name"].(string); ok && name == "" {
		return errors.ErrProductNameEmpty
	}

	// 检查所有分类是否存在
	for _, categoryID := range categoryIDs {
		_, err := s.categoryRepo.GetCategory(categoryID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.ErrCategoryNotFound
			}
			return fmt.Errorf("failed to check category %d: %w", categoryID, err)
		}
	}

	// 使用事务更新产品及其关联数据
	if err := s.productRepo.UpdateProductWithRelations(id, updates, images, categoryIDs); err != nil {
		slog.Error("Failed to update product with relations", "productId", id, "error", err)
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct 删除产品
func (s *productService) DeleteProduct(id uint) error {
	// 检查产品是否存在
	if _, err := s.productRepo.GetProduct(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProductNotFound
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// 调用repo层删除产品
	if err := s.productRepo.DeleteProduct(id); err != nil {
		slog.Error("Failed to delete product", "productId", id, "error", err)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
