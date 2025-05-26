package services

import (
	"fmt"
	"log/slog"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

// CategoryService 定义分类相关的业务逻辑接口
type CategoryService interface {
	// 基本查询方法
	ListCategories(params *query_params.QueryParams) ([]models.Category, *response.Pagination, error)
	GetCategoryTree(depth int, enabledOnly bool) ([]models.Category, error)
	GetCategory(id uint) (*models.Category, error)

	// 管理方法 - 可以根据需要后续实现
	// CreateCategory(category *models.Category) (uint, error)
	// UpdateCategory(id uint, updates map[string]interface{}) error
	// DeleteCategory(id uint) error
}

// categoryService 分类服务实现
type categoryService struct {
	categoryRepo repositories.CategoryRepository
}

// NewCategoryService 创建一个分类服务实例
func NewCategoryService(categoryRepo repositories.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

// ListCategories 获取分类列表
func (s *categoryService) ListCategories(params *query_params.QueryParams) ([]models.Category, *response.Pagination, error) {
	// 调用Repo层获取分类列表
	categories, total, err := s.categoryRepo.ListCategories(params)
	if err != nil {
		slog.Error("Failed to list categories", "error", err)
		return nil, nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// 没有数据时返回空数组
	if len(categories) == 0 {
		categories = []models.Category{}
	}

	// 构建分页信息
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return categories, pagination, nil
}

// GetCategoryTree 获取分类树结构
func (s *categoryService) GetCategoryTree(depth int, enabledOnly bool) ([]models.Category, error) {
	// 调用Repo层获取分类树，传入深度参数和是否显示所有分类
	categories, err := s.categoryRepo.GetCategoryTree(depth, enabledOnly)
	if err != nil {
		slog.Error("Failed to get category tree", "error", err, "depth", depth, "enabledOnly", enabledOnly)
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}

	// 没有数据时返回空数组
	if len(categories) == 0 {
		categories = []models.Category{}
	}

	return categories, nil
}

// GetCategory 获取单个分类详情
func (s *categoryService) GetCategory(id uint) (*models.Category, error) {
	// 调用repo层获取分类
	category, err := s.categoryRepo.GetCategory(id)
	if err != nil {
		slog.Error("Failed to get category", "categoryId", id, "error", err)
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}
