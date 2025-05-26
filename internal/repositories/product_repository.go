package repositories

import (
	"strconv"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

type ProductRepository interface {
	// 通用CRUD查询
	ListProducts(params *query_params.QueryParams) ([]models.Product, int, error)
	GetProduct(id uint) (*models.Product, error)
	CreateProduct(product *models.Product) error
	UpdateProduct(id uint, updates map[string]interface{}) error
	DeleteProduct(id uint) error

	// 事务支持
	CreateProductWithRelations(product *models.Product, images []models.ProductImage, categoryIDs []uint) error
	UpdateProductWithRelations(id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error
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
5个通用CRUD查询
*/

func (r *productRepository) ListProducts(params *query_params.QueryParams) ([]models.Product, int, error) {
	var products []models.Product
	var totalCount int64

	// 创建查询
	query := r.db.Model(&models.Product{})

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
		query = query.Order("products.updated_at DESC") // 默认按更新时间倒序
	}

	// 查询数据总数
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// 预加载关联数据
	query = query.Preload("Images").Preload("Categories")

	// 查询数据
	err = query.Find(&products).Error
	return products, int(totalCount), err
}

func (r *productRepository) GetProduct(id uint) (*models.Product, error) {
	var product models.Product
	// 预加载关联数据
	err := r.db.Preload("Images").Preload("Categories").First(&product, id).Error
	return &product, err
}

func (r *productRepository) CreateProduct(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) UpdateProduct(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error
}

func (r *productRepository) DeleteProduct(id uint) error {
	// 只要Model中有DeletedAt gorm.DeletedAt字段，GORM就会自动软删除
	return r.db.Delete(&models.Product{}, id).Error
}

/*
事务支持
*/

// CreateProductWithRelations 在同一事务中创建产品及其所有关联数据
func (r *productRepository) CreateProductWithRelations(product *models.Product, images []models.ProductImage, categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建产品基本信息
		if err := tx.Create(product).Error; err != nil {
			return err
		}

		// 2. 添加产品图片
		for i := range images {
			images[i].ProductID = product.ID // 设置产品ID
			if err := tx.Create(&images[i]).Error; err != nil {
				return err
			}
		}

		// 3. 添加产品分类关联
		for _, categoryID := range categoryIDs {
			productCategory := models.ProductCategory{
				ProductID:  product.ID,
				CategoryID: categoryID,
			}
			if err := tx.Create(&productCategory).Error; err != nil {
				return err
			}
		}

		// 所有操作成功
		return nil
	})
}

// UpdateProductWithRelations 在同一事务中智能更新产品及其所有关联数据
func (r *productRepository) UpdateProductWithRelations(id uint, updates map[string]interface{}, newImages []models.ProductImage, categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新产品基本信息
		if len(updates) > 0 {
			if err := tx.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}

		// 2. 智能更新图片 - 仅当提供了新图片列表时
		if newImages != nil {
			// 2.1 获取当前所有图片
			var existingImages []models.ProductImage
			if err := tx.Where("product_id = ?", id).Find(&existingImages).Error; err != nil {
				return err
			}

			// 2.2 创建URL映射用于快速查找
			existingImageMap := make(map[string]uint) // URL -> ID
			for _, img := range existingImages {
				existingImageMap[img.ImageURL] = img.ID
			}

			newImageMap := make(map[string]bool) // 记录新图片URL
			for _, img := range newImages {
				newImageMap[img.ImageURL] = true
			}

			// 2.3 删除不再需要的图片 - 使用Unscoped()强制永久删除
			for url, imgID := range existingImageMap {
				if !newImageMap[url] {
					if err := tx.Unscoped().Delete(&models.ProductImage{}, imgID).Error; err != nil {
						return err
					}
				}
			}

			// 2.4 只添加新的图片
			for _, img := range newImages {
				if _, exists := existingImageMap[img.ImageURL]; !exists {
					img.ProductID = id
					if err := tx.Create(&img).Error; err != nil {
						return err
					}
				}
			}
		}

		// 3. 更新分类关联 - 仅当提供了新分类列表时
		if categoryIDs != nil {
			// 4.1 获取当前分类关联
			var existingCategories []models.ProductCategory
			if err := tx.Where("product_id = ?", id).Find(&existingCategories).Error; err != nil {
				return err
			}

			// 3.2 创建映射用于快速查找
			existingCategoryMap := make(map[uint]bool)
			for _, pc := range existingCategories {
				existingCategoryMap[pc.CategoryID] = true
			}

			newCategoryMap := make(map[uint]bool)
			for _, catID := range categoryIDs {
				newCategoryMap[catID] = true
			}

			// 3.3 删除不再需要的分类关联 - 这里使用Unscoped()确保物理删除
			for _, pc := range existingCategories {
				if !newCategoryMap[pc.CategoryID] {
					if err := tx.Unscoped().Where("product_id = ? AND category_id = ?",
						id, pc.CategoryID).Delete(&models.ProductCategory{}).Error; err != nil {
						return err
					}
				}
			}

			// 3.4 只添加新的分类关联
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

		// 所有操作成功
		return nil
	})
}
