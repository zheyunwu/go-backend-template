package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/jwt"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 定义用户相关的业务逻辑接口
type UserService interface {
	/* 面向Admin的业务逻辑 */
	ListUsers(params *query_params.QueryParams) ([]models.User, *response.Pagination, error)
	GetUser(id uint) (*models.User, error)
	CreateUser(req *dto.RegisterWithPasswordRequest) (uint, error)
	UpdateUser(id uint, req *dto.UpdateProfileRequest) error
	DeleteUser(id uint) error
	BanUser(id uint, banned bool) error // banned为true时封禁，false时解除封禁

	/* 面向User的业务逻辑 */
	CheckUserExists(fieldType string, value string) (bool, error)
	UpdateProfile(id uint, req *dto.UpdateProfileRequest, authenticatedUser *models.User) error
	/* 传统注册登录相关 */
	RegisterWithPassword(req *dto.RegisterWithPasswordRequest) (uint, error)
	LoginWithPassword(emailOrPhone, password string) (string, error)
	/* 微信小程序端 */
	RegisterFromWechatMiniProgram(req *dto.RegisterFromWechatMiniProgramRequest, unionID *string, openID *string) (uint, error)
	LoginFromWechatMiniProgram(unionID *string, openID *string) (string, error)
}

// userService 用户服务实现
type userService struct {
	config   *config.Config
	userRepo repositories.UserRepository
}

// NewUserService 创建一个用户服务实例
func NewUserService(config *config.Config, userRepo repositories.UserRepository) UserService {
	return &userService{
		config:   config,
		userRepo: userRepo,
	}
}

/*
面向Admin的业务逻辑
*/

