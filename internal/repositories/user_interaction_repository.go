package repositories

import (
	"strconv"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

type UserInteractionRepository interface {
	// 点赞相关
	AddLike(userID, productID uint) error
	RemoveLike(userID, productID uint) error
	IsLiked(userID, productID uint) (bool, error)

	// 收藏相关
	AddFavorite(userID, productID uint) error
	RemoveFavorite(userID, productID uint) error
	IsFavorited(userID, productID uint) (bool, error)

	// 通用列表方法
	ListUserInteractedProducts(userID uint, params *query_params.QueryParams, interactionType string) ([]models.Product, int, error)

	// 统计功能
	GetProductLikeCount(productID uint) (int, error)
	GetProductFavoriteCount(productID uint) (int, error)
}

// userInteractionRepository 用户交互数据访问实现
type userInteractionRepository struct {
	db           *gorm.DB
	categoryRepo CategoryRepository
}

// NewUserInteractionRepository 创建用户交互数据访问实例
func NewUserInteractionRepository(db *gorm.DB, categoryRepo CategoryRepository) UserInteractionRepository {
	return &userInteractionRepository{
		db:           db,
		categoryRepo: categoryRepo,
	}
}

// AddLike 添加点赞
func (r *userInteractionRepository) AddLike(userID, productID uint) error {
	like := models.UserProductLike{
		UserID:    userID,
		ProductID: productID,
	}

	// 使用FirstOrCreate，避免重复点赞
	return r.db.Where(models.UserProductLike{UserID: userID, ProductID: productID}).
		FirstOrCreate(&like).Error
}

// RemoveLike 取消点赞
func (r *userInteractionRepository) RemoveLike(userID, productID uint) error {
	// 使用主键条件删除
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&models.UserProductLike{}).Error
}

// IsLiked 检查是否已点赞
func (r *userInteractionRepository) IsLiked(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserProductLike{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// AddFavorite 添加收藏
func (r *userInteractionRepository) AddFavorite(userID, productID uint) error {
	favorite := models.UserProductFavorite{
		UserID:    userID,
		ProductID: productID,
	}

	// 使用FirstOrCreate，避免重复收藏
	return r.db.Where(models.UserProductFavorite{UserID: userID, ProductID: productID}).
		FirstOrCreate(&favorite).Error
}

// RemoveFavorite 取消收藏
func (r *userInteractionRepository) RemoveFavorite(userID, productID uint) error {
	// 使用主键条件删除
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&models.UserProductFavorite{}).Error
}

// IsFavorited 检查是否已收藏
func (r *userInteractionRepository) IsFavorited(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserProductFavorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// ListUserInteractedProducts 获取用户点赞或收藏的产品
// interactionType: "like" 或 "favorite"
func (r *userInteractionRepository) ListUserInteractedProducts(userID uint, params *query_params.QueryParams, interactionType string) ([]models.Product, int, error) {
	var products []models.Product
	var total int64
	var tableName, orderField string

	// 根据交互类型确定表名和排序字段
	switch interactionType {
	case "like":
		tableName = "user_product_likes"
		orderField = "user_product_likes.created_at"
	case "favorite":
		tableName = "user_product_favorites"
		orderField = "user_product_favorites.created_at"
	default:
		return nil, 0, nil
	}

	// 创建查询
	query := r.db.Table("products").
		Joins("JOIN "+tableName+" ON products.id = "+tableName+".product_id").
		Where(tableName+".user_id = ? AND products.deleted_at IS NULL", userID)

	// 处理搜索 search
	if params.Search != "" {
		query = query.Where("products.name LIKE ? OR products.barcode LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 处理过滤 filter
	if params.Filter != nil {
		for key, value := range params.Filter {
			// 特殊处理: categories过滤
			if key == "categories" {
				if categoryIDs, ok := value.([]interface{}); ok && len(categoryIDs) > 0 {
					// 创建存储数字ID的切片
					var categoryIDsUint []uint

					// 将interface{}转换为uint类型的ID
					for _, id := range categoryIDs {
						switch v := id.(type) {
						case float64:
							// JSON数字默认解析为float64
							categoryIDsUint = append(categoryIDsUint, uint(v))
						case int:
							categoryIDsUint = append(categoryIDsUint, uint(v))
						case uint:
							categoryIDsUint = append(categoryIDsUint, v)
						case string:
							// 如果是字符串形式的数字，尝试转换
							if numID, err := strconv.ParseUint(v, 10, 64); err == nil {
								categoryIDsUint = append(categoryIDsUint, uint(numID))
							}
						}
					}

					// 扩展分类ID列表，包含所有指定分类的子分类
					expandedCategoryIDs, err := r.categoryRepo.ExpandCategoryIDsWithChildren(categoryIDsUint)
					if err != nil {
						return nil, 0, err
					}

					// 使用扩展后的分类ID列表过滤产品
					if len(expandedCategoryIDs) > 0 {
						// 使用EXISTS子查询，比JOIN和IN更高效
						query = query.Where("EXISTS (SELECT 1 FROM product_categories pc WHERE pc.product_id = products.id AND pc.category_id IN ?)", expandedCategoryIDs)
					}
				}
			} else {
				// 常规过滤条件处理
				query = query.Where("products."+key+" = ?", value)
			}
		}
	}

	// 处理排序 sort
	if params.Sort != "" {
		query = query.Order("products." + params.Sort)
	} else {
		query = query.Order(orderField + " DESC") // 默认按交互时间倒序
	}

	// 查询数据总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// 预加载关联数据
	query = query.Preload("Images").Preload("Categories")

	err = query.Find(&products).Error

	return products, int(total), err
}

// GetProductLikeCount 获取产品的点赞数量
func (r *userInteractionRepository) GetProductLikeCount(productID uint) (int, error) {
	var count int64
	err := r.db.Model(&models.UserProductLike{}).
		Where("product_id = ?", productID).
		Count(&count).Error
	return int(count), err
}

// GetProductFavoriteCount 获取产品的收藏数量
func (r *userInteractionRepository) GetProductFavoriteCount(productID uint) (int, error) {
	var count int64
	err := r.db.Model(&models.UserProductFavorite{}).
		Where("product_id = ?", productID).
		Count(&count).Error
	return int(count), err
}
