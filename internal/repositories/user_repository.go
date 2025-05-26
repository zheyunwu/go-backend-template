package repositories

import (
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

type UserRepository interface {
	// 通用CRUD查询
	ListUsers(params *query_params.QueryParams) ([]models.User, int, error)
	GetUser(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(id uint, updates map[string]interface{}) error
	DeleteUser(id uint) error
	// 定制查询
	GetUserByField(field string, value string) (*models.User, error)
	CreateUserProvider(userProvider *models.UserProvider) error
	GetUserByProvider(provider string, providerUID string) (*models.User, error)
	GetUserByUnionID(unionID string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

/*
5个通用CRUD查询
*/

func (r *userRepository) ListUsers(params *query_params.QueryParams) ([]models.User, int, error) {
	var users []models.User
	var totalCount int64

	query := r.db.Model(&models.User{})

	// 处理搜索 search、过滤 filter、排序 sort
	if params.Search != "" {
		query = query.Where("nickname LIKE ?", "%"+params.Search+"%")
	}

	if params.Filter != nil {
		for key, value := range params.Filter {
			query = query.Where(key+" = ?", value)
		}
	}

	if params.Sort != "" {
		query = query.Order(params.Sort)
	}

	// 查询数据总数
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// 应用分页 （注意这里无需验证分页参数，因为已经在ParseQueryParams Middleware中验证过了）
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// 加载关联数据
	query = query.Preload("UserProviders") // 预加载用户关联的第三方登录提供商

	// 查询数据
	err = query.Find(&users).Error
	return users, int(totalCount), err
}

func (r *userRepository) GetUser(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("UserProviders").First(&user, id).Error
	return &user, err
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UpdateUser(id uint, updates map[string]interface{}) error {
	// Partial update
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepository) DeleteUser(id uint) error {
	// 只要Model中有DeletedAt gorm.DeletedAt字段，GORM就会自动软删除
	return r.db.Delete(&models.User{}, id).Error
}

/*
定制查询
*/

func (r *userRepository) GetUserByField(field string, value string) (*models.User, error) {
	var user models.User
	err := r.db.Where(field+" = ?", value).First(&user).Error
	return &user, err
}

func (r *userRepository) CreateUserProvider(userProvider *models.UserProvider) error {
	return r.db.Create(userProvider).Error
}

func (r *userRepository) GetUserByProvider(provider string, providerUID string) (*models.User, error) {
	var user models.User
	err := r.db.Joins("JOIN user_providers ON users.id = user_providers.user_id").
		Where("user_providers.provider = ? AND user_providers.provider_uid = ?", provider, providerUID).
		First(&user).Error
	return &user, err
}

func (r *userRepository) GetUserByUnionID(unionID string) (*models.User, error) {
	var user models.User
	err := r.db.Joins("JOIN user_providers ON users.id = user_providers.user_id").
		Where("user_providers.wechat_union_id = ?", unionID).
		First(&user).Error
	return &user, err
}
