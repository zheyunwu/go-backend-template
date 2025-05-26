package repositories

import (
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	// 通用CRUD查询
	ListCategories(params *query_params.QueryParams) ([]models.Category, int, error)
	GetCategory(id uint) (*models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(id uint, updates map[string]interface{}) error
	DeleteCategory(id uint) error

	// 定制查询
	GetCategoryTree(depth int, enabledOnly bool) ([]models.Category, error)
	GetChildCategories(parentID uint) ([]models.Category, error)
	GetCategoryProducts(categoryID uint, params *query_params.QueryParams) ([]models.Product, int, error)

	// 用于其他repository调用的工具方法
	ExpandCategoryIDsWithChildren(categoryIDs []uint) ([]uint, error)
	GetAllChildCategoryIDs(parentID uint) ([]uint, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

/*
通用CRUD查询
*/

// ListCategories 获取分类列表
func (r *categoryRepository) ListCategories(params *query_params.QueryParams) ([]models.Category, int, error) {
	var categories []models.Category
	var totalCount int64

	// 创建查询
	query := r.db.Model(&models.Category{})

	// 处理搜索
	if params.Search != "" {
		query = query.Where("name LIKE ? OR name_zh LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 处理过滤
	if params.Filter != nil {
		for key, value := range params.Filter {
			query = query.Where(key+" = ?", value)
		}
	}

	// 处理排序
	if params.Sort != "" {
		query = query.Order(params.Sort)
	} else {
		query = query.Order("name ASC") // 默认按名称升序
	}

	// 获取总数
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// 执行查询
	err = query.Find(&categories).Error
	return categories, int(totalCount), err
}

// GetCategory 获取单个分类
func (r *categoryRepository) GetCategory(id uint) (*models.Category, error) {
	var category models.Category
	// 预加载子分类（如果需要）
	err := r.db.First(&category, id).Error
	return &category, err
}

// CreateCategory 创建分类
func (r *categoryRepository) CreateCategory(category *models.Category) error {
	return r.db.Create(category).Error
}

// UpdateCategory 更新分类
func (r *categoryRepository) UpdateCategory(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Category{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteCategory 删除分类（软删除）
func (r *categoryRepository) DeleteCategory(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

/*
定制查询
*/

// GetCategoryTree 获取分类树结构
func (r *categoryRepository) GetCategoryTree(depth int, enabledOnly bool) ([]models.Category, error) {
	// 先获取所有根分类（没有父分类的）
	var rootCategories []models.Category
	query := r.db.Where("parent_id IS NULL")

	// 根 enabledOnly参数决定是否只显示启用的分类
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	if err := query.Find(&rootCategories).Error; err != nil {
		return nil, err
	}

	// 递归获取每个根分类的子分类，传入当前深度为1
	for i := range rootCategories {
		if err := r.loadChildCategoriesWithDepth(&rootCategories[i], 1, depth, enabledOnly); err != nil {
			return nil, err
		}
	}

	return rootCategories, nil
}

// 递归加载子分类，带深度限制
func (r *categoryRepository) loadChildCategoriesWithDepth(category *models.Category, currentDepth, maxDepth int, enabledOnly bool) error {
	// 如果已经达到最大深度或者不限制深度(maxDepth <= 0)，则不继续加载子分类
	if maxDepth > 0 && currentDepth >= maxDepth {
		category.Children = []models.Category{} // 设置为空数组而不是nil
		return nil
	}

	var children []models.Category
	query := r.db.Where("parent_id = ?", category.ID)

	// 根据enabledOnly参数决定是否只显示启用的分类
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	if err := query.Find(&children).Error; err != nil {
		return err
	}

	// 继续递归加载每个子分类的子分类
	for i := range children {
		if err := r.loadChildCategoriesWithDepth(&children[i], currentDepth+1, maxDepth, enabledOnly); err != nil {
			return err
		}
	}

	// 设置子分类
	category.Children = children
	return nil
}

// GetChildCategories 获取直接子分类
func (r *categoryRepository) GetChildCategories(parentID uint) ([]models.Category, error) {
	var children []models.Category
	err := r.db.Where("parent_id = ?", parentID).Find(&children).Error
	return children, err
}

// GetCategoryProducts 获取分类下的产品
func (r *categoryRepository) GetCategoryProducts(categoryID uint, params *query_params.QueryParams) ([]models.Product, int, error) {
	var products []models.Product
	var totalCount int64

	// 使用连接表查询特定分类的产品
	query := r.db.Table("products").
		Joins("JOIN product_categories ON products.id = product_categories.product_id").
		Where("product_categories.category_id = ?", categoryID)

	// 处理搜索
	if params.Search != "" {
		query = query.Where("products.name LIKE ? OR products.barcode LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 处理过滤
	if params.Filter != nil {
		for key, value := range params.Filter {
			query = query.Where("products."+key+" = ?", value)
		}
	}

	// 处理排序
	if params.Sort != "" {
		query = query.Order("products." + params.Sort)
	} else {
		query = query.Order("products.updated_at DESC") // 默认按更新时间倒序
	}

	// 获取总数
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// 预加载关联数据
	query = query.Preload("Images").Preload("Retailers").Preload("Categories")

	// 执行查询
	err = query.Find(&products).Error
	return products, int(totalCount), err
}

/*
工具方法
*/

// ExpandCategoryIDsWithChildren 获取所有指定分类的ID及其所有子分类的ID
func (r *categoryRepository) ExpandCategoryIDsWithChildren(categoryIDs []uint) ([]uint, error) {
	// 如果没有分类ID，返回空结果
	if len(categoryIDs) == 0 {
		return []uint{}, nil
	}

	// 创建结果集，首先包含原始分类ID
	expandedIDs := make(map[uint]bool)
	for _, id := range categoryIDs {
		expandedIDs[id] = true
	}

	// 递归查找所有子分类
	for _, id := range categoryIDs {
		childIDs, err := r.GetAllChildCategoryIDs(id)
		if err != nil {
			return nil, err
		}

		// 将子分类ID添加到结果集
		for _, childID := range childIDs {
			expandedIDs[childID] = true
		}
	}

	// 将map转换为slice
	result := make([]uint, 0, len(expandedIDs))
	for id := range expandedIDs {
		result = append(result, id)
	}

	return result, nil
}

// GetAllChildCategoryIDs 递归获取指定分类的所有子分类ID
func (r *categoryRepository) GetAllChildCategoryIDs(parentID uint) ([]uint, error) {
	var childIDs []uint
	var childCategories []models.Category

	// 查询直接子分类
	if err := r.db.Where("parent_id = ?", parentID).Find(&childCategories).Error; err != nil {
		return nil, err
	}

	// 如果没有子分类，返回空结果
	if len(childCategories) == 0 {
		return []uint{}, nil
	}

	// 收集子分类ID
	for _, child := range childCategories {
		childIDs = append(childIDs, child.ID)

		// 递归获取子分类的子分类
		grandChildIDs, err := r.GetAllChildCategoryIDs(child.ID)
		if err != nil {
			return nil, err
		}

		// 添加到结果
		childIDs = append(childIDs, grandChildIDs...)
	}

	return childIDs, nil
}
