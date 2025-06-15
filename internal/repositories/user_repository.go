package repositories

import (
	"context" // Added for context
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/query_params"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	// General CRUD queries
	ListUsers(ctx context.Context, params *query_params.QueryParams, includeSoftDeleted ...bool) ([]models.User, int, error)
	GetUser(ctx context.Context, id uint, includeSoftDeleted ...bool) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, id uint, updates map[string]interface{}, includeSoftDeleted ...bool) error
	DeleteUser(ctx context.Context, id uint) error
	// Custom queries
	GetUserByField(ctx context.Context, field string, value string, includeSoftDeleted ...bool) (*models.User, error)
	CreateUserProvider(ctx context.Context, userProvider *models.UserProvider) error
	GetUserByProvider(ctx context.Context, provider string, providerUID string) (*models.User, error)
	GetUserByUnionID(ctx context.Context, unionID string) (*models.User, error)
	DeleteUserProvider(ctx context.Context, userID uint, provider string) error
	GetUserProvider(ctx context.Context, userID uint, provider string) (*models.UserProvider, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

/*
5 general CRUD queries
*/

// ListUsers retrieves a list of users based on query parameters.
func (r *userRepository) ListUsers(ctx context.Context, params *query_params.QueryParams, includeSoftDeleted ...bool) ([]models.User, int, error) {
	var users []models.User
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.User{}) // Add WithContext

	// If explicitly passed true, include soft-deleted records.
	if len(includeSoftDeleted) > 0 && includeSoftDeleted[0] {
		query = query.Unscoped()
	}

	// Handle search, filter, sort.
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

	// Get total count of records.
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination (note: no need to validate pagination params here as it's done in ParseQueryParams Middleware).
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Preload associated data.
	query = query.Preload("UserProviders") // Preload user's associated third-party login providers.

	// Execute query.
	err = query.Find(&users).Error
	return users, int(totalCount), err
}

// GetUser retrieves a single user by ID, optionally including soft-deleted records.
func (r *userRepository) GetUser(ctx context.Context, id uint, includeSoftDeleted ...bool) (*models.User, error) {
	var user models.User
	query := r.db.WithContext(ctx) // Add WithContext

	// If explicitly passed true, include soft-deleted records.
	if len(includeSoftDeleted) > 0 && includeSoftDeleted[0] {
		query = query.Unscoped()
	}

	err := query.Preload("UserProviders").First(&user, id).Error
	return &user, err
}

// CreateUser creates a new user.
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error // Add WithContext
}

// UpdateUser updates an existing user.
func (r *userRepository) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}, includeSoftDeleted ...bool) error {
	// Partial update.
	query := r.db.WithContext(ctx).Model(&models.User{}) // Add WithContext

	// If explicitly passed true, include soft-deleted records (though typically updates are on existing, non-deleted records).
	if len(includeSoftDeleted) > 0 && includeSoftDeleted[0] {
		query = query.Unscoped()
	}

	return query.Where("id = ?", id).Updates(updates).Error
}

// DeleteUser deletes a user (soft delete if gorm.DeletedAt field exists in the model).
func (r *userRepository) DeleteUser(ctx context.Context, id uint) error {
	// GORM automatically performs a soft delete if the model has a gorm.DeletedAt field.
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error // Add WithContext
}

/*
Custom queries
*/

// GetUserByField retrieves a user by a specific field and value.
func (r *userRepository) GetUserByField(ctx context.Context, field string, value string, includeSoftDeleted ...bool) (*models.User, error) {
	var user models.User
	query := r.db.WithContext(ctx) // Add WithContext

	// If explicitly passed true, include soft-deleted records.
	if len(includeSoftDeleted) > 0 && includeSoftDeleted[0] {
		query = query.Unscoped()
	}

	err := query.Where(field+" = ?", value).First(&user).Error
	return &user, err
}

func (r *userRepository) CreateUserProvider(ctx context.Context, userProvider *models.UserProvider) error {
	return r.db.WithContext(ctx).Create(userProvider).Error // Add WithContext
}

func (r *userRepository) GetUserByProvider(ctx context.Context, provider string, providerUID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Joins("JOIN user_providers ON users.id = user_providers.user_id"). // Add WithContext
		Where("user_providers.provider = ? AND user_providers.provider_uid = ?", provider, providerUID).
		First(&user).Error
	return &user, err
}

func (r *userRepository) GetUserByUnionID(ctx context.Context, unionID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Joins("JOIN user_providers ON users.id = user_providers.user_id"). // Add WithContext
		Where("user_providers.wechat_union_id = ?", unionID).
		First(&user).Error
	return &user, err
}

func (r *userRepository) DeleteUserProvider(ctx context.Context, userID uint, provider string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.UserProvider{}).Error // Add WithContext
}

func (r *userRepository) GetUserProvider(ctx context.Context, userID uint, provider string) (*models.UserProvider, error) {
	var userProvider models.UserProvider
	err := r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).First(&userProvider).Error // Add WithContext
	return &userProvider, err
}