// ListUsers 获取用户列表
func (s *userService) ListUsers(params *query_params.QueryParams) ([]models.User, *response.Pagination, error) {
	// 调用Repo层 获取用户列表
	userList, total, err := s.userRepo.ListUsers(params)
	if err != nil {
		slog.Error("Failed to list users", "error", err)
		return nil, nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 没有数据时返回空数组
	if len(userList) == 0 {
		userList = []models.User{}
	}

	// 构建分页信息
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return userList, pagination, nil
}

// GetUser 获取用户详情
func (s *userService) GetUser(id uint) (*models.User, error) {
	// 调用repo层获取用户
	user, err := s.userRepo.GetUser(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		slog.Error("Failed to get user from repository", "userId", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser 创建用户（传统Email/Phone + 密码）
func (s *userService) CreateUser(req *dto.RegisterWithPasswordRequest) (uint, error) {
	// 验证至少提供email或phone之一
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return 0, errors.ErrEmailOrPhoneNotProvided
	}

	// 验证是否有重复的邮箱或手机号
	if req.Email != nil && *req.Email != "" {
		if _, err := s.userRepo.GetUserByField("email", *req.Email); err == nil {
			return 0, errors.ErrEmailAlreadyExists
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if _, err := s.userRepo.GetUserByField("phone", *req.Phone); err == nil {
			return 0, errors.ErrPhoneAlreadyExists
		}
	}

	// 对密码进行哈希处理
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return 0, fmt.Errorf("error processing password: %w", err)
	}
	// 使用DTO转换为User模型
	user := req.ToModel(hashedPassword)

	// 调用repo层创建用户
	if err := s.userRepo.CreateUser(user); err != nil {
		slog.Error("Failed to create user",
			"name", user.Name,
			"email", req.Email,
			"phone", req.Phone,
			"error", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

// hashPassword 使用bcrypt哈希密码
func hashPassword(password string) (string, error) {
	// 使用推荐的cost值(10-12)生成哈希密码
	// bcrypt自动添加随机盐值并将其包含在哈希结果中
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(id uint, req *dto.UpdateProfileRequest) error {
	// 检查用户是否存在
	_, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// 调用DTO的ToMap方法构建更新字段映射
	updates := req.ToUpdatesMap()

	// 调用repo层进行更新
	if err := s.userRepo.UpdateUser(id, updates); err != nil {
		slog.Error("Failed to update user", "userId", id, "error", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	// 检查用户是否存在
	_, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// 调用repo层删除用户
	if err := s.userRepo.DeleteUser(id); err != nil {
		slog.Error("Failed to delete user", "userId", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// BanUser 封禁或解除封禁用户
func (s *userService) BanUser(id uint, isBanned bool) error {
	// 检查用户是否存在
	_, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// 构建更新字段映射
	updates := map[string]interface{}{
		"is_banned": isBanned,
	}
	if err := s.userRepo.UpdateUser(id, updates); err != nil {
		slog.Error("Failed to update ban status", "userId", id, "error", err)
		return fmt.Errorf("failed to update ban status: %w", err)
	}
	return nil
}

/*
面向User的业务逻辑
*/

// CheckUserExists 检查用户是否存在
func (s *userService) CheckUserExists(fieldType string, value string) (bool, error) {
	var err error

	switch fieldType {
	case "mini_program_open_id":
		// 通过UserProvider表查询微信小程序openID
		_, err = s.userRepo.GetUserByProvider("wechat_mini_program", value)
	case "union_id":
		// 通过UserProvider表查询微信UnionID
		_, err = s.userRepo.GetUserByUnionID(value)
	default:
		// 其他字段直接在User表中查询
		_, err = s.userRepo.GetUserByField(fieldType, value)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateProfile 更新用户信息
func (s *userService) UpdateProfile(id uint, req *dto.UpdateProfileRequest, authenticatedUser *models.User) error {
	// 查询目标资源，验证目标资源是否属于请求者userID
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}

	if user.ID != authenticatedUser.ID {
		slog.Warn("Permission denied for user update", "userId", id, "requesterId", authenticatedUser.ID)
		return errors.ErrPermissionDenied
	}

	// 将DTO转换为更新字段Map
	updates := req.ToUpdatesMap()

	// 调用repo层进行更新
	if err := s.userRepo.UpdateUser(id, updates); err != nil {
		slog.Error("Failed to update user", "userId", id, "error", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

/*
传统注册登录相关
*/

// RegisterWithPassword 使用密码注册用户
func (s *userService) RegisterWithPassword(req *dto.RegisterWithPasswordRequest) (uint, error) {
	// 转交给CreateUser处理
	return s.CreateUser(req)
}

// LoginWithPassword 验证用户密码并生成JWT token
func (s *userService) LoginWithPassword(emailOrPhone, password string) (string, error) {
	var user *models.User
	var err error

	// 先尝试用邮箱查找用户
	user, err = s.userRepo.GetUserByField("email", emailOrPhone)
	if err != nil && err == gorm.ErrRecordNotFound {
		// 如果邮箱找不到，尝试用手机号查找
		user, err = s.userRepo.GetUserByField("phone", emailOrPhone)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return "", errors.ErrUserNotFound
			}
			slog.Error("Failed to find user", "emailOrPhone", emailOrPhone, "error", err)
			return "", fmt.Errorf("database error: %w", err)
		}
	} else if err != nil {
		slog.Error("Failed to find user", "emailOrPhone", emailOrPhone, "error", err)
		return "", fmt.Errorf("database error: %w", err)
	}

	// 检查用户是否被封禁
	if user.IsBanned {
		return "", errors.ErrUserBanned
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		slog.Warn("Password verification failed", "userId", user.ID)
		return "", errors.ErrInvalidPassword
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		slog.Error("Failed to generate JWT", "error", err, "userId", user.ID)
		return "", fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.userRepo.UpdateUser(user.ID, map[string]interface{}{"last_login": now})

	return token, nil
}

/*
微信小程序端
*/

// RegisterUserFromWechatMiniProgram 从微信小程序注册用户
func (s *userService) RegisterFromWechatMiniProgram(req *dto.RegisterFromWechatMiniProgramRequest, unionID *string, openID *string) (uint, error) {
	// 在生产环境下检查敏感内容
	// if openID != nil && *openID != "" {
	// 	// 检查用户昵称是否包含敏感内容
	// 	if err := utils.CheckSensitiveContent(req.Name, *openID, utils.SecuritySceneProfile); err != nil {
	// 		return 0, err
	// 	}
	// }

	// 验证openID是否提供
	if openID == nil || *openID == "" {
		return 0, errors.ErrUserNotFound // 或者定义一个新的错误类型
	}

	// 验证：用户是否已在微信小程序端注册过
	if _, err := s.userRepo.GetUserByProvider("wechat_mini_program", *openID); err == nil {
		return 0, errors.ErrUserAlreadyExists
	}

	// 若有UnionID，看一下用户是否已在APP端注册过，如有，则直接取其关联的UserID
	if unionID != nil && *unionID != "" {
		existingUser, err := s.userRepo.GetUserByUnionID(*unionID)
		if err == nil && existingUser != nil && existingUser.ID > 0 {
			// 用户已在APP端注册过，用这个用户的ID
			userProvider := models.UserProvider{
				UserID:        existingUser.ID,
				Provider:      "wechat_mini_program",
				ProviderUID:   *openID,
				WechatUnionID: unionID,
			}
			if err := s.userRepo.CreateUserProvider(&userProvider); err != nil {
				slog.Error("Failed to create user provider",
					"userId", existingUser.ID,
					"provider", "wechat_mini_program",
					"providerUID", *openID,
					"unionID", unionID,
					"error", err)
				return 0, fmt.Errorf("failed to create user provider: %w", err)
			}
			slog.Info("UserProvider created successfully",
				"userId", existingUser.ID,
				"provider", "wechat_mini_program",
				"providerUID", *openID,
				"unionID", unionID)
			return existingUser.ID, nil
		} else if err != gorm.ErrRecordNotFound {
			slog.Error("Failed to find user by unionID", "unionID", *unionID, "error", err)
			return 0, fmt.Errorf("database error: %w", err)
		}
	}

	// 验证是否有重复的邮箱或手机号
	if req.Email != nil && *req.Email != "" {
		if _, err := s.userRepo.GetUserByField("email", *req.Email); err == nil {
			return 0, errors.ErrEmailAlreadyExists
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if _, err := s.userRepo.GetUserByField("phone", *req.Phone); err == nil {
			return 0, errors.ErrPhoneAlreadyExists
		}
	}

	// DTO转换为User模型
	user := req.ToModel()

	// 调用repo层创建用户
	err := s.userRepo.CreateUser(user)
	if err != nil {
		slog.Error("Failed to create user",
			"nickname", user.Name,
			"email", req.Email,
			"phone", req.Phone,
			"openId", openID,
			"error", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// 创建UserProvider记录
	userProvider := models.UserProvider{
		UserID:      user.ID,
		Provider:    "wechat_mini_program",
		ProviderUID: *openID,
	}
	// 如果有unionID，设置它
	if unionID != nil && *unionID != "" {
		userProvider.WechatUnionID = unionID
	}

	// 调用repo层创建UserProvider
	err = s.userRepo.CreateUserProvider(&userProvider)
	if err != nil {
		slog.Error("Failed to create user provider",
			"userId", user.ID,
			"provider", "wechat_mini_program",
			"providerUID", *openID,
			"unionID", unionID,
			"error", err)
		return 0, fmt.Errorf("failed to create user provider: %w", err)
	}

	slog.Info("User and UserProvider created successfully",
		"userId", user.ID,
		"provider", "wechat_mini_program",
		"providerUID", *openID,
		"unionID", unionID)

	return user.ID, nil
}

// LoginFromWechatMiniProgram 微信小程序登录
func (s *userService) LoginFromWechatMiniProgram(unionID *string, openID *string) (string, error) {
	// 如果unionID和openID都为nil，则直接返回错误
	if unionID == nil && openID == nil {
		return "", errors.ErrUserNotFound
	}

	var user *models.User
	var err error

	// 优先级策略：unionID > openID
	// 先尝试用unionID查找用户（如果提供了的话）
	if unionID != nil && *unionID != "" {
		user, err = s.userRepo.GetUserByUnionID(*unionID)
		if err != nil && err != gorm.ErrRecordNotFound {
			slog.Error("Failed to find user by unionID", "unionID", *unionID, "error", err)
			return "", fmt.Errorf("database error: %w", err)
		}
	}

	// 如果通过unionID没有找到用户，再尝试用openID查找
	if user == nil && openID != nil && *openID != "" {
		user, err = s.userRepo.GetUserByProvider("wechat_mini_program", *openID)
		if err != nil && err != gorm.ErrRecordNotFound {
			slog.Error("Failed to find user by provider", "provider", "wechat_mini_program", "providerUID", *openID, "error", err)
			return "", fmt.Errorf("database error: %w", err)
		}
	}

	// 如果还是没有找到用户，返回用户未找到错误
	if user == nil || user.ID == 0 {
		return "", errors.ErrUserNotFound
	}

	// 检查用户是否被封禁
	if user.IsBanned {
		return "", errors.ErrUserBanned
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		slog.Error("Failed to generate JWT", "error", err, "userId", user.ID)
		return "", fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.userRepo.UpdateUser(user.ID, map[string]interface{}{"last_login": now})

	slog.Info("WeChat mini program login successful",
		"userId", user.ID,
		"unionID", unionID,
		"openID", openID)

	return token, nil
}
